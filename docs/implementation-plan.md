# Implementation Plan

Implements HP-8 (policy-as-code enforcement) from
`open-automation-platform-research`.

## Phase 0 (done)

- PolicyContext input contract and decision types.
- OPA/Rego engine wrapper with a single aggregation entrypoint
  (`data.platform.decision`).
- Starter policy packs: unsigned-content, prod-inventory, survey-required,
  fork-cap, untrusted-source.
- Decision API (`/v1/policy/decision`, `/v1/policy/decisions`, health).
- Decision log as normalized audit events (stdout JSON + ring buffer).
- Unit tests evaluating real Rego; container image; k3s manifest.

## Phase 1

- Gateway pre-launch hook integration (owner: `awx-gateway`): build a
  PolicyContext for job launch, workflow, activation, import, promotion, and sync
  and deny before proxying (AAP-POLICY-002).
- Console policy views: policy list/detail, decision log, "why denied"
  drill-down (AAP-POLICY-003).

## Later

- Versioned policy bundles and a bundle store (OPA bundle API) so policy packs
  ship and update without redeploying the service.
- Policy dry-run endpoint used by the MCP server before an agent action.
- Emit decisions to the observability event pipeline (HP-11) for the dashboard.
- Per-org policy overlays and exemptions with audit.
