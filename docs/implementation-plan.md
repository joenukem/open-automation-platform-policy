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

## Full-Stack Release Contract

This repo is release-ready only as part of the
`open-automation-platform-deploy` full-stack profile. The deploy repo owns the
Helm, OLM, and standalone k3s VM surfaces; this repo owns the fail-closed policy
decision contract consumed by those surfaces.

Release requirements:

- Gateway and MCP mutating actions call policy before Controller, Hub, or EDA
  backends; release profiles must not offer a bypass path around policy for
  enforceable actions.
- `awx-gateway` is the policy-enforcement caller for browser/API launch paths.
  It resolves the gateway session through `/api/gateway/v1/me`, builds the
  `PolicyContext`, calls `POST /v1/policy/decision`, and fails closed before
  proxying to `/api/controller/v2/`, `/api/hub/v1/`, or `/api/eda/v1/`.
- Policy decisions are correlation-ready audit events and include denials
  without logging secrets or full sensitive event payloads.
- The `profiles/full-stack-e2e` proof suite must include policy evidence when
  policy is enabled: allow and deny decisions, gateway pre-launch enforcement,
  MCP fail-closed behavior, and no direct-origin leaks from the console.
- Core deployable release evidence for the currently enabled Controller/Hub/EDA
  stack must include `gateway-full-stack-fixture-bootstrap`,
  `gateway-full-stack-fixture-preflight`, `gateway-full-stack-e2e`,
  `gateway-console-shell-proof`, `gateway-cross-surface-nav-proof`,
  `gateway-hub-collection-sync-proof`, and
  `gateway-eda-rulebook-trigger-proof`; policy-specific proof extends that
  sequence when the policy service is enabled.
- Required proof ordering is part of the release contract:
  `gateway-full-stack-fixture-bootstrap` must precede
  `fixture_bootstrap=ok`, `fixture_bootstrap=ok` must precede
  `gateway-full-stack-fixture-preflight`, fixture preflight and
  `controller_job_template_launch=ok` must precede `gateway-full-stack-e2e`,
  `gateway-full-stack-e2e` must precede `gateway-console-shell-proof`, then
  `gateway-cross-surface-nav-proof`, then `gateway-hub-collection-sync-proof`,
  then `gateway-eda-rulebook-trigger-proof`, and `controller_job=terminal`
  must precede the EDA rulebook trigger pass marker.
- Helm, OLM, and standalone k3s VM releases consume the same policy bundle,
  image, and runtime configuration. Any repo-specific image must be
  public/source-built and digest-recorded by the image-builder release evidence.
- No policy input, decision log, fixture, rendered manifest, or proof report may
  contain credentials, OAuth tokens, kubeconfigs, private keys, or generated
  session data.
