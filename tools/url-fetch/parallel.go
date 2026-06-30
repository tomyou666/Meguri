package main

import (
	"context"
	"sync"
)

// runParallel は fn を各 index に対して並列実行し、index 順に結果を返す。
func runParallel(ctx context.Context, n int, fn func(context.Context, int)) {
	if n == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			fn(ctx, i)
		}(i)
	}
	wg.Wait()
}
