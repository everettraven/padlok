# Configuration Schema

This page documents the full schema for the `AuthenticationConfiguration` resource (`padlok.everettraven.github.io/v1alpha1`). Fields inherited from the upstream [Kubernetes `AuthenticationConfiguration`](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#using-authentication-configuration) are marked as such. Fields added by `padlok` are marked as **padlok extension**.

## AuthenticationConfiguration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | Yes | Must be `padlok.everettraven.github.io/v1alpha1`. |
| `kind` | string | Yes | Must be `AuthenticationConfiguration`. |
| `jwt` | list of [JWTAuthenticator](#jwtauthenticator) | Yes | JWT authenticator configurations. Must contain at least one entry. |

## JWTAuthenticator

| Field | Type | Required | Source | Description |
|-------|------|----------|--------|-------------|
| `issuer` | [Issuer](#issuer) | Yes | Upstream | OIDC provider connection settings. |
| `claimMappings` | [ClaimMappings](#claimmappings) | Yes | Upstream | Maps token claims to Kubernetes identity attributes. |
| `claimValidationRules` | list of [ClaimValidationRule](#claimvalidationrule) | No | Upstream | CEL rules to validate token claims. |
| `userValidationRules` | list of [UserValidationRule](#uservalidationrule) | No | Upstream | CEL rules to validate the final mapped identity. |
| `externalClaimsSources` | list of [ExternalClaimsSource](#externalclaimssource) | No | **padlok extension** | External HTTP sources to fetch additional claims. Max 5 entries. |

## Issuer

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `url` | string | Yes | Issuer URL (must match `iss` claim). Discovery is fetched from `{url}/.well-known/openid-configuration` unless `discoveryURL` is set. Must be unique across all JWT authenticators. |
| `discoveryURL` | string | No | Overrides the OIDC discovery endpoint. The `issuer` field in the discovery document must still match `url`. Must be different from `url`. Must be unique across all JWT authenticators. |
| `audiences` | list of strings | Yes | Accepted audience values. Must be non-empty. |
| `audienceMatchPolicy` | string | No | `"MatchAny"` or empty. Required to be `"MatchAny"` when multiple audiences are specified. |
| `certificateAuthority` | string | No | PEM-encoded CA certificates for the issuer's TLS connection. |

## ClaimMappings

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `username` | [PrefixedClaimOrExpression](#prefixedclaimorexpression) | Yes | Kubernetes username. Must produce a string. |
| `groups` | [PrefixedClaimOrExpression](#prefixedclaimorexpression) | No | Kubernetes groups. Must produce a string or string array. |
| `uid` | [ClaimOrExpression](#claimorexpression) | No | Kubernetes user ID. Must produce a string. |
| `extra` | list of [ExtraMapping](#extramapping) | No | Extra key-value attributes. |

## PrefixedClaimOrExpression

`claim`/`prefix` and `expression` are mutually exclusive.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `claim` | string | Conditional | JWT claim name. Mutually exclusive with `expression`. |
| `prefix` | string | Conditional | Prefix prepended to the claim value. Required when `claim` is set (use `""` for no prefix). |
| `expression` | string | Conditional | CEL expression producing the attribute value. Has access to `claims`. |

## ClaimOrExpression

`claim` and `expression` are mutually exclusive. One must be set.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `claim` | string | Conditional | JWT claim name. |
| `expression` | string | Conditional | CEL expression producing the attribute value. Has access to `claims`. |

## ExtraMapping

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `key` | string | Yes | Extra attribute key. Must be a domain-prefix path (e.g., `example.org/foo`), lowercase, and unique. |
| `valueExpression` | string | Yes | CEL expression producing a string or string array. Empty values mean the mapping is not present. |

## ClaimValidationRule

`claim`/`requiredValue` and `expression`/`message` are mutually exclusive.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `claim` | string | Conditional | Required claim name. Mutually exclusive with `expression`. |
| `requiredValue` | string | No | Required claim value. Only used with `claim`. |
| `expression` | string | Conditional | CEL expression that must return `true`. Has access to `claims`. |
| `message` | string | No | Error message when expression returns `false`. |

## UserValidationRule

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `expression` | string | Yes | CEL expression that must return `true`. Has access to `user` (UserInfo object). |
| `message` | string | No | Error message when expression returns `false`. |

## ExternalClaimsSource

**padlok extension** — no upstream equivalent.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `authentication` | [Authentication](#authentication) | No | How to authenticate to the external source. Defaults to anonymous. |
| `tls` | [TLS](#tls) | No | TLS settings for the connection to the external source. |
| `url` | [SourceURL](#sourceurl) | Yes | The URL to fetch claims from. |
| `mappings` | list of [SourcedClaimMapping](#sourcedclaimmapping) | Yes | Claim extraction expressions. 1–16 entries. Names must be unique across all external sources. |
| `conditions` | list of [ExternalSourceCondition](#externalsourcecondition) | No | Conditions for consulting this source. Max 16 entries. Expressions must be unique. |

## Authentication

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | `"RequestProvidedToken"` or `"ClientCredential"`. |
| `clientCredential` | [ClientCredentialConfig](#clientcredentialconfig) | Conditional | Required when `type` is `"ClientCredential"`. Must not be set otherwise. |

## ClientCredentialConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `clientID` | string | Yes | OAuth2 client identifier. Printable ASCII only. |
| `clientSecret` | string | Yes | OAuth2 client secret. Printable ASCII only. |
| `tokenEndpoint` | string | Yes | HTTPS URL for obtaining access tokens. Must have a host and path. Must not contain query parameters, fragments, or user information. |
| `scopes` | list of strings | No | OAuth2 scopes to request. Each scope must be non-empty and contain only printable ASCII (excluding spaces, double quotes, and backslashes). |
| `tls` | [TLS](#tls) | No | TLS settings for the token endpoint connection. |

## TLS

At least one field must be set when this object is specified.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `certificateAuthority` | string | No | PEM-encoded CA certificates. Must be non-empty and valid PEM when set. |

## SourceURL

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `hostname` | string | Yes | RFC1123 DNS subdomain, optionally with a port. Max 253 characters. |
| `pathExpression` | string | Yes | CEL expression returning a list of strings for the URL path. Has access to `claims`. |

## SourcedClaimMapping

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Claim name to produce. Lowercase alpha and underscores only. Max 256 characters. Must be unique across all external sources. |
| `expression` | string | Yes | CEL expression extracting the claim from the response. Has access to `response.body`. Must produce a string. |

## ExternalSourceCondition

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `expression` | string | Yes | CEL expression that must return `true` for the source to be consulted. Has access to `claims`. Must be unique within the source's conditions list. |
