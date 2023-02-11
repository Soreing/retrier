# Retrier
Retrier is a small package that makes retrying anything easier with custom or predefined retry functions.

## Usage
Create a retrier by providing an upper limit to retries and a delay function.
```golang
ret := retrier.NewRetrier(
    10, // -1 for no limit
    retrier.ConstantDelay(time.Second),
)
```
You can use one of the predefined delay functions, or provide your own.
```golang
ret := retrier.NewRetrier(
    -1,
    func(count int) time.Duration {
        m := rand.Intn(60)
        return time.Second * time.Duration(m)
    },
)
```
Use the Run or RunCtx functions to run any task. The retrier will automatically retry the task if it returns an error until the retry cap is reached or the context is canceled.
```golang
ret.RunCtx(
    context.TODO(),
    func(ctx context.Context) error {
        resp, err := http.Get("https://some-api.com")
        if err != nil {
            fmt.Println("request failed")
            return err
        } else {
            fmt.Println(resp)
            return nil
        }
    },
)
```
## Delay Functions
| Function | Delay | Example |
|----------|-------|---------|
| No Delay                 | `0`               | 0, 0, 0, 0, 0   |
| Constant Delay           | `c`               | 1, 1, 1, 1, 1   |
| Linear Delay             | `r*c`             | 1, 2, 3, 4, 5   |
| Capped Linear Delay      | `min(r*c, cap)`   | 1, 2, 3, 3, 3   |
| Exponential Delay        | `a*b^r`           | 2, 4, 8, 16, 32 |
| Capped Exponential Delay | `min(a*b^r, cap)` | 2, 4, 8, 10, 10 |