package retrier

import (
	"context"
	"fmt"
	"math"
	"time"
)

// Retrier controls how to to run the retry function. A task will be retried
// up to a set retry count with some delay between the retries defined by a
// delay function.
type Retrier struct {
	// max is the upper limit of retries. The task can not be retried more than
	// the specified number. To disable the limit, set -1 as the value.
	max int

	// delayf returns some amount of duration to wait before retrying a task.
	// The function takes the retry count as a parameter to allow for increasing
	// delay between retries.
	delayf func(int) time.Duration
}

// NewRetrier creates a retrier from max retries and a delay function.
func NewRetrier(
	max int,
	delayf func(int) time.Duration,
) *Retrier {
	return &Retrier{
		max:    max,
		delayf: delayf,
	}
}

// NoDelay returns a delay function that has no delay between retries.
func NoDelay() func(int) time.Duration {
	return func(retries int) time.Duration {
		return 0
	}
}

// ConstantDelay returns a delay function that creates a constant wait duration
// between retries. The delay will be the same between the retries.
func ConstantDelay(
	delay time.Duration,
) func(int) time.Duration {
	return func(retries int) time.Duration {
		millis := delay
		return time.Duration(millis) * time.Millisecond
	}
}

// LinearDelay returns a delay function that creates a linearly increasing
// wait duration between retries. The delay is calculated by (step*retries).
func LinearDelay(
	step time.Duration,
) func(int) time.Duration {
	return func(retries int) time.Duration {
		return step + time.Duration(retries)*step
	}
}

// CappedLinearDelay returns a delay function that creates a linearly increasing
// wait duration between retries up to a specific limit where delay can not be
// longer. The delay is calculated by min((delay*retries), cap)
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

// ExponentialDelay returns a delay function that creates an exponentially
// increasing wait duration between retries. The delay is calculated by
// (coef*base^retries).
func ExponentialDelay(
	coef int,
	base int,
) func(int) time.Duration {
	return func(retries int) time.Duration {
		millis := coef * int(math.Pow(float64(base), float64(retries)))
		return time.Duration(millis) * time.Millisecond
	}
}

// ExponentialDelay returns a delay function that creates an exponentially
// increasing wait duration between retries up to a specific limit where delay
// can not be longer.. The delay is calculated by (coef*base^retries).
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

// Run executes a work task with the background context.
func (r *Retrier) Run(work func() (error, bool)) error {
	return r.RunCtx(
		context.Background(),
		func(ctx context.Context) (error, bool) {
			return work()
		},
	)
}

// RunCtx executes a work task in the context of a retrier until the task
// decides not to retry, or if the maximum retries have been reached, or if the
// context has been canceled and retrying should stop.
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

// sleep stops the execution for some duration, or until the context has
// been canceled.
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
