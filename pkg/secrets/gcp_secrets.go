package secrets

import (
	"context"

	ocispec "github.com/opencontainers/runtime-spec/specs-go"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type gcpSecretsProvider struct{}

var gcpClient *secretmanager.Client

func (g *gcpSecretsProvider) Setup(ctx context.Context, _ *ocispec.Spec) error {
	if gcpClient != nil {
		return nil
	}
	var err error
	gcpClient, err = secretmanager.NewClient(ctx)
	if err != nil {
		gcpClient = nil
	}
	return err
}

func (g *gcpSecretsProvider) GetValue(ctx context.Context, path string) (value string, err error) {
	res, err := gcpClient.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: path,
	})
	if err != nil {
		return
	}
	value = res.Payload.String()
	return
}
