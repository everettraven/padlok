# Configuring OIDC Issuers

The `issuer` block within each JWT authenticator configures how `padlok` connects to an OIDC provider and validates tokens. For background on how Kubernetes handles OIDC token verification — including signing key discovery, JWKS endpoints, and token validation — see the [Kubernetes OIDC documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#openid-connect-tokens).

## Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `url` | string | Yes | The issuer URL. Must match the `iss` claim in presented JWTs. Discovery information is fetched from `{url}/.well-known/openid-configuration` unless `discoveryURL` is set. Must be unique across all JWT authenticators. |
| `discoveryURL` | string | No | Overrides the OIDC discovery endpoint. The `issuer` field in the fetched discovery document must still match `url`. Useful when the discovery/JWKS endpoints are hosted at a different location than the issuer (e.g., locally in the cluster). Must be different from `url`. |
| `audiences` | list of strings | Yes | Accepted audience values. At least one entry must match the `aud` claim in presented JWTs. Must be non-empty. |
| `audienceMatchPolicy` | string | No | How `audiences` is matched against the `aud` claim. Allowed values: `"MatchAny"` (at least one entry must match) or empty/unset (same behavior when a single audience is specified). For exact audience matching, use a `claimValidationRule` with a CEL expression instead. |
| `certificateAuthority` | string | No | PEM-encoded CA certificates used to validate the TLS connection when fetching discovery information from the issuer. If unset, the system certificate store is used. |

## Examples

### Basic Issuer

```yaml
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

### Issuer with Custom CA

For issuers using self-signed or internal CA certificates:

```yaml
jwt:
  - issuer:
      url: https://idp.internal.example.com
      audiences:
        - my-k8s-cluster
      certificateAuthority: |
        -----BEGIN CERTIFICATE-----
        <PEM-encoded CA certificate>
        -----END CERTIFICATE-----
    claimMappings:
      username:
        claim: "sub"
        prefix: ""
```

### Overriding the Discovery URL

When the OIDC discovery and JWKS endpoints are served from a different location than the token issuer (e.g., a cluster-local mirror):

```yaml
jwt:
  - issuer:
      url: https://idp.example.com
      discoveryURL: https://oidc.oidc-namespace/.well-known/openid-configuration
      audiences:
        - my-k8s-cluster
      certificateAuthority: |
        -----BEGIN CERTIFICATE-----
        <PEM-encoded CA certificate for oidc.oidc-namespace>
        -----END CERTIFICATE-----
    claimMappings:
      username:
        claim: "sub"
        prefix: ""
```

The `certificateAuthority` is used to verify the TLS connection to the discovery URL. The leaf certificate must have a hostname matching the discovery URL's host (`oidc.oidc-namespace` in this example).

### Multiple Audiences

To accept tokens issued to multiple client IDs:

```yaml
jwt:
  - issuer:
      url: https://idp.example.com
      audiences:
        - web-client
        - cli-client
      audienceMatchPolicy: MatchAny
    claimMappings:
      username:
        claim: "sub"
        prefix: ""
```

For strict audience matching (require an exact set), use a claim validation rule:

```yaml
jwt:
  - issuer:
      url: https://idp.example.com
      audiences:
        - web-client
    claimValidationRules:
      - expression: 'claims.aud == ["web-client", "api-client"]'
        message: "token must be issued to exactly web-client and api-client"
    claimMappings:
      username:
        claim: "sub"
        prefix: ""
```

### Multiple Issuers

Each JWT authenticator entry must have a unique issuer URL. Configure multiple issuers to accept tokens from different providers:

```yaml
jwt:
  - issuer:
      url: https://corporate-idp.example.com
      audiences:
        - k8s-cluster
    claimMappings:
      username:
        claim: "email"
        prefix: "corp:"
  - issuer:
      url: https://partner-idp.example.com
      audiences:
        - k8s-cluster
    claimMappings:
      username:
        claim: "sub"
        prefix: "partner:"
```

Authenticators are tried in order. Since each has a unique issuer URL, at most one will attempt to cryptographically validate any given token.
