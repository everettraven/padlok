# TLS Configuration

`padlok` requires TLS for its HTTPS server and supports configuring TLS for outbound connections to OIDC providers and external claims sources.

## Server TLS

The `padlok` server requires a TLS certificate and private key. These are specified via CLI flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--tls-cert-file` | `tls.crt` | Path to the TLS certificate file. |
| `--tls-private-key-file` | `tls.key` | Path to the TLS private key file. |
| `--tls-min-version` | (runtime default) | Minimum TLS version. Must be one of the [supported TLS versions](https://pkg.go.dev/crypto/tls#pkg-constants) (e.g., `VersionTLS12`, `VersionTLS13`). |
| `--tls-cipher-suites` | (runtime default) | Comma-separated list of TLS cipher suites. |

### Generating Self-Signed Certificates

For development or testing, generate a self-signed certificate with OpenSSL:

```bash
openssl req -x509 -newkey rsa:4096 \
  -keyout tls.key -out tls.crt \
  -sha256 -days 365 -nodes \
  -subj "/CN=padlok" \
  -addext "subjectAltName=DNS:padlok"
```

Set the `CN` and `subjectAltName` to the hostname that the Kubernetes API server will use to reach `padlok`. The API server's webhook config must trust this certificate (or its CA) via the `certificate-authority` field.

## OIDC Issuer TLS

When connecting to an OIDC provider for discovery and JWKS fetching, `padlok` uses the system certificate store by default. For providers with self-signed or internal CA certificates, specify the CA in the issuer configuration:

```yaml
jwt:
  - issuer:
      url: https://idp.internal.example.com
      certificateAuthority: |
        -----BEGIN CERTIFICATE-----
        <PEM-encoded CA certificate>
        -----END CERTIFICATE-----
```

When `discoveryURL` is set, the `certificateAuthority` is used to verify the connection to the discovery URL. The leaf certificate must have a hostname matching the discovery URL's host.

See [Configuring OIDC Issuers](configuring-oidc-issuers.md) for details.

## External Claims Source TLS

Each external claims source can specify its own TLS CA certificate for the connection to the external endpoint:

```yaml
externalClaimsSources:
  - tls:
      certificateAuthority: |
        -----BEGIN CERTIFICATE-----
        <PEM-encoded CA certificate>
        -----END CERTIFICATE-----
    url: { ... }
    mappings: [ ... ]
```

When using `ClientCredential` authentication, the token endpoint connection can also have its own TLS configuration:

```yaml
externalClaimsSources:
  - authentication:
      type: ClientCredential
      clientCredential:
        clientID: "my-client"
        clientSecret: "my-secret"
        tokenEndpoint: "https://auth.example.com/oauth2/token"
        tls:
          certificateAuthority: |
            -----BEGIN CERTIFICATE-----
            <PEM-encoded CA certificate for the token endpoint>
            -----END CERTIFICATE-----
    tls:
      certificateAuthority: |
        -----BEGIN CERTIFICATE-----
        <PEM-encoded CA certificate for the external source>
        -----END CERTIFICATE-----
    url: { ... }
    mappings: [ ... ]
```

The top-level `tls` applies to the external claims endpoint. The `clientCredential.tls` applies to the OAuth2 token endpoint. These can use different CAs if the endpoints are served by different infrastructure.

See [External Claims Sources](external-claims-sources.md) for details.
