package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joenukem/open-automation-platform-policy/internal/audit"
	"github.com/joenukem/open-automation-platform-policy/internal/policy"
)

func newServer(t *testing.T) (*Server, *audit.Log) {
	t.Helper()
	e, err := policy.NewEngine(context.Background(), "../../policies")
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	log := audit.NewLog(100, nil)
	return New(e, log), log
}

func TestDecisionDeny(t *testing.T) {
	s, log := newServer(t)
	body, _ := json.Marshal(policy.PolicyContext{
		Action: policy.ActionJobLaunch,
		Actor:  policy.Actor{Username: "dev", Teams: []string{"developers"}},
		Target: policy.Target{Inventory: "prod-db"},
		Job:    policy.Job{Template: "deploy", Forks: 10},
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/policy/decision", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var d policy.Decision
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if d.Allow {
		t.Errorf("allow = true, want false")
	}
	if len(d.Denials) != 1 || d.Denials[0].Policy != "prod-inventory" {
		t.Errorf("denials = %v, want prod-inventory", d.Denials)
	}
	if got := log.Recent(10); len(got) != 1 || got[0].Outcome != "deny" || got[0].MatchedPolicy != "prod-inventory" {
		t.Errorf("audit = %+v, want one deny event for prod-inventory", got)
	}
}

func TestDecisionAllow(t *testing.T) {
	s, _ := newServer(t)
	body, _ := json.Marshal(policy.PolicyContext{
		Action: policy.ActionJobLaunch,
		Actor:  policy.Actor{Username: "dev", Teams: []string{"developers"}},
		Target: policy.Target{Inventory: "staging"},
		Job:    policy.Job{Template: "deploy", Forks: 5},
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/policy/decision", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	var d policy.Decision
	_ = json.Unmarshal(rec.Body.Bytes(), &d)
	if !d.Allow || len(d.Denials) != 0 {
		t.Errorf("allow=%v denials=%v, want allow with no denials", d.Allow, d.Denials)
	}
}

func TestDecisionBadRequest(t *testing.T) {
	s, _ := newServer(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/policy/decision", bytes.NewReader([]byte("{not json")))
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestDecisionsList(t *testing.T) {
	s, _ := newServer(t)
	body, _ := json.Marshal(policy.PolicyContext{
		Action:  policy.ActionContentPromote,
		Actor:   policy.Actor{Username: "curator"},
		Content: policy.Content{Name: "acme.utils", Signed: false},
	})
	s.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/v1/policy/decision", bytes.NewReader(body)))

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/policy/decisions", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var out struct {
		Events []audit.Event `json:"events"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Events) != 1 || out.Events[0].MatchedPolicy != "unsigned-content" {
		t.Errorf("events = %+v, want one unsigned-content deny", out.Events)
	}
}

func TestHealth(t *testing.T) {
	s, _ := newServer(t)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
