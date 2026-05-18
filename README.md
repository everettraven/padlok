# padlok

`padlok` is a [Kubernetes webhook token authentication](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication) service that builds on the foundation laid by the [Structured Authentication Configuration](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#api-server-authn-config) feature, extending it for enterprise environments.

## What does padlok add?

- **Claims Resolution** — Resolve claims from sources beyond what is present in the token itself, enabling richer identity mapping configurations for enterprise environments.

## How It Works

`padlok` implements the [webhook token authentication](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication) protocol, making it a drop-in replacement for Structured Authentication Configuration.
No changes are needed on the Kubernetes side beyond pointing to `padlok` as the webhook authenticator.

The following highlights how a request to the Kubernetes API server goes through authentication when `padlok` is configured:
1. The Kubernetes API server receives a request with a bearer token and sends a `TokenReview` to `padlok`'s `/authenticate` endpoint.
2. `padlok` validates the JWT against the configured OIDC issuer.
3. If configured, additional claims are resolved from external sources (e.g., a UserInfo endpoint or Microsoft Graph API).
4. Token and external claims are mapped to a Kubernetes identity (username, groups, uid, extra) using [CEL expressions](https://kubernetes.io/docs/reference/using-api/cel/).
5. The authenticated identity is returned to the API server.

## Configuration

`padlok` is configured with an `AuthenticationConfiguration` resource that defines how tokens should be validated and mapped to Kubernetes identities. The configuration file is specified at startup via the `--config` flag and is automatically reloaded when changes are detected.

`padlok`'s `AuthenticationConfiguration` type mirrors the [Kubernetes `AuthenticationConfiguration`](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#api-server-authn-config) type, extending it with additional fields like `externalClaimsSources`. If you already have a Structured Authentication Configuration, you can use it as a starting point — update the `apiVersion` to `padlok.everettraven.github.io/v1alpha1` and add any extensions you need.

A minimal configuration that authenticates tokens from an OIDC provider:

```yaml
apiVersion: padlok.everettraven.github.io/v1alpha1
kind: AuthenticationConfiguration
jwt:
  - issuer:
      url: https://idp.example.com
      audiences:
        - my-k8s-cluster
    claimMappings:
      username:
        claim: "sub"
        prefix: ""
```

To resolve claims from an external source, add an `externalClaimsSources` entry. The following example fetches group membership from the OIDC provider's UserInfo endpoint when the token does not already contain a `groups` claim:

```yaml
apiVersion: padlok.everettraven.github.io/v1alpha1
kind: AuthenticationConfiguration
jwt:
  - issuer:
      url: https://idp.example.com
      audiences:
        - my-k8s-cluster
    claimMappings:
      username:
        claim: "sub"
        prefix: ""
      groups:
        expression: "claims.groups.split(',')"
    externalClaimsSources:
      - authentication:
          type: RequestProvidedToken
        url:
          hostname: idp.example.com
          pathExpression: "['protocol', 'openid-connect', 'userinfo']"
        mappings:
          - name: groups
            expression: "response.body.groups.join(',')"
        conditions:
          - expression: "!has(claims.groups)"
```

## Getting Started

Detailed installation and usage documentation is on its way. In the meantime, check back on this repository or watch it for updates.

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, conventions, and contribution workflow.

## License

`padlok` is licensed under the [Apache License 2.0](LICENSE).

