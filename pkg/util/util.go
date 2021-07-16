package util

import (
	"strings"

	ocispec "github.com/opencontainers/runtime-spec/specs-go"
)

const extSecretPrefix = "ext-secret"

func EnvStrToKeyValue(in string) (key, value string) {
	spl := strings.Split(in, "=")
	if len(spl) < 2 {
		key = spl[0]
		return
	}
	key, value = spl[0], strings.Join(spl[1:], "=")
	return
}

func GetEnvValueByKey(spec *ocispec.Spec, lookup string) string {
	for _, env := range spec.Process.Env {
		key, val := EnvStrToKeyValue(env)
		if key == lookup {
			return val
		}
	}
	return ""
}

func IsExtSecret(in string) bool { return strings.HasPrefix(in, extSecretPrefix+":") }

func ParseEnvValueToSecret(in string) (provider, path string) {
	in = strings.TrimPrefix(in, extSecretPrefix+":")
	spl := strings.Split(in, ":")
	if len(spl) < 2 {
		path = in
		return
	}
	provider, path = spl[0], strings.Join(spl[1:], ":")
	return
}
