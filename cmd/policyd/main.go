// Command policyd is the open automation platform policy-as-code decision
// service. It loads Rego policy packs and serves allow/deny decisions for
// enforceable platform actions (job launch, workflow, activation, import,
// promotion, sync) to the gateway pre-launch hook and the MCP server.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joenukem/open-automation-platform-policy/internal/api"
	"github.com/joenukem/open-automation-platform-policy/internal/audit"
	"github.com/joenukem/open-automation-platform-policy/internal/policy"
)

func main() {
	addr := flag.String("addr", envOr("POLICY_ADDR", ":8181"), "listen address")
	dir := flag.String("policies", envOr("POLICY_DIR", "./policies"), "directory of Rego policy packs and data")
	flag.Parse()

	ctx := context.Background()
	engine, err := policy.NewEngine(ctx, *dir)
	if err != nil {
		log.Fatalf("load policies from %s: %v", *dir, err)
	}
	log.Printf("policy: loaded packs from %s", *dir)

	logger := audit.NewLog(2000, os.Stdout)
	srv := &http.Server{
		Addr:              *addr,
		Handler:           api.New(engine, logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("policy: listening on %s", *addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("serve: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Printf("policy: stopped")
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
