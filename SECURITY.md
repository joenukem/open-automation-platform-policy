# Security Policy

Report suspected vulnerabilities privately to the project maintainers before
opening a public issue or pull request.

Do not include secrets, access tokens, kubeconfigs, private inventory data, or
policy inputs containing sensitive values in reports, fixtures, logs, or tests.

This repository fails closed for policy decisions. Security-sensitive changes
must preserve `POST /v1/policy/decision`, keep policy configuration in
`policies/data.json`, and run `scripts/local-validation` before handoff.
