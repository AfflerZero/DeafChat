# Security TODO and Best Practices

## TODO

- [x] Remove hardcoded SMTP defaults from code.
- [ ] Rotate any previously exposed SMTP credentials and store them only in secret management.
- [x] Add room validation/limits and websocket room/client/global quotas.
- [x] Harden rate-limit client IP extraction with trusted-proxy controls.
- [x] Add CSP, Referrer-Policy, and Permissions-Policy response headers.
- [x] Handle crypto RNG errors when generating websocket client IDs.
- [x] Add CI security checks (CodeQL, dependency review, secret scanning, Dependabot).

## Best practices

- Keep secrets out of source code and never provide real credentials as default config values.
- Validate and constrain all external input (format, length, and allowable values).
- Add explicit abuse controls for real-time systems (per-room caps, global caps, idle cleanup).
- Trust forwarding headers only from known proxy CIDRs.
- Use defense-in-depth HTTP security headers and HTTPS in production.
- Run automated dependency, static security, and secret scanning checks in CI on every PR.
