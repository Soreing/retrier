package retrier

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewRetrier tests if creating a new retrier is successful and the object
// fields are initialized accurately
func TestNewRetrier(t *testing.T) {
	tests := []struct {
		Name  string
		Max   int
		Delay func(int) time.Duration
	}{
		{
			Name: "Create new retrier",
			Max:  5,
			Delay: func(int) time.Duration {
				return time.Second
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			retr := NewRetrier(test.Max, test.Delay)

			if assert.NotNil(t, retr) {
				assert.Equal(t, test.Max, retr.max)
				assert.NotNil(t, retr.delayf)
			}
		})
	}
}

// TestNoDelay tests if the no delay function returns 0 duration in all cases
func TestNoDelay(t *testing.T) {
	tests := []struct {
		Name  string
		Count int
		Delay time.Duration
	}{
		{
			Name:  "First call",
			Count: 0,
			Delay: 0,
		},
		{
			Name:  "Second call",
			Count: 1,
			Delay: 0,
		},
		{
			Name:  "Nth call",
			Count: 25,
			Delay: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			fn := NoDelay()
			dur := fn(test.Count)

			assert.Equal(t, test.Delay, dur)
		})
	}
}

// TestNoDelay tests if the constant delay function returns the same duration
// for each function call that it was initialized with
func TestConstantDelay(t *testing.T) {
	tests := []struct {
		Name  string
		Count int
		Delay time.Duration
	}{
		{
			Name:  "First call",
			Count: 0,
			Delay: time.Second,
		},
		{
			Name:  "Second call",
			Count: 1,
			Delay: time.Second,
		},
		{
			Name:  "Nth call",
			Count: 25,
			Delay: time.Second,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			fn := ConstantDelay(test.Delay)
			dur := fn(test.Count)

			assert.Equal(t, test.Delay, dur)
		})
	}
}

// TestLinearDelay tests if the linear delay function returns the delay it was
// initialized with on the first call, then it increases the delay with the same
// amount for each subsequent call
func TestLinearDelay(t *testing.T) {
	tests := []struct {
		Name     string
		Count    int
		DelayIn  time.Duration
		DelayOut time.Duration
	}{
		{
			Name:     "First call",
			Count:    0,
			DelayIn:  time.Second,
			DelayOut: time.Second,
		},
		{
			Name:     "Second call",
			Count:    1,
			DelayIn:  time.Second,
			DelayOut: time.Second + time.Second,
		},
		{
			Name:     "Nth call",
			Count:    25,
			DelayIn:  time.Second,
			DelayOut: time.Second + time.Second*25,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			fn := LinearDelay(test.DelayIn)
			dur := fn(test.Count)

			assert.Equal(t, test.DelayOut, dur)
		})
	}
}

// TestCappedLinearDelay tests if the capped linear delay function returns the
// delay it was initialized with on the first call, then it increases the delay
// with the same amount for each subsequent call until it reaches a limit,
// where the delay must be the specified limit for each subsequent call
func TestCappedLinearDelay(t *testing.T) {
	tests := []struct {
		Name     string
		Count    int
		DelayIn  time.Duration
		DelayCap time.Duration
		DelayOut time.Duration
	}{
		{
			Name:     "First call",
			Count:    0,
			DelayIn:  time.Second,
			DelayCap: time.Second * 10,
			DelayOut: time.Second,
		},
		{
			Name:     "Second call",
			Count:    1,
			DelayIn:  time.Second,
			DelayCap: time.Second * 10,
			DelayOut: time.Second + time.Second,
		},
		{
			Name:     "Nth call within limit",
			Count:    5,
			DelayIn:  time.Second,
			DelayCap: time.Second * 10,
			DelayOut: time.Second + time.Second*5,
		},
		{
			Name:     "Nth call outside of limit",
			Count:    25,
			DelayIn:  time.Second,
			DelayCap: time.Second * 10,
			DelayOut: time.Second * 10,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			fn := CappedLinearDelay(test.DelayIn, test.DelayCap)
			dur := fn(test.Count)

			assert.Equal(t, test.DelayOut, dur)
		})
	}
}

// TestExponentialDelay tests if the exponential delay function returns the
// delay it was initialized with on the first call, then increasing the delay
// by an exponent for each subsequent call
func TestExponentialDelay(t *testing.T) {
	tests := []struct {
		Name     string
		Count    int
		Base     int
		DelayIn  time.Duration
		DelayOut time.Duration
	}{
		{
			Name:     "First call",
			Count:    0,
			Base:     5,
			DelayIn:  time.Second,
			DelayOut: time.Second,
		},
		{
			Name:     "Second call",
			Count:    1,
			Base:     4,
			DelayIn:  time.Second,
			DelayOut: time.Second * 4,
		},
		{
			Name:     "Third call",
			Count:    2,
			Base:     3,
			DelayIn:  time.Second,
			DelayOut: time.Second * 3 * 3,
		},
		{
			Name:     "Nth call",
			Count:    25,
			Base:     2,
			DelayIn:  time.Second,
			DelayOut: time.Second * 33554432,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			fn := ExponentialDelay(test.DelayIn, test.Base)
			dur := fn(test.Count)

			assert.Equal(t, test.DelayOut, dur)
		})
	}
}

// TestCappedExponentialDelay tests if the capped exponential delay function returns
// the delay it was initialized with on the first call, then increasing the delay
// by an exponent for each subsequent call until it reaches a limit, where the
// delay must be the specified limit for each subsequent call
func TestCappedExponentialDelay(t *testing.T) {
	tests := []struct {
		Name     string
		Count    int
		Base     int
		DelayIn  time.Duration
		DelayCap time.Duration
		DelayOut time.Duration
	}{
		{
			Name:     "First call",
			Count:    0,
			Base:     5,
			DelayIn:  time.Second,
			DelayCap: time.Hour,
			DelayOut: time.Second,
		},
		{
			Name:     "Second call",
			Count:    1,
			Base:     4,
			DelayIn:  time.Second,
			DelayCap: time.Hour,
			DelayOut: time.Second * 4,
		},
		{
			Name:     "Third call",
			Count:    2,
			Base:     3,
			DelayIn:  time.Second,
			DelayCap: time.Hour,
			DelayOut: time.Second * 3 * 3,
		},
		{
			Name:     "Nth call within limit",
			Count:    5,
			Base:     2,
			DelayIn:  time.Second,
			DelayCap: time.Hour,
			DelayOut: time.Second * 32,
		},
		{
			Name:     "Nth call outside limit",
			Count:    25,
			Base:     2,
			DelayIn:  time.Second,
			DelayCap: time.Hour,
			DelayOut: time.Hour,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			fn := CappedExponentialDelay(
				test.DelayIn,
				test.Base,
				test.DelayCap,
			)

			dur := fn(test.Count)

			assert.Equal(t, test.DelayOut, dur)
		})
	}
}

// TestSleep tests if the sleep function can pause the execution for some
// duration or returns preemptively when the context is canceled
func TestSleep(t *testing.T) {
	tests := []struct {
		Name     string
		Duration time.Duration
		Timeout  time.Duration
		Elapsed  time.Duration
		Error    error
	}{
		{
			Name:     "Sleep for some duration",
			Duration: time.Millisecond * 2,
			Timeout:  time.Millisecond * 5,
			Elapsed:  time.Millisecond * 2,
			Error:    nil,
		},
		{
			Name:     "Context times out during sleep",
			Duration: time.Millisecond * 20,
			Timeout:  time.Millisecond * 5,
			Elapsed:  time.Millisecond * 5,
			Error:    context.DeadlineExceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx, cncl := context.WithTimeout(
				context.TODO(),
				test.Timeout,
			)
			defer cncl()

			var err error
			ch := make(chan bool)
			st := time.Now()
			go func() {
				err = sleep(ctx, test.Duration)
				ch <- true
			}()
			select {
			case <-time.After(time.Second):
				panic("test function hang")
			case <-ch:
			}
			dif := time.Since(st)

			if test.Error != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.Error.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.GreaterOrEqual(t, dif, test.Elapsed)
		})
	}
}

// TestRunCtx tests if a task can be ran by the retrier until it succeeds or
// fails using a provided context
func TestRunCtx(t *testing.T) {
	tests := []struct {
		Name    string
		Max     int
		Timeout time.Duration
		Delay   func(int) time.Duration
		Task    func(ctx context.Context) (error, bool)
		Elapsed time.Duration
		Error   error
	}{
		{
			Name:    "Task succeeds immediately",
			Max:     5,
			Timeout: time.Millisecond * 100,
			Delay:   ConstantDelay(time.Millisecond * 5),
			Task: func(ctx context.Context) (error, bool) {
				return nil, false
			},
			Elapsed: 0,
			Error:   nil,
		},
		{
			Name:    "Task succeeds after some tries",
			Max:     5,
			Timeout: time.Millisecond * 100,
			Delay:   ConstantDelay(time.Millisecond * 5),
			Task: func(ctx context.Context) (error, bool) {
				cnt := ctx.Value("count")
				if v, ok := cnt.(*int); !ok {
					return fmt.Errorf("invalid value"), false
				} else if *v < 3 {
					*v++
					return fmt.Errorf("count too small"), true
				} else {
					return nil, false
				}
			},
			Elapsed: time.Millisecond * 15,
			Error:   nil,
		},
		{
			Name:    "Task fatally fails",
			Max:     5,
			Timeout: time.Millisecond * 100,
			Delay:   ConstantDelay(time.Millisecond * 5),
			Task: func(ctx context.Context) (error, bool) {
				return fmt.Errorf("fatal error"), false
			},
			Elapsed: 0,
			Error:   fmt.Errorf("fatal error"),
		},
		{
			Name:    "Task fails after max retries",
			Max:     5,
			Timeout: time.Millisecond * 100,
			Delay:   ConstantDelay(time.Millisecond * 5),
			Task: func(ctx context.Context) (error, bool) {
				return fmt.Errorf("error"), true
			},
			Elapsed: time.Millisecond * 25,
			Error:   fmt.Errorf("failed after max retries: error"),
		},
		{
			Name:    "Context times out during retries",
			Max:     -1,
			Timeout: time.Millisecond * 100,
			Delay:   ConstantDelay(time.Millisecond * 5),
			Task: func(ctx context.Context) (error, bool) {
				return fmt.Errorf("error"), true
			},
			Elapsed: time.Millisecond * 100,
			Error:   context.DeadlineExceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			retr := NewRetrier(test.Max, test.Delay)
			ctx := context.WithValue(context.TODO(), "count", new(int))
			ctx, cncl := context.WithTimeout(ctx, test.Timeout)
			defer cncl()

			var err error
			ch := make(chan bool)
			st := time.Now()
			go func() {
				err = retr.RunCtx(ctx, test.Task)
				ch <- true
			}()
			select {
			case <-time.After(time.Second):
				panic("test function hang")
			case <-ch:
			}
			dif := time.Since(st)

			if test.Error != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.Error.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.GreaterOrEqual(t, dif, test.Elapsed)
		})
	}
}

// TestRun tests if a task can be ran by the retrier
func TestRun(t *testing.T) {
	tests := []struct {
		Name    string
		Max     int
		Delay   func(int) time.Duration
		Task    func() (error, bool)
		Elapsed time.Duration
		Error   error
	}{
		{
			Name:  "Task succeeds immediately",
			Max:   5,
			Delay: ConstantDelay(time.Millisecond * 5),
			Task: func() (error, bool) {
				return nil, false
			},
			Elapsed: 0,
			Error:   nil,
		},
		{
			Name:  "Task fails after max retries",
			Max:   5,
			Delay: ConstantDelay(time.Millisecond * 5),
			Task: func() (error, bool) {
				return fmt.Errorf("error"), true
			},
			Elapsed: time.Millisecond * 25,
			Error:   fmt.Errorf("failed after max retries: error"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			retr := NewRetrier(test.Max, test.Delay)

			var err error
			ch := make(chan bool)
			st := time.Now()
			go func() {
				err = retr.Run(test.Task)
				ch <- true
			}()
			select {
			case <-time.After(time.Second):
				panic("test function hang")
			case <-ch:
			}
			dif := time.Since(st)

			if test.Error != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.Error.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.GreaterOrEqual(t, dif, test.Elapsed)
		})
	}
}
