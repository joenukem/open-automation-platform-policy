// Package policy defines the platform policy input contract (PolicyContext) and
// the decision types returned by the policy engine.
package policy

// Action is the enforceable platform action a PolicyContext describes.
type Action string

const (
	ActionJobLaunch        Action = "job.launch"
	ActionWorkflowLaunch   Action = "workflow.launch"
	ActionRulebookActivate Action = "rulebook.activate"
	ActionContentImport    Action = "content.import"
	ActionContentPromote   Action = "content.promote"
	ActionSyncRun          Action = "sync.run"
)

// Actor is the identity requesting the action, as resolved by the gateway.
type Actor struct {
	Username    string   `json:"username"`
	Teams       []string `json:"teams,omitempty"`
	IsSuperuser bool     `json:"is_superuser,omitempty"`
}

// Target describes what the action operates on.
type Target struct {
	Inventory   string   `json:"inventory,omitempty"`
	Credentials []string `json:"credentials,omitempty"`
	Limit       string   `json:"limit,omitempty"`
}

// Job carries job/workflow launch attributes relevant to policy.
type Job struct {
	Template      string `json:"template,omitempty"`
	Forks         int    `json:"forks,omitempty"`
	SurveyEnabled bool   `json:"survey_enabled,omitempty"`
	Privileged    bool   `json:"privileged,omitempty"`
}

// Content carries collection/content attributes relevant to policy.
type Content struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	Signed  bool   `json:"signed,omitempty"`
	Source  string `json:"source,omitempty"`
}

// PolicyContext is the normalized input evaluated for every enforceable action.
// It is the stable contract between the gateway pre-launch hook, the MCP server,
// and the policy engine.
type PolicyContext struct {
	Action  Action  `json:"action"`
	Actor   Actor   `json:"actor"`
	Org     string  `json:"org,omitempty"`
	Target  Target  `json:"target,omitempty"`
	Job     Job     `json:"job,omitempty"`
	Content Content `json:"content,omitempty"`
}

// Denial is a single policy that rejected the action, with a human reason.
type Denial struct {
	Policy string `json:"policy"`
	Reason string `json:"reason"`
}

// Decision is the aggregated result of evaluating all policy packs.
type Decision struct {
	Allow   bool     `json:"allow"`
	Denials []Denial `json:"denials"`
}
