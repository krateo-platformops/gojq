// Command shared-input-cow demonstrates the single behavioral delta of this
// fork: del/delpaths are copy-on-write over any node gojq does not own, so a
// SHARED input can be filtered concurrently without a per-call deep copy.
//
// Run it from the repo checkout with the race detector — that is the point:
//
//	go run -race ./examples/shared-input-cow
//
// On this fork it is race-clean and the shared input survives byte-identical.
// On upstream gojq (v0.12.19) the same program is a data race: upstream's
// deleteEmpty recurses into sub-trees still aliased to the caller's input and
// writes them in place. (A single-threaded before/after compare cannot see
// that write — it usually stores the same value back — which is why this
// example, like the fork's delete_empty_cow_test.go guard, is concurrent.)
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"

	// The replace-key import path: consumers know this fork as the upstream
	// module (replace github.com/itchyny/gojq => github.com/krateo-platformops/gojq vX.Y.Z);
	// inside this repo the in-fork `replace … => ./` resolves it to the checkout.
	"github.com/itchyny/gojq"
)

func sharedInput() map[string]any {
	// Deleted paths have SIBLINGS at every level — the shape that makes
	// upstream's deleteEmpty walk into aliased input sub-trees.
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

func deepCopy(v any) any {
	switch t := v.(type) {
	case map[string]any:
		c := make(map[string]any, len(t))
		for k, w := range t {
			c[k] = deepCopy(w)
		}
		return c
	case []any:
		c := make([]any, len(t))
		for i, w := range t {
			c[i] = deepCopy(w)
		}
		return c
	default:
		return v
	}
}

func main() {
	query, err := gojq.Parse(`delpaths([["b","c","d"], ["list",0]])`)
	if err != nil {
		log.Fatal(err)
	}
	code, err := gojq.Compile(query) // compiled once, reused read-only by all goroutines
	if err != nil {
		log.Fatal(err)
	}

	input := sharedInput()      // ONE value, shared by every goroutine below
	snapshot := deepCopy(input) // private reference copy for the final compare

	const writers, readers, iters = 8, 4, 200
	var wg sync.WaitGroup

	for w := 0; w < writers; w++ { // concurrent del/delpaths over the SHARED input
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				iter := code.Run(input)
				for {
					v, ok := iter.Next()
					if !ok {
						break
					}
					if err, isErr := v.(error); isErr {
						log.Fatalf("query error: %v", err)
					}
				}
			}
		}()
	}
	for r := 0; r < readers; r++ { // concurrent readers walking the same input
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				if _, err := json.Marshal(input); err != nil {
					log.Fatalf("reader: %v", err)
				}
			}
		}()
	}
	wg.Wait()

	if !reflect.DeepEqual(input, snapshot) {
		log.Fatal("shared input was mutated — the CoW patch is not in effect")
	}
	out, _ := json.Marshal(input)
	fmt.Printf("ok: %d concurrent delpaths runs, shared input untouched: %s\n",
		writers*iters, out)
}
