package policy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/open-policy-agent/opa/rego"
)

// query is the single aggregation entrypoint every policy pack contributes to.
const query = "data.platform.decision"

// Engine evaluates a PolicyContext against the loaded Rego policy packs.
type Engine struct {
	prepared rego.PreparedEvalQuery
}

// NewEngine loads all .rego and data files from dir and prepares the decision
// query. The policy packs all live in the `platform` package and contribute to
// a shared `deny` set aggregated by decision.rego.
func NewEngine(ctx context.Context, dir string) (*Engine, error) {
	r := rego.New(
		rego.Query(query),
		rego.Load([]string{dir}, nil),
	)
	pq, err := r.PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("prepare policy query: %w", err)
	}
	return &Engine{prepared: pq}, nil
}

// Evaluate runs the policy packs against pc and returns the aggregated decision.
func (e *Engine) Evaluate(ctx context.Context, pc PolicyContext) (Decision, error) {
	// Round-trip through JSON so Rego sees plain maps with json tag names.
	raw, err := json.Marshal(pc)
	if err != nil {
		return Decision{}, fmt.Errorf("marshal policy context: %w", err)
	}
	var input map[string]any
	if err := json.Unmarshal(raw, &input); err != nil {
		return Decision{}, fmt.Errorf("unmarshal policy context: %w", err)
	}

	rs, err := e.prepared.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return Decision{}, fmt.Errorf("evaluate policy: %w", err)
	}
	if len(rs) == 0 || len(rs[0].Expressions) == 0 {
		// No decision document produced: fail closed.
		return Decision{Allow: false, Denials: []Denial{{
			Policy: "engine",
			Reason: "policy produced no decision (fail closed)",
		}}}, nil
	}

	// The decision expression is a map[string]any; re-marshal into Decision.
	val := rs[0].Expressions[0].Value
	out, err := json.Marshal(val)
	if err != nil {
		return Decision{}, fmt.Errorf("marshal decision: %w", err)
	}
	var d Decision
	if err := json.Unmarshal(out, &d); err != nil {
		return Decision{}, fmt.Errorf("unmarshal decision: %w", err)
	}
	if d.Denials == nil {
		d.Denials = []Denial{}
	}
	return d, nil
}
