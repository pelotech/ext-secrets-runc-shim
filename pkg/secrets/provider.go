package secrets

import (
	"context"
	"fmt"

	ocispec "github.com/opencontainers/runtime-spec/specs-go"
)

// Provider is the exposed interface for retrieving secrets from various backends
// to populate environment variables with.
type Provider interface {
	// Setup should prepare the provider for operations. It is given the spec
	// of the container to detect additional configurations it needs. When possible,
	// a provider should try to act as a "singleton" and cache objects globally
	// after this call, such that new provider instances can reuse them. Each process
	// operates in the context of a single container, so this is generally safe.
	Setup(ctx context.Context, spec *ocispec.Spec) error

	// GetValue should lookup the secret at path and return the value or any error.
	GetValue(ctx context.Context, path string) (value string, err error)
}

func GetProviderByName(name string) (Provider, error) {
	switch name {
	case "vault": // vault
		return &vaultProvider{}, nil
	case "ssm": // aws ssm
		return &awsSSMProvider{}, nil
	case "gsm": // google secret manager
		return &gcpSecretsProvider{}, nil
	case "akv": // azure key vault
		return &azureSecretsProvider{}, nil
	default:
		return nil, fmt.Errorf("no secret provider for '%s'", name)
	}
}
