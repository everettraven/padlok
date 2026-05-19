# Validation Rules

Validation rules are CEL expressions that must evaluate to `true` for authentication to succeed. There are two types: claim validation rules (evaluated before identity mapping) and user validation rules (evaluated after identity mapping). For background on the upstream behavior, see the [Kubernetes validation rules documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#claim-validation-rules).

## Claim Validation Rules

Claim validation rules validate token claims before they are mapped to a Kubernetes identity. They have access to the `claims` variable.

Each rule must produce a boolean. If any rule returns `false`, authentication is rejected with the configured `message`.

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `expression` | string | Yes (mutually exclusive with `claim`) | A CEL expression that must return `true`. Has access to `claims`. |
| `message` | string | No | Custom error message returned when the expression returns `false`. |
| `claim` | string | Yes (mutually exclusive with `expression`) | The name of a required claim. |
| `requiredValue` | string | No | When used with `claim`, the claim must have this exact value. If not set, the claim must be present with an empty string value. |

The `claim`/`requiredValue` form and the `expression`/`message` form are mutually exclusive.

### Examples

Require that the email is verified:

```yaml
claimValidationRules:
  - expression: "has(claims.email_verified) && claims.email_verified == true"
    message: "email must be verified"
```

Require a specific claim value using the `claim`/`requiredValue` shorthand:

```yaml
claimValidationRules:
  - claim: "tenant"
    requiredValue: "acme-corp"
```

Require that the token was issued within the last hour:

```yaml
claimValidationRules:
  - expression: "claims.iat > timestamp(now - duration('1h'))"
    message: "token must have been issued within the last hour"
```

## User Validation Rules

User validation rules validate the final mapped Kubernetes identity after claim mappings have been applied. They have access to the `user` variable, which is a [UserInfo](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#userinfo-v1-authentication-k8s-io) object with fields: `username`, `uid`, `groups`, and `extra`.

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `expression` | string | Yes | A CEL expression that must return `true`. Has access to `user`. |
| `message` | string | No | Custom error message returned when the expression returns `false`. |

### Examples

Prevent identities with the `system:` prefix (reserved for Kubernetes components):

```yaml
userValidationRules:
  - expression: "!user.username.startsWith('system:')"
    message: "username must not have the system: prefix"
```

Require that the user belongs to at least one group:

```yaml
userValidationRules:
  - expression: "user.groups.size() > 0"
    message: "user must belong to at least one group"
```

## Combining Both

Claim validation rules and user validation rules can be used together. Claim validation rules run first (against the raw token claims), then claim mappings are applied, then user validation rules run against the mapped identity.

```yaml
jwt:
  - issuer:
      url: https://idp.example.com
      audiences:
        - my-k8s-cluster
    claimValidationRules:
      - expression: "has(claims.email_verified) && claims.email_verified == true"
        message: "email must be verified"
    claimMappings:
      username:
        claim: "email"
        prefix: ""
      groups:
        expression: "claims.groups.split(',')"
    userValidationRules:
      - expression: "!user.username.startsWith('system:')"
        message: "system: prefix is reserved"
```
