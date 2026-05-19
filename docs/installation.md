# Installation

This page covers how to build, deploy, and connect `padlok` to a Kubernetes cluster.

## Prerequisites

- A Kubernetes cluster with access to configure the API server's authentication flags.
- An OIDC identity provider (e.g., Keycloak, Azure AD/Entra ID, Okta, Dex).
- TLS certificates for `padlok`'s HTTPS server.
- A container runtime (Docker or Podman) for building and running the `padlok` image.

## Building from Source

Clone the repository and build the binary:

```bash
git clone https://github.com/everettraven/padlok.git
cd padlok
make build
```

This produces a `bin/padlok` binary.

## Building the Container Image

```bash
make image
```

By default, the image is tagged as `quay.io/rh_ee_bpalmer/padlok:latest`. Override the tag with:

```bash
make image IMAGE_TAG="my-registry/padlok:v0.1.0"
```

## Deploying padlok

`padlok` runs as a standalone HTTPS server. It needs to be network-accessible to the Kubernetes API server.

### Running the Server

```bash
padlok run \
  --config /path/to/config.yaml \
  --tls-cert-file /path/to/tls.crt \
  --tls-private-key-file /path/to/tls.key \
  --secure-port 6443
```

See [CLI Reference](cli-reference.md) for the full list of flags.

### Running as a Container

```bash
docker run -d \
  -v /path/to/config.yaml:/cfg/config.yaml \
  -v /path/to/tls.crt:/certs/tls.crt \
  -v /path/to/tls.key:/certs/tls.key \
  my-registry/padlok:v0.1.0 run \
    --config=/cfg/config.yaml \
    --tls-cert-file=/certs/tls.crt \
    --tls-private-key-file=/certs/tls.key
```

## Configuring the Kubernetes API Server

The API server needs two things:

1. A **webhook token authentication config file** — a kubeconfig-style file that tells the API server where `padlok` is and how to connect to it.
2. The `--authentication-token-webhook-config-file` flag pointing to that file.

### Webhook Config File

Create a kubeconfig-style file that points to `padlok`:

```yaml
apiVersion: v1
kind: Config

clusters:
- cluster:
    server: https://<padlok-host>:6443/authenticate
    certificate-authority: /path/to/padlok-ca.pem
  name: padlok

users:
- name: apiserver

current-context: webhook
contexts:
- context:
    cluster: padlok
    user: apiserver
  name: webhook
```

- `server` must point to `padlok`'s `/authenticate` endpoint.
- `certificate-authority` should contain the CA certificate that signed `padlok`'s TLS certificate, so the API server trusts the connection.

### API Server Flags

Configure the API server with:

```
--authentication-token-webhook-config-file=/path/to/webhook-config.yaml
--authentication-token-webhook-version=v1
```

See the [Kubernetes webhook token authentication documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication) for more details on the API server configuration.

## Creating the padlok Configuration

Create an `AuthenticationConfiguration` file for `padlok`. See [AuthenticationConfiguration](authentication-configuration.md) for the full reference, or start with the [Quickstart](quickstart.md) for a minimal example.

## Next Steps

- [Quickstart](quickstart.md) — A minimal end-to-end example.
- [AuthenticationConfiguration](authentication-configuration.md) — Configuration reference.
- [TLS Configuration](tls-configuration.md) — TLS setup details.
