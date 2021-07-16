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

package shim

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	ocispec "github.com/opencontainers/runtime-spec/specs-go"
)

func (s *service) getCfgFile() string { return filepath.Join(s.curdir, "config.json") }

func (s *service) getSpec(ctx context.Context) (*ocispec.Spec, error) {
	body, err := ioutil.ReadFile(s.getCfgFile())
	if err != nil {
		return nil, err
	}
	out := &ocispec.Spec{}
	return out, json.Unmarshal(body, out)
}

func (s *service) writeSpec(ctx context.Context, spec *ocispec.Spec) error {
	out, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(s.getCfgFile(), out, 0644)
}
