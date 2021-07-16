/*
Copyright 2021 Pelotech.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package secrets

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd/log"
	ocispec "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/hashicorp/vault/api"

	"github.com/pelotech/ext-secrets-runc-shim/pkg/util"
)

var vaultClient *api.Client

type vaultProvider struct{}

func (v *vaultProvider) Setup(_ context.Context, spec *ocispec.Spec) error {
	if vaultClient != nil {
		return nil
	}
	var err error
	vaultClient, err = getVaultClient(spec)
	return err
}

func (v *vaultProvider) GetValue(_ context.Context, path string) (string, error) {
	return getVaultSecret(vaultClient, path)
}

const defaultTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount"

func getVaultClient(spec *ocispec.Spec) (*api.Client, error) {
	vaultAddr := util.GetAnnotation(spec, "vault-addr")
	if vaultAddr == "" {
		return nil, errors.New("could not detect vault address from container spec")
	}
	authRole := util.GetAnnotation(spec, "vault-auth-role")
	if authRole == "" {
		authRole = "default"
	}
	cfg := api.DefaultConfig()
	cfg.Address = vaultAddr
	var tokenPath string
	for _, mount := range spec.Mounts {
		if mount.Destination == defaultTokenPath {
			tokenPath = filepath.Join(mount.Source, "token")
		}
	}
	if tokenPath == "" {
		return nil, errors.New("could not find a kubernetes service account token for the container")
	}
	token, err := getK8sAuth(tokenPath, authRole, cfg)
	if err != nil {
		return nil, err
	}
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	client.SetToken(token.Auth.ClientToken)
	return client, nil
}

func getVaultSecret(client *api.Client, pathAndKey string) (value string, err error) {
	spl := strings.Split(pathAndKey, "#")
	if len(spl) < 2 {
		err = fmt.Errorf("invalid vault path specification: %s", pathAndKey)
		return
	}

	path, dataKey := strings.TrimPrefix(spl[0], "/"), strings.Join(spl[1:], "#")

	mountPath, v2, err := isKVv2(path, client)
	if err != nil {
		err = fmt.Errorf("preflight check for %s failed: %s", path, err.Error())
		return
	}

	if v2 {
		path = addPrefixToVKVPath(path, mountPath, "data")
		log.L.Debugf("Detected KVv2 mount path, adjusting read path to %s", path)
	}

	log.L.Debugf("Fetching value from vault at %s", path)
	resp, err := client.Logical().Read(path)
	if err != nil {
		err = fmt.Errorf("error reading %s from vault: %s", path, err.Error())
		return
	}
	if resp == nil || resp.Data == nil {
		err = fmt.Errorf("no secret data found at %s", path)
		return
	}
	var data map[string]interface{}
	if v2 {
		data = resp.Data["data"].(map[string]interface{})
	} else {
		data = resp.Data
	}
	val, ok := data[dataKey]
	if !ok {
		err = fmt.Errorf("no %s key found at %s", dataKey, path)
		return
	}

	value = fmt.Sprintf("%v", val)
	return
}

func getK8sAuth(tokenPath, authRole string, vaultConfig *api.Config) (*api.Secret, error) {
	tokenBytes, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return nil, err
	}
	authURLStr := fmt.Sprintf("%s/v1/auth/kubernetes/login", vaultConfig.Address)
	// TODO: role should be configurable
	body := []byte(fmt.Sprintf(`{
		"jwt": "%s",
		"role": "%s"
	}`, string(tokenBytes), authRole))
	req, err := http.NewRequest(http.MethodPost, authURLStr, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	res, err := vaultConfig.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(string(resBody))
	}
	authResponse := &api.Secret{}
	return authResponse, json.Unmarshal(resBody, authResponse)
}

// copied from https://github.com/hashicorp/vault/blob/main/command/kv_helpers.go

func kvPreflightVersionRequest(client *api.Client, path string) (string, int, error) {
	// We don't want to use a wrapping call here so save any custom value and
	// restore after
	currentWrappingLookupFunc := client.CurrentWrappingLookupFunc()
	client.SetWrappingLookupFunc(nil)
	defer client.SetWrappingLookupFunc(currentWrappingLookupFunc)
	currentOutputCurlString := client.OutputCurlString()
	client.SetOutputCurlString(false)
	defer client.SetOutputCurlString(currentOutputCurlString)

	r := client.NewRequest("GET", "/v1/sys/internal/ui/mounts/"+path)
	resp, err := client.RawRequest(r)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		// If we get a 404 we are using an older version of vault, default to
		// version 1
		if resp != nil && resp.StatusCode == 404 {
			return "", 1, nil
		}

		return "", 0, err
	}

	secret, err := api.ParseSecret(resp.Body)
	if err != nil {
		return "", 0, err
	}
	if secret == nil {
		return "", 0, errors.New("nil response from pre-flight request")
	}
	var mountPath string
	if mountPathRaw, ok := secret.Data["path"]; ok {
		mountPath = mountPathRaw.(string)
	}
	options := secret.Data["options"]
	if options == nil {
		return mountPath, 1, nil
	}
	versionRaw := options.(map[string]interface{})["version"]
	if versionRaw == nil {
		return mountPath, 1, nil
	}
	version := versionRaw.(string)
	switch version {
	case "", "1":
		return mountPath, 1, nil
	case "2":
		return mountPath, 2, nil
	}

	return mountPath, 1, nil
}

func isKVv2(path string, client *api.Client) (string, bool, error) {
	mountPath, version, err := kvPreflightVersionRequest(client, path)
	if err != nil {
		return "", false, err
	}

	return mountPath, version == 2, nil
}

func addPrefixToVKVPath(p, mountPath, apiPrefix string) string {
	switch {
	case p == mountPath, p == strings.TrimSuffix(mountPath, "/"):
		return path.Join(mountPath, apiPrefix)
	default:
		p = strings.TrimPrefix(p, mountPath)
		return path.Join(mountPath, apiPrefix, p)
	}
}
