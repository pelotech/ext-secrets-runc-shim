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
	keyVaultBaseURL = util.GetEnvValueByKey(spec, "KEYVAULT_BASE_URL")
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
