# Quickstart

This guide walks through a minimal `padlok` setup that authenticates OIDC tokens and maps the `sub` claim to a Kubernetes username.

## Prerequisites

- A running OIDC identity provider with a known issuer URL and client ID.
- TLS certificates for `padlok` (see [TLS Configuration](tls-configuration.md)).
- `padlok` installed (see [Installation](installation.md)).

## 1. Create the padlok Configuration

Create a file called `config.yaml`:

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

This configuration:
- Accepts tokens issued by `https://idp.example.com`.
- Requires the token's `aud` claim to include `my-k8s-cluster`.
- Maps the `sub` claim directly to the Kubernetes username with no prefix.

If your OIDC provider uses a custom or self-signed CA, add the `certificateAuthority` field to the issuer block:

```yaml
    issuer:
      url: https://idp.example.com
      audiences:
        - my-k8s-cluster
      certificateAuthority: |
        -----BEGIN CERTIFICATE-----
        <PEM-encoded CA certificate>
        -----END CERTIFICATE-----
```

## 2. Start padlok

```bash
padlok run \
  --config config.yaml \
  --tls-cert-file tls.crt \
  --tls-private-key-file tls.key
```

`padlok` starts an HTTPS server on port 6443 (the default).

## 3. Configure the Kubernetes API Server

Create a webhook config file (`webhook-config.yaml`):

```yaml
apiVersion: v1
kind: Config

clusters:
- cluster:
    server: https://<padlok-host>:6443/authenticate
    certificate-authority: /path/to/padlok-ca.pem
  name: padlok

users:
- name: apiserver

current-context: webhook
contexts:
- context:
    cluster: padlok
    user: apiserver
  name: webhook
```

Add the following flags to the API server:

```
--authentication-token-webhook-config-file=/path/to/webhook-config.yaml
--authentication-token-webhook-version=v1
```

See the [Kubernetes webhook token authentication documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication) for more details.

## 4. Test Authentication

Obtain a token from your OIDC provider and use it with `kubectl`:

```bash
kubectl --token="<your-jwt-token>" auth whoami
```

If authentication succeeds, the output will show the Kubernetes identity that `padlok` mapped from the token — including the username, groups, uid, and any extra attributes.

## Next Steps

- [Claim Mappings](claim-mappings.md) — Map groups, uid, and extra attributes.
- [External Claims Sources](external-claims-sources.md) — Fetch claims from external endpoints.
- [Validation Rules](validation-rules.md) — Add claim and user validation.
