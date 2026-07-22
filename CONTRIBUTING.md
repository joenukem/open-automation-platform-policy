# Contributing

Keep changes scoped to the policy service contract. Policy packs must live in
the `platform` Rego package and contribute `deny` objects; `decision.rego`
remains the single aggregation entrypoint.

Before submitting changes, run:

```sh
scripts/local-validation
```

Use only public/community/source-built images. Do not add authenticated Red Hat
registry dependencies, mutable `latest` image tags, or secrets in policy input,
decision logs, documentation, examples, or tests.
