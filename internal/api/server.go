// Package api exposes the policy decision HTTP surface consumed by the gateway
// pre-launch hook and the MCP server.
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/joenukem/open-automation-platform-policy/internal/audit"
	"github.com/joenukem/open-automation-platform-policy/internal/policy"
)

// Evaluator is the subset of the policy engine the server needs.
type Evaluator interface {
	Evaluate(ctx context.Context, pc policy.PolicyContext) (policy.Decision, error)
}

// Server wires the evaluator and decision log into an http.Handler.
type Server struct {
	engine Evaluator
	log    *audit.Log
	mux    *http.ServeMux
}

// New builds the policy API handler.
func New(engine Evaluator, log *audit.Log) *Server {
	s := &Server{engine: engine, log: log, mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("GET /readyz", s.handleHealth)
	s.mux.HandleFunc("POST /v1/policy/decision", s.handleDecision)
	s.mux.HandleFunc("GET /v1/policy/decisions", s.handleDecisions)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleDecision evaluates a PolicyContext and records the outcome. The action
// is always allowed or denied explicitly; a malformed request is a 400 and is
// not treated as an allow.
func (s *Server) handleDecision(w http.ResponseWriter, r *http.Request) {
	var pc policy.PolicyContext
	if err := json.NewDecoder(r.Body).Decode(&pc); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid policy context: " + err.Error()})
		return
	}
	if pc.Action == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "action is required"})
		return
	}
	d, err := s.engine.Evaluate(r.Context(), pc)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.record(pc, d)
	writeJSON(w, http.StatusOK, d)
}

func (s *Server) record(pc policy.PolicyContext, d policy.Decision) {
	if s.log == nil {
		return
	}
	if d.Allow {
		s.log.Record(audit.Event{
			Action:  string(pc.Action),
			Actor:   pc.Actor.Username,
			Org:     pc.Org,
			Outcome: "allow",
		})
		return
	}
	for _, dn := range d.Denials {
		s.log.Record(audit.Event{
			Action:        string(pc.Action),
			Actor:         pc.Actor.Username,
			Org:           pc.Org,
			Outcome:       "deny",
			MatchedPolicy: dn.Policy,
			Reason:        dn.Reason,
		})
	}
}

func (s *Server) handleDecisions(w http.ResponseWriter, r *http.Request) {
	n := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			n = parsed
		}
	}
	events := []audit.Event{}
	if s.log != nil {
		events = s.log.Recent(n)
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": events})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
