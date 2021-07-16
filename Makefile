build:
	cd cmd/containerd-shim-ext-secrets-runc-v1 && \
		CGO_ENABLED=0 go build -o ../../test/shim/containerd-shim-ext-secrets-runc-v1 .

KUBECTL  ?= kubectl
HELM     ?= helm
K8S_VER  ?= v1.21.1
BIN_DIR  ?= $(CURDIR)/bin

K3D         ?= bin/k3d
K3D_VERSION ?= v4.4.7
$(K3D):
	mkdir -p "$(BIN_DIR)"
	curl -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | \
		USE_SUDO=false TAG=$(K3D_VERSION) K3D_INSTALL_DIR="$(BIN_DIR)" bash

CONTEXT          ?= k3d-$(CLUSTER_NAME)
CLUSTER_NAME     ?= ext-secrets
K3D_CLUSTER_ARGS ?= 

k3d-up: $(K3D) build
	$(K3D) cluster create $(K3D_CLUSTER_ARGS) \
		--volume $(CURDIR)/test/shim:/usr/local/bin@server[0] \
		--volume $(CURDIR)/test/containerd:/var/lib/rancher/k3s/agent/etc/containerd@server[0] \
		--port 8200:8200@loadbalancer \
		$(CLUSTER_NAME)
	$(MAKE) install-vault

k3d-down: $(K3D)
	$(K3D) cluster delete $(CLUSTER_NAME)

VAULT_IMAGE = vault:1.7.3
VAULT_CHART = https://helm.releases.hashicorp.com/vault-0.13.0.tgz

install-vault: $(K3D)
	# Pre load vault image for faster startup
	docker image inspect $(VAULT_IMAGE) > /dev/null || docker pull $(VAULT_IMAGE)
	$(K3D) image import --cluster $(CLUSTER_NAME) vault:1.7.3

	$(HELM) upgrade --install \
		--kube-context $(CONTEXT) \
		-f test/vault/values.yaml \
		vault $(VAULT_CHART)

	@echo
	@echo -n "++ Waiting for vault to startup"
	@while ! kubectl get pod vault-0 2> /dev/null | grep Running 1> /dev/null ; do echo -n '.' && sleep 3 ; done
	@echo
	@echo "++ Vault has started"
	$(MAKE) init-vault

TEST_PASSWORD = supersecret
VAULT_EXEC = $(KUBECTL) --context $(CONTEXT) exec -it vault-0 -- vault
init-vault:
	$(VAULT_EXEC) operator init -key-shares=1 -key-threshold=1 -format=json > test/cluster-keys.json
	$(VAULT_EXEC) operator unseal `cat test/cluster-keys.json | jq -r ".unseal_keys_b64[]"`
	$(VAULT_EXEC) login `cat test/cluster-keys.json | jq -r ".root_token"`
	$(VAULT_EXEC) secrets enable -path=secrets kv-v2
	$(VAULT_EXEC) kv put secrets/my-secret password=$(TEST_PASSWORD)
	$(VAULT_EXEC) auth enable kubernetes
	$(VAULT_EXEC) write auth/kubernetes/config \
        kubernetes_host="https://kubernetes.default.svc.cluster.local" \
		issuer="https://kubernetes.default.svc.cluster.local"
	echo 'path "secrets/*" { capabilities = ["read"] }' | \
		$(VAULT_EXEC) policy write vault-shim-test -
	$(VAULT_EXEC) write auth/kubernetes/role/vault-shim \
		bound_service_account_names=default \
		bound_service_account_namespaces=default \
		policies=vault-shim-test \
		ttl=24h

apply-pod:
	$(KUBECTL) apply --context $(CONTEXT) -f test/manifests/pod.yaml

testacc:
	$(MAKE) k3d-up apply-pod
	$(KUBECTL) --context $(CONTEXT) wait pod \
		--for condition=Ready \
		test-pod
	@if $(KUBECTL) logs test-pod | grep $(TEST_PASSWORD) > /dev/null ; then \
		echo "Pod received correct secret"; \
	else \
		echo "Pod did not receive correct secret" && exit 1; \
	fi;
	$(MAKE) k3d-down


DIST ?= $(CURDIR)/dist

GOBIN ?= $(shell go env GOPATH)/bin
GOX   ?= $(GOBIN)/gox
$(GOX):
	GO111MODULE=off go get github.com/mitchellh/gox

LDFLAGS         ?= "-s -w"
COMPILE_TARGETS ?= "linux/amd64 linux/arm linux/arm64"
COMPILE_OUTPUT  ?= "$(DIST)/{{.Dir}}_{{.OS}}_{{.Arch}}"
.PHONY: dist
dist: $(GOX)
	mkdir -p $(DIST)
	cd cmd/containerd-shim-ext-secrets-runc-v1 && \
		CGO_ENABLED=0 $(GOX) -osarch=$(COMPILE_TARGETS) -output=$(COMPILE_OUTPUT) -ldflags=$(LDFLAGS)
	upx -9 $(DIST)/*