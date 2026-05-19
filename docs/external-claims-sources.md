# External Claims Sources

External claims sources are `padlok`'s primary extension to the Kubernetes Structured Authentication Configuration. They allow fetching additional claims from HTTP endpoints during authentication, before claim-to-identity mapping occurs. There is no upstream Kubernetes equivalent for this feature.

## Overview

Each external claims source defines:
- **Where** to fetch claims from (URL).
- **How** to authenticate to the external source (authentication method).
- **What** to extract from the response (mappings).
- **When** to consult the source (conditions).

External claims are merged into the token's claim set before claim mappings run. This means claim mapping expressions can reference externally sourced claims the same way they reference token claims — via the `claims` variable.

## Limits

- A maximum of **5** external claims sources can be configured per JWT authenticator.
- Each source can have up to **16** mappings and **16** conditions.
- Each external source request has a **500ms timeout**.
- Mapping names must be unique across all external claims sources within a JWT authenticator.

## Graceful Degradation

If an external source is unavailable or returns an error, authentication is **not** rejected. The externally sourced claims simply won't be present in the claim set, and authentication continues in a degraded state. Errors are logged. This ensures that authentication is not entirely dependent on external source availability.

## Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `authentication` | object | No | How to authenticate to the external source. Defaults to anonymous when omitted. |
| `tls` | object | No | TLS settings for the HTTP connection to the external source. |
| `url` | object | Yes | The URL to fetch claims from, composed of a hostname and a CEL path expression. |
| `mappings` | list | Yes | CEL expressions that extract claims from the response. Must have 1–16 entries. |
| `conditions` | list | No | CEL expressions that determine whether to consult this source. Max 16 entries. |

## Authentication

The `authentication` block configures how `padlok` authenticates to the external source.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes (when `authentication` is set) | The authentication method. Must be `"RequestProvidedToken"` or `"ClientCredential"`. |
| `clientCredential` | object | Conditional | Required when `type` is `"ClientCredential"`. Must not be set otherwise. |

### RequestProvidedToken

Reuses the bearer token from the original request to authenticate with the external source. This is useful when the token has the necessary scopes/audiences to access the external endpoint (e.g., an OIDC UserInfo endpoint).

```yaml
externalClaimsSources:
  - authentication:
      type: RequestProvidedToken
    url: { ... }
    mappings: [ ... ]
```

### ClientCredential

Uses the [OAuth2 client credentials grant](https://datatracker.ietf.org/doc/html/rfc6749#section-4.4) to obtain a separate access token for the external source. This is useful when the external source requires its own credentials (e.g., Microsoft Graph API).

```yaml
externalClaimsSources:
  - authentication:
      type: ClientCredential
      clientCredential:
        clientID: "my-client-id"
        clientSecret: "my-client-secret"
        tokenEndpoint: "https://login.example.com/oauth2/token"
        scopes:
          - "https://graph.example.com/.default"
        tls:
          certificateAuthority: |
            -----BEGIN CERTIFICATE-----
            ...
            -----END CERTIFICATE-----
    url: { ... }
    mappings: [ ... ]
```

#### ClientCredential Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `clientID` | string | Yes | OAuth2 client identifier. Must contain only printable ASCII characters. |
| `clientSecret` | string | Yes | OAuth2 client secret. Must contain only printable ASCII characters. |
| `tokenEndpoint` | string | Yes | HTTPS URL to obtain an access token from. Must have a host and path, and must not contain query parameters, fragments, or user information. |
| `scopes` | list of strings | No | OAuth2 scopes to request. If not set, the token endpoint's default scopes are used. |
| `tls` | object | No | TLS settings for the connection to the token endpoint. |

### Anonymous (default)

When the `authentication` block is omitted entirely, requests to the external source are made without any authentication credentials.

## URL

The `url` block defines the endpoint to query.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `hostname` | string | Yes | A valid RFC1123 DNS subdomain name, optionally with a port (e.g., `idp.example.com` or `idp.example.com:8443`). Must not exceed 253 characters. |
| `pathExpression` | string | Yes | A CEL expression that returns a list of strings used to construct the URL path. Token claims are available via `claims`. |

The final URL is constructed as `https://{hostname}/{path}`, where `{path}` is built by joining and URL-encoding the list elements returned by the path expression.

```yaml
url:
  hostname: idp.example.com
  pathExpression: "['realms', 'k8s', 'protocol', 'openid-connect', 'userinfo']"
```

This produces: `https://idp.example.com/realms/k8s/protocol/openid-connect/userinfo`

The path expression has access to `claims`, so paths can be dynamic:

```yaml
url:
  hostname: api.example.com
  pathExpression: "['users'] + [claims.sub] + ['profile']"
```

If `claims.sub` is `"jane.doe"`, this produces: `https://api.example.com/users/jane.doe/profile`

## Mappings

Each mapping entry extracts a claim from the external source's JSON response body.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | The claim name that will be available in `claims` during claim mapping. Must consist of only lowercase alpha characters and underscores. Max 256 characters. Must be unique across all external sources. |
| `expression` | string | Yes | A CEL expression that extracts the claim value from the response. The full JSON response body is available via `response.body`. Must produce a string value. |

```yaml
mappings:
  - name: groups
    expression: "response.body.groups.join(',')"
  - name: department
    expression: "response.body.department"
```

**Warning:** Externally sourced claims override token claims with the same name during the claim mapping process. Use conditions to guard against unintentional overrides.

## Conditions

Conditions are optional CEL expressions that determine whether the external source should be consulted. All conditions must evaluate to `true` for the source to be queried. Token claims are available via the `claims` variable.

```yaml
conditions:
  - expression: "!has(claims.groups)"
```

This example only consults the external source when the token does not already contain a `groups` claim.

## TLS

The `tls` block configures the HTTP client's TLS settings for the connection to the external source.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `certificateAuthority` | string | No | PEM-encoded CA certificates to validate the external source's TLS certificate. If not set, the system certificate store is used. |

## Full Example

Fetch group membership from an OIDC UserInfo endpoint when the token doesn't already contain groups:

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
      groups:
        expression: "claims.groups.split(',')"
    externalClaimsSources:
      - authentication:
          type: RequestProvidedToken
        tls:
          certificateAuthority: |
            -----BEGIN CERTIFICATE-----
            ...
            -----END CERTIFICATE-----
        url:
          hostname: idp.example.com
          pathExpression: "['realms', 'k8s', 'protocol', 'openid-connect', 'userinfo']"
        mappings:
          - name: groups
            expression: "response.body.groups.join(',')"
        conditions:
          - expression: "!has(claims.groups)"
```
