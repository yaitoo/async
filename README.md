# Async
Async is an asynchronous task package for Go.

![License](https://img.shields.io/badge/license-MIT-green.svg)
[![Tests](https://github.com/yaitoo/async/actions/workflows/tests.yml/badge.svg)](https://github.com/yaitoo/async/actions/workflows/tests.yml)
[![GoDoc](https://godoc.org/github.com/yaitoo/async?status.png)](https://godoc.org/github.com/yaitoo/async)
[![Codecov](https://codecov.io/gh/yaitoo/async/branch/main/graph/badge.svg)](https://codecov.io/gh/yaitoo/async)
[![GitHub Release](https://img.shields.io/github/v/release/yaitoo/async)](https://github.com/yaitoo/sqle/blob/main/CHANGELOG.md)
[![Go Report Card](https://goreportcard.com/badge/yaitoo/async)](http://goreportcard.com/report/yaitoo/async)


## Features
- Wait/WaitAny
- `context.Context` with `timeout`, `cancel`  support
- Works with generic instead of `interface{}`

## Tutorials
see more [examples](./awaiter_test.go)

### Install async
- install latest commit from `main` branch
```
go get github.com/yaitoo/async@main
```

- install latest release
```
go get github.com/yaitoo/async@latest
```

### Wait 
wait all tasks to completed

```
t := async.New[int](func(ctx context.Context) (int, error) {
		return 1, nil
	}, func(ctx context.Context) (int, error) {
		return 2, nil
	})

result, err := t.Wait(context.Background())

if err == nil {
  fmt.Println(result)  //[1,2]
}else{
  fmt.Println(err)
}

```


### WaitAny
wait any task to completed

```
t := async.New[int](func(ctx context.Context) (int, error) {
    time.Sleep(2 * time.Second)
		return 1, nil
	}, func(ctx context.Context) (int, error) {
		return 2, nil
	})

result, err := t.WaitAny(context.Background())

if err == nil {
  fmt.Println(result)  //2
}else{
  fmt.Println(err)
}

```

### Timeout
cancel all tasks if it is timeout
```
t := async.New[int](func(ctx context.Context) (int, error) {
    time.Sleep(2 * time.Second)
		return 1, nil
	}, func(ctx context.Context) (int, error) {
     time.Sleep(2 * time.Second)
		return 2, nil
	})

result, err := t.WaitAny(context.Timeout(context.Background(), 1 * time.Second))
//result, err := t.Wait(context.Timeout(context.Background(), 1 * time.Second))

if err == nil {
  fmt.Println(result)  
}else{
  fmt.Println(err) // context.DeadlineExceeded
}
```

### Cancel
manually cancel all tasks

```
t := async.New[int](func(ctx context.Context) (int, error) {
    time.Sleep(2 * time.Second)
		return 1, nil
	}, func(ctx context.Context) (int, error) {
     time.Sleep(2 * time.Second)
		return 2, nil
	})

ctx, cancel := context.WithCancel(context.Background())
go func(){
  time.Sleep(1 * time.Second)
  cancel()
}()

result, err := t.WaitAny(ctx)
// result, err := t.Wait(ctx)

if err == nil {
  fmt.Println(result)  
}else{
  fmt.Println(err) // context.Cancelled
}

```


## Contributing
Contributions are welcome! If you're interested in contributing, please feel free to [contribute](CONTRIBUTING.md)


## License
[MIT License](LICENSE)