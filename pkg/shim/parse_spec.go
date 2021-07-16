package shim

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/log"
	ocispec "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/pelotech/ext-secrets-runc-shim/pkg/secrets"
	"github.com/pelotech/ext-secrets-runc-shim/pkg/util"
)

func parseSpec(ctx context.Context, spec *ocispec.Spec) error {
	envSecrets := make(map[int]map[string]string)
	newEnv := make([]string, len(spec.Process.Env))

	for i, envVar := range spec.Process.Env {
		key, value := util.EnvStrToKeyValue(envVar)
		if !util.IsExtSecret(value) {
			newEnv[i] = envVar
			continue
		}
		envSecrets[i] = map[string]string{key: value}
	}

	if len(envSecrets) == 0 {
		return nil
	}

	for idx, data := range envSecrets {
		for key, fullPath := range data {
			providerName, path := util.ParseEnvValueToSecret(fullPath)
			if providerName == "" {
				log.L.Debugf("Could not detect provider from path: %s", fullPath)
				newEnv[idx] = fmt.Sprintf("%s=", key)
				continue
			}
			provider, err := secrets.GetProviderByName(providerName)
			if err != nil {
				log.L.Debugf("Could not detect provider from path: %s", fullPath)
				newEnv[idx] = fmt.Sprintf("%s=", key)
				continue
			}
			if err := provider.Setup(ctx, spec); err != nil {
				log.L.Debug(err.Error())
				newEnv[idx] = fmt.Sprintf("%s=", key)
				continue
			}
			value, err := provider.GetValue(ctx, path)
			if err != nil {
				log.L.Debugf("Could not detect provider from path: %s", fullPath)
				newEnv[idx] = fmt.Sprintf("%s=", key)
				continue
			}
			newEnv[idx] = fmt.Sprintf("%s=%s", key, value)
		}
	}

	spec.Process.Env = newEnv

	return nil
}
