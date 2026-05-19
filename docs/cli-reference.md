# CLI Reference

## `padlok run`

Starts the `padlok` webhook token authentication server.

```bash
padlok run [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | string | `""` | Path to the `AuthenticationConfiguration` file. Required. The file is watched for changes and automatically reloaded. |
| `--secure-port` | string | `6443` | The port on which to serve HTTPS. |
| `--tls-cert-file` | string | `tls.crt` | Path to the TLS certificate file for the HTTPS server. |
| `--tls-private-key-file` | string | `tls.key` | Path to the TLS private key file for the HTTPS server. |
| `--tls-min-version` | string | (runtime default) | Minimum TLS version for the server. If not specified, a default is used. |
| `--tls-cipher-suites` | string array | (runtime default) | TLS cipher suites for the server. If not specified, a default set is used. |
