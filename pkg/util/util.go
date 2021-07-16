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

package util

import (
	"strings"

	ocispec "github.com/opencontainers/runtime-spec/specs-go"
)

const extSecretPrefix = "ext-secret"
const extSecretAnnotationPrefix = "ext-secrets.runc.io"

func EnvStrToKeyValue(in string) (key, value string) {
	spl := strings.Split(in, "=")
	if len(spl) < 2 {
		key = spl[0]
		return
	}
	key, value = spl[0], strings.Join(spl[1:], "=")
	return
}

func GetAnnotation(spec *ocispec.Spec, lookup string) string {
	for key, val := range spec.Annotations {
		key = strings.TrimPrefix(key, extSecretAnnotationPrefix+"/")
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
