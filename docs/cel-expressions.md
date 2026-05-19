# CEL Expressions

`padlok` uses [Common Expression Language (CEL)](https://kubernetes.io/docs/reference/using-api/cel/) throughout its configuration for claim mapping, validation, and external claims resolution. For the base CEL environment available in Kubernetes authentication, see the [upstream documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#cel-expressions-in-authentication).

## Variables

Different configuration contexts provide different CEL variables:

| Context | Variable | Type | Description |
|---------|----------|------|-------------|
| Claim validation rules | `claims` | map | Token claims (and any merged external claims). |
| Claim mappings (username, groups, uid, extra) | `claims` | map | Token claims (and any merged external claims). |
| User validation rules | `user` | UserInfo | The mapped Kubernetes identity (`username`, `uid`, `groups`, `extra`). |
| External source path expressions | `claims` | map | Token claims (external claims are not yet available). |
| External source conditions | `claims` | map | Token claims (external claims are not yet available). |
| External source mappings | `response` | object | The external source's HTTP response. `response.body` contains the parsed JSON body. |

## Accessing Claims

Claims are accessed via dot notation or index notation:

```cel
claims.sub
claims.email
claims["preferred_username"]
```

Nested claims use dot notation:

```cel
claims.address.street
claims.realm_access.roles
```

## Checking for Claim Existence

Use `has()` to check whether a claim exists before accessing it:

```cel
has(claims.groups)
```

Use the optional chaining operator `?.` with `orValue()` for safe access with a default:

```cel
claims.?email_verified.orValue(false)
```

## String Operations

```cel
claims.sub                                  // direct access
"prefix:" + claims.sub                      // concatenation
claims.email.split("@")[0]                  // split and index
claims.username.startsWith("admin")         // prefix check
claims.name.lowerAscii()                    // lowercase
```

## List Operations

```cel
claims.groups.split(",")                    // string to list
claims.roles.filter(r, r.startsWith("k8s-"))  // filter
claims.groups.join(",")                     // list to string
claims.roles.size()                         // count
claims.roles.exists(r, r == "admin")        // existence check
```

## Conditional Expressions

```cel
has(claims.nickname) ? claims.nickname : claims.sub
(has(claims.is_admin) && claims.is_admin) ? "true" : ""
```

## Common Patterns

### Map `sub` to username with no prefix

```yaml
username:
  claim: "sub"
  prefix: ""
```

Or equivalently with a CEL expression:

```yaml
username:
  expression: "claims.sub"
```

### Split a comma-separated groups claim

```yaml
groups:
  expression: "claims.groups.split(',')"
```

### Filter groups by prefix

```yaml
groups:
  expression: "claims.roles.filter(r, r.startsWith('k8s:'))"
```

### Conditionally set an extra attribute

```yaml
extra:
  - key: "example.org/admin"
    valueExpression: '(has(claims.is_admin) && claims.is_admin) ? "true" : ""'
```

Empty string values cause the extra mapping to not be present.

### Extract claims from an external source response

In external source mapping expressions, the response body is available via `response.body`:

```yaml
mappings:
  - name: groups
    expression: "response.body.groups.join(',')"
  - name: department
    expression: "response.body.department"
```

### Build a dynamic URL path from claims

```yaml
url:
  hostname: api.example.com
  pathExpression: "['users'] + [claims.sub] + ['groups']"
```

If `claims.sub` is `"jane.doe"`, this produces: `https://api.example.com/users/jane.doe/groups`
