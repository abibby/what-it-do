package parallel

import (
	"context"
	"iter"
	"sync"
)

func Seq[V any](ctx context.Context, s iter.Seq[V], cb func(ctx context.Context, v V) error) error {
	wg := &sync.WaitGroup{}

	ctx, cancel := context.WithCancelCause(ctx)
	for v := range s {
		wg.Add(1)
		go func(v V) {
			defer wg.Done()

			err := cb(ctx, v)
			if err != nil {
				cancel(err)
				return
			}
		}(v)
	}

	wg.Wait()

	return ctx.Err()
}

func FlatMap[V, U any](ctx context.Context, s iter.Seq[V], mapper func(ctx context.Context, s V) (iter.Seq[U], error)) iter.Seq[U] {
	return func(yield func(U) bool) {
		mtx := &sync.Mutex{}
		wg := &sync.WaitGroup{}

		ctx, cancel := context.WithCancelCause(ctx)
		for v := range s {
			wg.Add(1)
			go func(v V) {
				defer wg.Done()

				seq, err := mapper(ctx, v)
				if err != nil {
					cancel(err)
					return
				}

				mtx.Lock()
				defer mtx.Unlock()

				for v := range seq {
					yield(v)
				}
			}(v)
		}
		wg.Wait()
	}
}
