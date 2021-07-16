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
