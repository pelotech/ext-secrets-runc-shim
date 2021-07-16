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
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault"
	ocispec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pelotech/ext-secrets-runc-shim/pkg/util"
)

type azureSecretsProvider struct{}

var (
	keyVaultBaseURL string
	keyVaultClient  *keyvault.BaseClient
)

func (a *azureSecretsProvider) Setup(_ context.Context, spec *ocispec.Spec) error {
	if keyVaultClient != nil {
		return nil
	}
	keyVaultBaseURL = util.GetAnnotation(spec, "keyvault-base-url")
	if keyVaultBaseURL == "" {
		return errors.New("could not detect base keyvault url from container spec")
	}
	client := keyvault.New()
	keyVaultClient = &client
	return nil
}

func (a *azureSecretsProvider) GetValue(ctx context.Context, path string) (value string, err error) {
	res, err := keyVaultClient.GetSecret(ctx, keyVaultBaseURL, path, "")
	if err != nil {
		return
	}
	if res.Value == nil {
		err = fmt.Errorf("no secret data found at %s", path)
		return
	}
	value = *res.Value
	return
}
