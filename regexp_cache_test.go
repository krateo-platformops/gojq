package gojq

// regexp_cache_test.go — the regexps test/match/capture/scan/split/sub/gsub
// compile are cached per RUN (on the env), not on the compiler. A compiled
// *Code is reused across runs (and goroutines); a compiler-level cache kept
// every distinct pattern any run ever built from its input, for as long as
// the program lived.
//
//	go test -run '^$' -bench RegexpCache -benchmem -count=6

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"testing"
)

func compileForTest(t testing.TB, q string, opts ...CompilerOption) *Code {
	t.Helper()
	pq, err := Parse(q)
	if err != nil {
		t.Fatal(err)
	}
	code, err := Compile(pq, opts...)
	if err != nil {
		t.Fatal(err)
	}
	return code
}

// drain collects a run's outputs and fails the test on a run error. It calls
// t.Fatal: call it on the test goroutine only (workers use collect).
func drain(t testing.TB, iter Iter) []any {
	t.Helper()
	out, err := collect(iter)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// collect is drain without the testing.TB: it returns the run's error, so a
// worker goroutine can report it with t.Errorf and return.
func collect(iter Iter) ([]any, error) {
	var out []any
	for {
		v, ok := iter.Next()
		if !ok {
			return out, nil
		}
		if err, isErr := v.(error); isErr {
			return out, err
		}
		out = append(out, v)
	}
}

func retainedHeap() uint64 {
	runtime.GC()
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.HeapAlloc
}

// TestRegexpCache_ProgramRetainsNoPatternAcrossRuns — one compiled program run
// over N distinct data-built patterns keeps no regexp once the runs end: the
// retained heap after GC does not grow with N. With the cache on the compiler
// it grew by ~1 KiB per pattern (50,000 patterns: tens of MiB) for as long as
// the program lived.
func TestRegexpCache_ProgramRetainsNoPatternAcrossRuns(t *testing.T) {
	code := compileForTest(t, `.prefix as $p | .name | test("^" + $p + "-[0-9]+$")`)
	run := func(from, n int) {
		for i := from; i < from+n; i++ {
			in := map[string]any{"name": fmt.Sprintf("p%06d-1", i), "prefix": fmt.Sprintf("p%06d", i)}
			if got := drain(t, code.Run(in)); !reflect.DeepEqual(got, []any{true}) {
				t.Fatalf("pattern %d: %v", i, got)
			}
		}
	}
	run(0, 100) // warm up everything that is not per pattern
	base := retainedHeap()
	const n = 20000
	run(100, n)
	grown := int64(retainedHeap()) - int64(base)
	// The bound is absolute, not derived from the cache: 20,000 compiled
	// regexps are megabytes; a flat heap moves by GC noise only.
	const bound = 512 << 10
	if grown > bound {
		t.Fatalf("retained heap grew %d KiB over %d distinct patterns (bound %d KiB): the program keeps its runs' regexps",
			grown>>10, n, bound>>10)
	}
	runtime.KeepAlive(code)
}

// TestRegexpCache_ReusedWithinARun — within one run, a pattern used for many
// items is compiled once: the run's cache holds one regexp per distinct
// (pattern, flags), however many items it tests.
func TestRegexpCache_ReusedWithinARun(t *testing.T) {
	code := compileForTest(t, `[.[] | select(test("^app-[0-9]+$")) | sub("app-"; "x-"; "g")]`)
	items := make([]any, 1000)
	for i := range items {
		items[i] = fmt.Sprintf("app-%d", i)
	}
	iter := code.Run(items)
	got := drain(t, iter)
	if len(got) != 1 || len(got[0].([]any)) != len(items) {
		t.Fatalf("unexpected output %v", got)
	}
	// test's pattern, and sub's (pattern, flags) pair: two regexps, not 2,000.
	if n := len(iter.(*env).regexps); n != 2 {
		t.Fatalf("the run compiled %d regexps for 1,000 items over 2 patterns", n)
	}
}

// TestRegexpCache_ConcurrentRunsShareNothing — one compiled program, many
// goroutines, each run testing its items against its own pattern ($p) and a
// pattern common to every run: every result correct, and (under -race) no
// regexp state shared between the runs.
func TestRegexpCache_ConcurrentRunsShareNothing(t *testing.T) {
	code := compileForTest(t, `[.[] | (test($p) and test("^item-"))]`, WithVariables([]string{"$p"}))
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for r := 0; r < 200; r++ {
				name := fmt.Sprintf("item-%d-%d", g, r)
				got, err := collect(code.Run([]any{name, "item-other", "other"}, "^"+name+"$"))
				if err != nil {
					t.Errorf("goroutine %d run %d: %v", g, r, err)
					return
				}
				if want := []any{[]any{true, false, false}}; !reflect.DeepEqual(got, want) {
					t.Errorf("goroutine %d run %d: %v, want %v", g, r, got, want)
					return
				}
			}
		}(g)
	}
	wg.Wait()
}

var regexpCacheSink []any

// BenchmarkRegexpCache_OnePatternManyItems — the cost of caching per run: one
// regexp reused for 10,000 items within each run (compiled once per run), and
// the same filter run on one item (where a compiler-level cache would have
// saved the run's only compile).
func BenchmarkRegexpCache_OnePatternManyItems(b *testing.B) {
	code := compileForTest(b, `[.[] | select(test("^app-[0-9]+$"))] | length`)
	for _, n := range []int{1, 10000} {
		items := make([]any, n)
		for i := range items {
			items[i] = fmt.Sprintf("app-%d", i)
		}
		b.Run(fmt.Sprintf("items=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				regexpCacheSink = drain(b, code.Run(items))
			}
		})
	}
}
