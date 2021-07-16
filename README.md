# ext-secrets-runc-shim

A containerd, runc-based, shim for replacing environment variables with secrets from arbitrary external engines.

## Quickstart

### Installation

There is likely a better way to do this, but in the meantime, the quickest way to set a node up with this shim
is to replace the `runtime_type` of the default `runc` shim.

First, go to the [releases page](https://github.com/pelotech/ext-secrets-runc-shim/releases) and download the binary for your system architecture. Once it is downloaded, place it in the default `PATH` on your Kubernetes node(s). It is important that you name the binary `containerd-shim-ext-secrets-runc-v1`. You _may_ replace the `ext-secrets-runc` part depending on the `runtime_type` you specify below.

_Alternative to downloading, clone this repository and run `make`. The output will be in `test/shim`_

_While this project is very early-stages POC, an obvious more persistent and scalable installation would be to bake the binary and following configurations into your node image(s) or bootstrap._

Edit `/etc/containerd/config.toml` and replace the contents of the following section as so:

```toml
[plugins.cri.containerd.runtimes.runc]
  runtime_type = "io.containerd.ext-secrets-runc.v1"
```

And that's it! All pods on this node should now run via the shim. No Webhooks, no Custom Resources, no CLI commands.

### Usage

Usage will vary depending on the secret provider. But the commonality amongst all of them is how they are invoked.
Simply, replace the `value` key in your environment variable configurations with something like the following:

```yaml
      env:
      - name: PASSWORD
        value: ext-secret:ssm:secrets/my-secret-password
``` 

Where the breakdown of the "path" expressed in `value` is: `ext-secret:<provider>:<secret_path>`.

Caveats apply depending on the secret provider used. See below for more details on what each provider assumes/requires.

## Secret Providers

Below is a table of the secret providers implemented and/or tested. 
Since this project is stil POC, tested in this case implies a basic functionality test has been done.

| Provider              | Tag     | Implemented        | Tested             |
|:---------------------:|:-------:|:------------------:|:------------------:|
| Vault                 | `vault` | :heavy_check_mark: | :heavy_check_mark: |
| AWS SSM               | `ssm`   | :heavy_check_mark: | :x: |
| Google Secret Manager | `gsm`   | :heavy_check_mark: | :x: |
| Azure Key Vault       | `akv`   | :heavy_check_mark: | :x: |

_Feel free to open a PR to track the implementation of other secret storage engines._

## Caveats

### Vault

Kubernetes service account authentication is used to retrieve a vault token. 
The service account of the pod being created is used. Additionally, you have to specify
at a minimum the `VAULT_ADDR` (and any other needed configurations) via the containers environment.
This address **must** resolve from outside the Kubernetes network.

See the simple [test pod](test/manifests/pod.yaml) for an example.

### SSM

The default credential chain on the node running the pod is used when retrieving the secret value.

### Google Secret Manager

The default credential chain on the node running the pod is used when retrieving the secret value.


### Azure Key Vault

The default credential chain on the node running the pod is used when retrieving the secret value.
You must also set the `KEYVAULT_BASE_URL` in your pod's environment (similarly as done with Vault above).