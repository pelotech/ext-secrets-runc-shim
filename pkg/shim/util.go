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
