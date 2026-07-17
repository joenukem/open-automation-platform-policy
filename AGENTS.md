# Repository Instructions

- This repo implements AAP 2.7 policy enforcement (HP-8) for the open platform.
- Keep the gateway pre-launch hook contract stable: `POST /v1/policy/decision`
  takes a `PolicyContext` and returns `{allow, denials[]}`.
- Fail closed. A malformed request or a missing decision is a deny, never an
  implicit allow.
- All policy packs live in the `platform` Rego package and contribute `deny`
  objects; `decision.rego` aggregates. Do not add a second query entrypoint.
- Configuration (approved teams, caps, trusted sources) lives in
  `policies/data.json`, not hard-coded in Rego.
- Do not store secrets in policy input or decision logs.
- Use only public/community/source-built images; no authenticated Red Hat
  registries.
- Run `scripts/local-validation` before handoff.
