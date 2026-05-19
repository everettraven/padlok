# Concepts

This page covers the key terms and concepts used throughout `padlok`. For general Kubernetes authentication concepts, see the [Kubernetes authentication documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/).

## Webhook Token Authentication

Kubernetes supports delegating token authentication to an external service via the [webhook token authentication](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication) protocol. When configured, the API server sends a `TokenReview` request containing the bearer token to the webhook service, which responds with the authenticated identity or a rejection.

`padlok` implements this protocol, acting as the webhook authenticator that the API server delegates to.

## OIDC and JWT

[OpenID Connect (OIDC)](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#openid-connect-tokens) is an identity layer built on OAuth 2.0. Identity providers (IdPs) issue JSON Web Tokens (JWTs) containing **claims** — key-value pairs that describe the token holder (e.g., `sub`, `email`, `groups`).

`padlok` validates JWTs by discovering the issuer's public signing keys through the standard OIDC discovery mechanism (`/.well-known/openid-configuration` and the JWKS endpoint).

## Claims

Claims are the key-value pairs inside a JWT. Standard claims include `iss` (issuer), `sub` (subject), `aud` (audience), and `exp` (expiration). Identity providers may also include custom claims like `groups`, `email`, or `department`.

`padlok` maps claims to Kubernetes identity attributes using [claim mappings](claim-mappings.md).

## External Claims Sources

External claims sources are `padlok`'s primary extension to the Kubernetes Structured Authentication Configuration. They allow fetching additional claims from HTTP endpoints — such as OIDC UserInfo endpoints or Microsoft Graph API — that are not present in the token itself.

Each external source is configured with:
- An **authentication method** for accessing the source (reuse the request token, use OAuth2 client credentials, or anonymous).
- A **URL** built from a hostname and a CEL path expression.
- **Mappings** that extract claims from the HTTP response using CEL expressions.
- Optional **conditions** that control when the source should be consulted.

Externally sourced claims are merged into the claim set before identity mapping occurs. See [External Claims Sources](external-claims-sources.md) for full details.

## CEL Expressions

[Common Expression Language (CEL)](https://kubernetes.io/docs/reference/using-api/cel/) is used throughout `padlok` for flexible claim mapping, validation, and external source configuration. CEL expressions can access token claims via the `claims` variable, the mapped user identity via the `user` variable, and external source responses via the `response` variable.

See [CEL Expressions](cel-expressions.md) for the expression environment and common patterns.

## Claim Mappings

Claim mappings define how token claims (and any externally sourced claims) are transformed into a Kubernetes identity. The four identity attributes are:

- **username** (required) — The Kubernetes username.
- **groups** (optional) — The Kubernetes groups.
- **uid** (optional) — The Kubernetes user ID.
- **extra** (optional) — Arbitrary key-value metadata attached to the identity.

Each attribute can be set via a direct claim reference or a CEL expression. See [Claim Mappings](claim-mappings.md).

## Validation Rules

Validation rules are CEL expressions that must evaluate to `true` for authentication to succeed:

- **Claim validation rules** — Validate token claims before identity mapping (e.g., require `email_verified == true`).
- **User validation rules** — Validate the final mapped identity (e.g., reject usernames starting with `system:`).

See [Validation Rules](validation-rules.md).

## Configuration Hot-Reload

`padlok` watches its configuration file for changes and automatically reloads it when modifications are detected. The new configuration is validated before it takes effect — if validation fails, the previous configuration remains active and an error is logged.
