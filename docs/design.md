# Open Automation Platform Policy Design

## AAP 2.7 Scope

This repo owns the policy decision surface used by gateway, MCP, Developer Hub,
catalog, and workflow actions. The policy service does not own Controller, Hub,
EDA, portal, or workflow resource state; it returns allow/deny decisions based
on the request context and configured policy data.

## Required Runtime Contract

| Surface | Required behavior | Release evidence |
| --- | --- | --- |
| Decision API | Serve `POST /v1/policy/decision` with a stable `PolicyContext` and `{allow, denials[]}` response. | Positive allow, explicit deny, malformed request deny, and audit correlation. |
| Gateway integration | Gateway and MCP mutating paths call policy before upstream mutation and fail closed on missing policy response. | Policy preflight evidence, denied mutation blocked, allowed mutation reaches upstream. |
| Policy data | Store approved teams, caps, trusted sources, and environment rules in policy data, not hard-coded logic. | Data-driven decision proof and policy data checksum. |
| Audit and redaction | Log correlation IDs, actor, service, action, decision, and denial reasons without secrets or payload values. | Redacted audit report and no secret leakage. |

## Forbidden Completion Policy

No mock services, fake upstreams, stub responses, canned responses,
scaffold-only handlers, placeholder-only resources, render-only parity,
dry-run-only parity, or verify-on-existence checks can satisfy product
completion, release readiness, deployability, or AAP 2.7 parity. Unit tests may
use isolated test doubles only for code behavior, never as parity evidence.
