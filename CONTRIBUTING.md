# Contributing to padlok

Thank you for your interest in contributing to padlok! This guide covers the development setup, workflow, and conventions used in this project.

## Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) or [Podman](https://podman.io/getting-started/installation)
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [openssl](https://www.openssl.org/) (for generating development TLS certificates)

By default the Makefile uses Podman as the container runtime. To use Docker instead, set `CONTAINER_RUNTIME`:

```sh
export CONTAINER_RUNTIME=docker
```

## Development Environment

The project includes a local development environment that runs a kind Kubernetes cluster with Keycloak as the OIDC provider and padlok as the webhook authenticator.

### Hosts file setup

The kind control plane, padlok, and Keycloak all run on the same container network (`kind`). Keycloak issues tokens with an issuer tied to its container hostname, so the issuer URL in every token is `https://keycloak:8443/realms/k8s`. For token validation to succeed, all parties — including your browser during the Device Code flow — must reach Keycloak at that exact hostname.

Add the following entry to your hosts file (`/etc/hosts` on Linux/macOS, `C:\Windows\System32\drivers\etc\hosts` on Windows):

```
127.0.0.1 keycloak
```

Without this, the issuer in the token will not match the URL your browser uses, and authentication will fail.

### Bring up the full environment

```sh
make up
```

This generates TLS certificates, starts Keycloak, builds and runs the padlok webhook, creates a kind cluster configured to use padlok for authentication, and retrieves a test token. The final step uses the OAuth 2.0 Device Code flow — you will need to open the URL printed in your terminal and authenticate in a browser to complete the token retrieval.

### Tear it down

```sh
make down
```

### Individual targets

| Target | Description |
|---|---|
| `make build` | Build the `padlok` binary to `bin/padlok` |
| `make image` | Build the container image |
| `make unit` | Run unit tests |
| `make keycloak` | Start the Keycloak OIDC provider |
| `make keycloak-certificate` | Generate self-signed TLS certificates for Keycloak |
| `make webhook` | Run the padlok webhook container |
| `make webhook-certificate` | Generate self-signed TLS certificates for the webhook |
| `make cluster` | Create (or recreate) the kind cluster |
| `make generate-config` | Generate the padlok configuration from the template |
| `make token` | Obtain a test token from Keycloak using the OAuth 2.0 Device Code flow (interactive — requires you to open a URL and authenticate in a browser) |

## Making Changes

1. Fork the repository and create a branch from `main`.
2. Make your changes.
3. Add or update tests as needed.
4. Run the tests:
   ```sh
   make unit
   ```
5. Ensure the code compiles:
   ```sh
   make build
   ```
6. Commit your changes following the [commit message conventions](#commit-messages) below.
7. Open a pull request against `main`.

## Project Layout

```
main.go              Entrypoint
pkg/
  cmd/               CLI commands (Cobra)
  apis/              AuthenticationConfiguration types and validation
  handlers/          HTTP handlers (/authenticate endpoint)
  authenticator/jwt/ JWT authentication logic
  oidc/              OIDC integration
  server/            HTTP server setup
  internal/          Internal utilities and upstream forks
dev/
  kind/              kind cluster configuration
  keycloak/          Keycloak realm import data
  certs/             Generated TLS certificates (git-ignored)
  cfg/               Generated configuration files (git-ignored)
```

## Commit Messages

This project uses semantic commit prefixes. Format your commit messages as:

```
<type>: <short summary>
```

Common types:

| Type | When to use |
|---|---|
| `feat` | New feature or capability |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `dev` | Development tooling, build, or environment changes |
| `test` | Adding or updating tests |
| `refactor` | Code change that neither fixes a bug nor adds a feature |

## Code Conventions

- Follow standard Go conventions and idioms.
- Use `go fmt` to format your code.
- Keep packages focused — the project mirrors Kubernetes API patterns in `pkg/apis/` and `pkg/authenticator/`.
- Tests live alongside the code they test (`_test.go` files in the same package).

## License

By contributing to padlok, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE).
