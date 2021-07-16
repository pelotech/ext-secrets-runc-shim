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
	"fmt"

	ocispec "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go/aws"
)

type awsSSMProvider struct{}

var ssmClient *ssm.Client

func (a *awsSSMProvider) Setup(ctx context.Context, _ *ocispec.Spec) error {
	if ssmClient != nil {
		return nil
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	ssmClient = ssm.NewFromConfig(cfg)
	return nil
}

func (a *awsSSMProvider) GetValue(ctx context.Context, path string) (value string, err error) {
	res, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(path),
		WithDecryption: true,
	})
	if err != nil {
		return
	}
	if res.Parameter == nil || res.Parameter.Value == nil {
		err = fmt.Errorf("value of ssm parameter '%s' is empty", path)
		return
	}
	value = *res.Parameter.Value
	return
}
