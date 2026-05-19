# Claim Mappings

Claim mappings define how JWT token claims (and any externally sourced claims) are transformed into a Kubernetes identity. For background on the upstream behavior, see the [Kubernetes claim mappings documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#claim-mappings).

## Identity Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `username` | `PrefixedClaimOrExpression` | Yes | The Kubernetes username. Must produce a string value. |
| `groups` | `PrefixedClaimOrExpression` | No | The Kubernetes groups. Must produce a string or string array. Empty values, `[]`, and `null` are treated as no groups. |
| `uid` | `ClaimOrExpression` | No | The Kubernetes user ID. Must produce a string value. |
| `extra` | list of `ExtraMapping` | No | Arbitrary key-value metadata attached to the identity. Each entry produces a string or string array. |

## Claim vs. Expression

Each mapping attribute (except `extra`) can be set using either a direct **claim** reference or a CEL **expression**. These are mutually exclusive.

### Using `claim` (with optional `prefix`)

The `claim` field names a JWT claim directly. The `prefix` field is prepended to the claim value. When using `claim`, `prefix` is required (set to `""` for no prefix).

```yaml
claimMappings:
  username:
    claim: "sub"
    prefix: ""
  groups:
    claim: "groups"
    prefix: ""
```

### Using `expression`

The `expression` field is a [CEL expression](cel-expressions.md) that produces the attribute value. Token claims are available via the `claims` variable.

```yaml
claimMappings:
  username:
    expression: "claims.email"
  groups:
    expression: "claims.roles.filter(r, r.startsWith('k8s-'))"
```

## Username

The `username` mapping is required. It must produce a singular string value.

When using `claim: "email"`, you should verify that the email is verified by including a validation rule. This mirrors the behavior Kubernetes applies automatically when `--oidc-username-claim=email` is used:

```yaml
claimMappings:
  username:
    claim: "email"
    prefix: ""
claimValidationRules:
  - expression: "has(claims.email_verified) && claims.email_verified == true"
    message: "email must be verified"
```

## Groups

The `groups` mapping is optional. The claim's value must be a string or string array. When using a CEL expression, it must produce a string or string array value.

```yaml
claimMappings:
  groups:
    expression: "claims.groups.split(',')"
```

## UID

The `uid` mapping is optional. It must produce a singular string value.

```yaml
claimMappings:
  uid:
    claim: "sub"
```

## Extra

The `extra` field is a list of key-value mappings. Each entry has a `key` and a `valueExpression` (CEL expression).

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `key` | string | Yes | The extra attribute key. Must be a domain-prefix path (e.g., `example.org/foo`), lowercase, and unique across all extra mappings. |
| `valueExpression` | string | Yes | A CEL expression producing a string or string array. Empty values are treated as the mapping not being present. |

```yaml
claimMappings:
  extra:
    - key: "example.org/department"
      valueExpression: "claims.department"
    - key: "example.org/admin"
      valueExpression: '(has(claims.is_admin) && claims.is_admin) ? "true" : ""'
```

In the second example, the `admin` extra attribute is only present when the `is_admin` claim is `true`.

## Full Example

```yaml
jwt:
  - issuer:
      url: https://idp.example.com
      audiences:
        - my-k8s-cluster
    claimMappings:
      username:
        expression: "claims.email"
      groups:
        expression: "claims.groups.split(',')"
      uid:
        claim: "sub"
      extra:
        - key: "example.org/department"
          valueExpression: "claims.department"
    claimValidationRules:
      - expression: "has(claims.email_verified) && claims.email_verified == true"
        message: "email must be verified"
```
