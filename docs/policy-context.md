# PolicyContext Input Contract

`PolicyContext` is the normalized input the gateway pre-launch hook and the MCP
server send to `POST /v1/policy/decision`. It is the stable contract between the
platform edge and the policy engine.

## Shape

```json
{
  "action": "job.launch | workflow.launch | rulebook.activate | content.import | content.promote | sync.run",
  "actor":  { "username": "string", "teams": ["string"], "is_superuser": false },
  "org":    "string",
  "target": { "inventory": "string", "credentials": ["string"], "limit": "string" },
  "job":    { "template": "string", "forks": 0, "survey_enabled": false, "privileged": false },
  "content":{ "name": "string", "version": "string", "signed": false, "source": "string" }
}
```

Only `action` is required. Fields irrelevant to an action may be omitted; policy
packs that reference an absent field simply do not fire (Rego treats missing
references as undefined, so an unrelated policy cannot accidentally deny).

## Field Notes

- `actor.teams` drive team-based policies (e.g. `prod-inventory`). The gateway
  populates them from the authenticated session / centralized RBAC (HP-9).
- `job.forks` and `job.privileged` come from the resolved job template + launch
  request; `survey_enabled` reflects the effective survey state.
- `content.source` is the origin URL used by `untrusted-source`; `content.signed`
  is the verified signature state used by `unsigned-content`.

## Decision Response

```json
{ "allow": true, "denials": [] }
{ "allow": false, "denials": [ { "policy": "prod-inventory", "reason": "..." } ] }
```

Callers must treat any non-200 response, and any `allow: false`, as a block.
