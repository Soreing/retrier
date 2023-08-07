package retrier

import (
	"context"
	"fmt"
	"math"
	"time"
)

type Retrier struct {
	max    int                     // Max number of retries (-1 for infinite)
	delayf func(int) time.Duration // Returns delay duration for retry count
}

// Creates a new custom retrier structure
// Retries are delayed by the given function's values
func NewRetrier(
	max int,
	delayf func(int) time.Duration,
) *Retrier {
	return &Retrier{
		max:    max,
		delayf: delayf,
	}
}

// No Delay
func NoDelay() func(int) time.Duration {
	return func(retries int) time.Duration {
		return 0
	}
}

// Delays are a constant amount
func ConstantDelay(
	delay time.Duration,
) func(int) time.Duration {
	return func(retries int) time.Duration {
		millis := delay
		return time.Duration(millis) * time.Millisecond
	}
}

// Delay is calculated by (delay*retries)
func LinearDelay(
	step time.Duration,
) func(int) time.Duration {
	return func(retries int) time.Duration {
		return step + time.Duration(retries)*step
	}
}

// Delay is calculated by min((delay*retries), cap)
func CappedLinearDelay(
	step time.Duration,
	cap time.Duration,
) func(int) time.Duration {
	return func(retries int) time.Duration {
		delay := step + time.Duration(retries)*step
		if delay < cap {
			return delay
		} else {
			return cap
		}
	}
}

// Delay is calculated by coef*base^retries
func ExponentialDelay(
	coef int,
	base int,
) func(int) time.Duration {
	return func(retries int) time.Duration {
		millis := coef * int(math.Pow(float64(base), float64(retries)))
		return time.Duration(millis) * time.Millisecond
	}
}

// Delay is calculated by min(coef*base^retries, cap)
func CappedExponentialDelay(
	coef int,
	base int,
	cap int,
) func(int) time.Duration {
	return func(retries int) time.Duration {
		raw := coef * int(math.Pow(float64(base), float64(retries)))
		millis := int(math.Min(float64(raw), float64(cap)))
		return time.Duration(millis) * time.Millisecond
	}
}

// Runs a task in the retrier with background context
func (r *Retrier) Run(work func() (error, bool)) error {
	return r.RunCtx(
		context.Background(),
		func(ctx context.Context) (error, bool) {
			return work()
		},
	)
}

// Runs a task in the retrier with custom context
func (r *Retrier) RunCtx(
	ctx context.Context,
	work func(ctx context.Context) (error, bool),
) error {
	retries := 0

	for {
		err, ret := work(ctx)
		if !ret {
			return err
		} else if r.max != -1 && retries >= r.max {
			return fmt.Errorf("failed after max retries: %w", err)
		} else {
			err := r.sleep(ctx, r.delayf(retries))
			if err != nil {
				return err
			}
			retries++
		}
	}
}

// Sleeps till the timer is up or the context is cancelled
func (r *Retrier) sleep(
	ctx context.Context,
	dur time.Duration,
) error {
	t := time.After(dur)
	select {
	case <-t:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
