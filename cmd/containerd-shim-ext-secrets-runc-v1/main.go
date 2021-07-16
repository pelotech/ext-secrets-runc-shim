package main

import (
	"github.com/containerd/containerd/runtime/v2/shim"

	extSecretsShim "github.com/pelotech/ext-secrets-runc-shim/pkg/shim"
)

func main() {
	// init and execute the shim
	shim.Run("io.containerd.test-shim.v1", extSecretsShim.New)
}
