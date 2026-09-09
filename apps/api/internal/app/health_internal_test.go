package app

import (
	"context"
	"testing"
	"time"
)

func TestReadinessChecker_DoesNotStarveLaterChecksWhenEarlierCheckIsSlow(t *testing.T) {
	const slowCheck = 1900 * time.Millisecond

	checker := readinessChecker{
		checks: []dependencyCheck{
			{
				name: "database",
				run: func(ctx context.Context) error {
					timer := time.NewTimer(slowCheck)
					defer timer.Stop()
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-timer.C:
						return nil
					}
				},
			},
			{name: "schema", run: func(context.Context) error { return nil }},
			{name: "object_store", run: func(context.Context) error { return nil }},
		},
	}

	ready, checks := checker.evaluate(context.Background())
	if !ready {
		t.Fatalf("expected ready=true, checks=%v", checks)
	}
	for _, name := range []string{"database", "schema", "object_store"} {
		if checks[name] != "ok" {
			t.Fatalf("check %q=%q; shared deadline would starve later checks", name, checks[name])
		}
	}
}

func TestReadinessChecker_FailsWhenCheckExceedsPerCheckTimeout(t *testing.T) {
	checker := readinessChecker{
		checks: []dependencyCheck{
			{
				name: "database",
				run: func(ctx context.Context) error {
					timer := time.NewTimer(3 * time.Second)
					defer timer.Stop()
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-timer.C:
						return nil
					}
				},
			},
			{name: "schema", run: func(context.Context) error { return nil }},
		},
	}

	ready, checks := checker.evaluate(context.Background())
	if ready {
		t.Fatal("expected ready=false when a check exceeds its timeout")
	}
	if checks["database"] != "failed" {
		t.Fatalf("database check=%q", checks["database"])
	}
	if checks["schema"] != "ok" {
		t.Fatalf("schema check=%q", checks["schema"])
	}
}
