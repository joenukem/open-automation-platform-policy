package policy

import (
	"context"
	"sort"
	"testing"
)

func newTestEngine(t *testing.T) *Engine {
	t.Helper()
	e, err := NewEngine(context.Background(), "../../policies")
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	return e
}

func denialPolicies(d Decision) []string {
	out := make([]string, 0, len(d.Denials))
	for _, dn := range d.Denials {
		out = append(out, dn.Policy)
	}
	sort.Strings(out)
	return out
}

func TestEvaluate(t *testing.T) {
	e := newTestEngine(t)

	tests := []struct {
		name      string
		pc        PolicyContext
		wantAllow bool
		wantDeny  []string
	}{
		{
			name: "clean job launch is allowed",
			pc: PolicyContext{
				Action: ActionJobLaunch,
				Actor:  Actor{Username: "dev", Teams: []string{"developers"}},
				Target: Target{Inventory: "staging"},
				Job:    Job{Template: "deploy", Forks: 10},
			},
			wantAllow: true,
			wantDeny:  []string{},
		},
		{
			name: "forks over cap denied",
			pc: PolicyContext{
				Action: ActionJobLaunch,
				Actor:  Actor{Username: "dev", Teams: []string{"developers"}},
				Target: Target{Inventory: "staging"},
				Job:    Job{Template: "deploy", Forks: 100},
			},
			wantAllow: false,
			wantDeny:  []string{"fork-cap"},
		},
		{
			name: "prod inventory with unapproved team denied",
			pc: PolicyContext{
				Action: ActionJobLaunch,
				Actor:  Actor{Username: "dev", Teams: []string{"developers"}},
				Target: Target{Inventory: "prod-db"},
				Job:    Job{Template: "deploy", Forks: 10},
			},
			wantAllow: false,
			wantDeny:  []string{"prod-inventory"},
		},
		{
			name: "prod inventory with approved team allowed",
			pc: PolicyContext{
				Action: ActionJobLaunch,
				Actor:  Actor{Username: "ops", Teams: []string{"sre"}},
				Target: Target{Inventory: "prod-db"},
				Job:    Job{Template: "deploy", Forks: 10},
			},
			wantAllow: true,
			wantDeny:  []string{},
		},
		{
			name: "privileged template without survey denied",
			pc: PolicyContext{
				Action: ActionJobLaunch,
				Actor:  Actor{Username: "ops", Teams: []string{"sre"}},
				Target: Target{Inventory: "staging"},
				Job:    Job{Template: "root-patch", Forks: 5, Privileged: true, SurveyEnabled: false},
			},
			wantAllow: false,
			wantDeny:  []string{"survey-required"},
		},
		{
			name: "privileged template with survey allowed",
			pc: PolicyContext{
				Action: ActionJobLaunch,
				Actor:  Actor{Username: "ops", Teams: []string{"sre"}},
				Target: Target{Inventory: "staging"},
				Job:    Job{Template: "root-patch", Forks: 5, Privileged: true, SurveyEnabled: true},
			},
			wantAllow: true,
			wantDeny:  []string{},
		},
		{
			name: "promote unsigned content denied",
			pc: PolicyContext{
				Action:  ActionContentPromote,
				Actor:   Actor{Username: "curator", Teams: []string{"content"}},
				Content: Content{Name: "acme.utils", Version: "1.0.0", Signed: false},
			},
			wantAllow: false,
			wantDeny:  []string{"unsigned-content"},
		},
		{
			name: "promote signed content allowed",
			pc: PolicyContext{
				Action:  ActionContentPromote,
				Actor:   Actor{Username: "curator", Teams: []string{"content"}},
				Content: Content{Name: "acme.utils", Version: "1.0.0", Signed: true},
			},
			wantAllow: true,
			wantDeny:  []string{},
		},
		{
			name: "import from untrusted source denied",
			pc: PolicyContext{
				Action:  ActionContentImport,
				Actor:   Actor{Username: "curator", Teams: []string{"content"}},
				Content: Content{Name: "sketchy.thing", Source: "https://evil.example.com/x"},
			},
			wantAllow: false,
			wantDeny:  []string{"untrusted-source"},
		},
		{
			name: "import from trusted source allowed",
			pc: PolicyContext{
				Action:  ActionContentImport,
				Actor:   Actor{Username: "curator", Teams: []string{"content"}},
				Content: Content{Name: "community.general", Source: "https://github.com/ansible-collections/community.general"},
			},
			wantAllow: true,
			wantDeny:  []string{},
		},
		{
			name: "multiple violations produce multiple denials",
			pc: PolicyContext{
				Action: ActionJobLaunch,
				Actor:  Actor{Username: "dev", Teams: []string{"developers"}},
				Target: Target{Inventory: "prod-web"},
				Job:    Job{Template: "root-patch", Forks: 200, Privileged: true, SurveyEnabled: false},
			},
			wantAllow: false,
			wantDeny:  []string{"fork-cap", "prod-inventory", "survey-required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := e.Evaluate(context.Background(), tt.pc)
			if err != nil {
				t.Fatalf("Evaluate: %v", err)
			}
			if d.Allow != tt.wantAllow {
				t.Errorf("allow = %v, want %v (denials: %v)", d.Allow, tt.wantAllow, denialPolicies(d))
			}
			got := denialPolicies(d)
			if len(got) != len(tt.wantDeny) {
				t.Fatalf("denials = %v, want %v", got, tt.wantDeny)
			}
			for i := range got {
				if got[i] != tt.wantDeny[i] {
					t.Errorf("denials = %v, want %v", got, tt.wantDeny)
					break
				}
			}
		})
	}
}
