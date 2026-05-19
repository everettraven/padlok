# Troubleshooting

When diagnosing authentication issues with `padlok`, consult two sources of logs:

- **`padlok` logs** — `padlok` uses [klog](https://github.com/kubernetes/klog) for structured logging. Look here for configuration validation errors, JWT validation failures, external claims resolution errors, and configuration reload events.
- **Kubernetes API server logs** — The API server logs webhook authenticator connection errors, `TokenReview` failures, and authorization decisions. Look here when the API server cannot reach `padlok` or when authentication succeeds but authorization fails.

Together, these two log sources cover the full authentication path — from the API server sending a `TokenReview` to `padlok`, through token validation and claims resolution, to the API server acting on the response.
