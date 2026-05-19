# AuthenticationConfiguration

`padlok` is configured with an `AuthenticationConfiguration` resource that defines how tokens should be validated and mapped to Kubernetes identities.

## File Format

The configuration file is a YAML document with the following top-level structure:

```yaml
apiVersion: padlok.everettraven.github.io/v1alpha1
kind: AuthenticationConfiguration
jwt:
  - issuer: { ... }
    claimMappings: { ... }
    claimValidationRules: [ ... ]
    userValidationRules: [ ... ]
    externalClaimsSources: [ ... ]
```

## Relationship to Upstream Kubernetes

`padlok`'s `AuthenticationConfiguration` mirrors the [Kubernetes `AuthenticationConfiguration`](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#using-authentication-configuration) type. All fields from the upstream type are supported with the same behavior. `padlok` extends the type with additional fields — most notably `externalClaimsSources`.

If you already have a Structured Authentication Configuration for the Kubernetes API server, you can use it as a starting point. Update the `apiVersion` to `padlok.everettraven.github.io/v1alpha1` and add any `padlok`-specific fields you need.

## Top-Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | Yes | Must be `padlok.everettraven.github.io/v1alpha1`. |
| `kind` | string | Yes | Must be `AuthenticationConfiguration`. |
| `jwt` | list | Yes | A list of JWT authenticator configurations. Must contain at least one entry. |

## JWT Authenticator

Each entry in the `jwt` list configures an authenticator for a single OIDC issuer. Multiple entries allow authenticating tokens from different issuers. Authenticators are tried in order; since each must have a unique issuer URL, at most one will attempt to validate any given token.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `issuer` | object | Yes | OIDC provider connection settings. See [Configuring OIDC Issuers](configuring-oidc-issuers.md). |
| `claimMappings` | object | Yes | Maps token claims to Kubernetes identity attributes. See [Claim Mappings](claim-mappings.md). |
| `claimValidationRules` | list | No | CEL rules to validate token claims. See [Validation Rules](validation-rules.md). |
| `userValidationRules` | list | No | CEL rules to validate the final mapped identity. See [Validation Rules](validation-rules.md). |
| `externalClaimsSources` | list | No | External HTTP sources to fetch additional claims. Max 5 entries. See [External Claims Sources](external-claims-sources.md). |

## Configuration File Path

The configuration file is specified at startup with the `--config` flag:

```bash
padlok run --config /path/to/config.yaml
```

## Hot-Reload

`padlok` watches the configuration file for changes and automatically reloads it when modifications are detected. The file contents are hashed to avoid unnecessary reloads when the content has not actually changed.

When a change is detected:
1. The new configuration is read and validated.
2. If validation succeeds, a new token authenticator is built and swapped in. The previous authenticator is shut down.
3. If validation fails, the previous configuration remains active and an error is logged.

This allows updating claim mappings, validation rules, and external claims sources without restarting `padlok`.

## Validation

`padlok` validates the configuration at startup and on every reload. Validation covers both the upstream Kubernetes fields (delegated to the upstream validation logic) and `padlok`-specific fields like `externalClaimsSources`. If the configuration is invalid, `padlok` will fail to start (on initial load) or log an error and keep the previous configuration (on reload).

See [Configuration Schema](configuration-schema.md) for the full field-level validation rules.
