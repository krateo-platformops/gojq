package gojq_test

// delete_empty_cow_test.go — snowplow CI guard for the allocator-aware
// deleteEmpty patch (Ship 2a / 0.30.209).
//
// This fork patches deleteEmpty (func.go) so that del/delpaths over a
// value NOT owned by gojq's allocator (the caller's shared input)
// copy-on-writes instead of mutating it in place. Upstream's deleteEmpty
// recurses into the copy-on-write spine's ALIASED sibling sub-trees and
// writes them in place (`v[k] = deleteEmpty(child)`, `delete(v, k)`).
//
// CRITICAL — why this guard is CONCURRENT + -race, not a value compare:
// upstream's in-place write usually stores the SAME value back (the
// recursion is a no-op on a survivor), so a single-threaded
// reflect.DeepEqual of the input before/after CANNOT see the write. The
// write is only observable as a DATA RACE when a second goroutine reads
// the shared sub-tree concurrently — which is exactly the production
// hazard (concurrent serves of the SAME cached entry.Items). So the guard
// MUST run del/delpaths over a SHARED input from many goroutines under
// `-race`. With the patch: writes land on private CoW nodes → clean.
// Without it: writes hit the shared aliased maps → race → FAIL.
//
// snowplow relies on this so api.listEnvelopeValue can serve a SHALLOW
// envelope aliasing the cached entry.Items without a per-serve deep copy.
// If a future re-vendor of github.com/itchyny/gojq drops the patch, the
// FORK's own `go test -race ./...` must fail HERE. (Mirrors snowplow's
// TestShip2a_DeleteEmpty_CoW_NoInputMutation; that one is the in-process
// proof, this one is the in-fork guard so the patch can't silently lapse.)
//
// RUN WITH -race. The fork's CI must invoke `go test -race ./...`.

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"
	"testing"

	"github.com/itchyny/gojq"
)

// cowGuardTree builds a JSON tree where the deleted/updated path has
// SIBLINGS at every level — the shape that makes upstream's deleteEmpty
// recurse into aliased input sub-trees.
func cowGuardTree() map[string]any {
	return map[string]any{
		"a": "AAA",
		"b": map[string]any{
			"c": map[string]any{"d": 1.0, "e": "keep-e"},
			"f": "sibling-f",
			"g": map[string]any{"h": "keep-h"},
		},
		"list": []any{"x", "y", "z"},
	}
}

func cowGuardDeepCopy(v any) any {
	switch t := v.(type) {
	case map[string]any:
		c := make(map[string]any, len(t))
		for k, x := range t {
			c[k] = cowGuardDeepCopy(x)
		}
		return c
	case []any:
		c := make([]any, len(t))
		for i, x := range t {
			c[i] = cowGuardDeepCopy(x)
		}
		return c
	default:
		return v
	}
}

func cowGuardRun(t *testing.T, query string, data any) {
	t.Helper()
	q, err := gojq.Parse(query)
	if err != nil {
		t.Fatalf("parse %q: %v", query, err)
	}
	code, err := gojq.Compile(q)
	if err != nil {
		t.Fatalf("compile %q: %v", query, err)
	}
	iter := code.RunWithContext(context.Background(), data)
	for {
		v, more := iter.Next()
		if !more {
			break
		}
		if rerr, ok := v.(error); ok {
			t.Fatalf("runtime %q: %v", query, rerr)
		}
		_, _ = json.Marshal(v) // read-only consumption
	}
}

// cowGuardQueries are the destructive ops that reach deleteEmpty (del/
// delpaths) — the writers the patch must keep off the input.
var cowGuardQueries = []string{
	`del(.a)`,
	`del(.b.c)`,
	`del(.b.c.d)`,
	`delpaths([["a"],["b","c","d"]])`,
	`del(.list[1])`,
	`del(.b.c.d, .b.g.h)`,
}

// TestDeleteEmpty_CoW_NoSharedInputWrite is the fork-internal CI guard.
// MUST be run with -race. N goroutines run destructive queries over ONE
// SHARED input tree concurrently. With the allocator-aware-deleteEmpty
// patch the input is never written → -race clean + input byte-identical.
// Without the patch, concurrent in-place writes to the shared aliased
// sub-trees trip the race detector (and the fatal concurrent-map-write).
func TestDeleteEmpty_CoW_NoSharedInputWrite(t *testing.T) {
	shared := cowGuardTree()
	snapshot := cowGuardDeepCopy(shared)

	const goroutines = 24
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 40; i++ {
				cowGuardRun(t, cowGuardQueries[(g+i)%len(cowGuardQueries)], shared)
			}
		}(g)
	}
	wg.Wait()

	// Post-condition: the shared input is byte-identical (the patch never
	// wrote it). This complements the -race detection above.
	if !reflect.DeepEqual(shared, snapshot) {
		t.Fatalf("PATCH MISSING: deleteEmpty mutated the shared input\n got=%#v\nwant=%#v",
			shared, snapshot)
	}
}

// TestDeleteEmpty_CoW_ResultEquivalence asserts the patch does not change
// jq semantics: del/delpaths over a shared input yields a result equal to
// the same op over a private copy (a single-threaded correctness check
// alongside the -race guard above).
func TestDeleteEmpty_CoW_ResultEquivalence(t *testing.T) {
	for _, q := range cowGuardQueries {
		t.Run(q, func(t *testing.T) {
			resOf := func(data any) any {
				query, err := gojq.Parse(q)
				if err != nil {
					t.Fatalf("parse: %v", err)
				}
				code, err := gojq.Compile(query)
				if err != nil {
					t.Fatalf("compile: %v", err)
				}
				iter := code.RunWithContext(context.Background(), data)
				v, _ := iter.Next()
				if rerr, ok := v.(error); ok {
					t.Fatalf("runtime: %v", rerr)
				}
				return v
			}
			shared := resOf(cowGuardTree())
			private := resOf(cowGuardDeepCopy(cowGuardTree()))
			if !reflect.DeepEqual(shared, private) {
				t.Errorf("%q result diverges shared-vs-private:\n shared=%#v\nprivate=%#v",
					q, shared, private)
			}
		})
	}
}
