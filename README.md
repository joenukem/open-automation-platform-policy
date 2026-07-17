# Open Automation Platform Policy

Policy-as-code decision service for the open automation platform. It is the open
implementation of AAP 2.7 **policy enforcement**: an Open Policy Agent (OPA) /
Rego engine that evaluates a normalized `PolicyContext` before any enforceable
platform action runs — job launch, workflow launch, rulebook activation, content
import, content promotion, and sync.

The gateway calls this service as a **pre-launch admission hook** and denies the
action before proxying it to the backend when a policy rejects it. The MCP server
calls the same endpoint so agent-initiated actions are governed identically.

## Decision Flow

```
gateway / MCP  --PolicyContext-->  policyd  --data.platform.decision-->  {allow, denials[]}
                                     |
                                     +--> decision log (audit event, stdout JSON + /v1/policy/decisions)
```

Fail-closed: a malformed request is rejected (never an implicit allow), and a
policy that produces no decision denies the action.

## API

- `POST /v1/policy/decision` — body is a `PolicyContext`; returns
  `{"allow": bool, "denials": [{"policy", "reason"}]}`.
- `GET /v1/policy/decisions?limit=N` — recent decisions as normalized events.
- `GET /healthz`, `GET /readyz` — liveness/readiness.

See [docs/policy-context.md](docs/policy-context.md) for the input contract.

## Starter Policy Packs

Bundled in `policies/` (package `platform`), configured by `policies/data.json`:

- `unsigned-content` — block promotion of unsigned collections.
- `prod-inventory` — restrict production-inventory launches to approved teams.
- `survey-required` — require a survey for privileged job templates.
- `fork-cap` — cap concurrent forks per launch.
- `untrusted-source` — deny imports from sources not on the trusted list.

Add a `.rego` file in package `platform` that contributes `deny` objects to add a
policy; `decision.rego` aggregates them.

## Run

```sh
go run ./cmd/policyd -addr :8181 -policies ./policies
# or
POLICY_ADDR=:8181 POLICY_DIR=/policies ./policyd
```

## Validate

```sh
scripts/local-validation   # gofmt, go vet, go test (real Rego evaluation)
```

## Container / k3s

```sh
podman build -t open-automation-platform-policy:dev .
kubectl apply -f deploy/k3s/policy.yaml
```

Targets k3s; uses only public/community/source-built images; no authenticated
Red Hat registry inputs.
