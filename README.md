# Async
Async is an asynchronous task package for Go.

![License](https://img.shields.io/badge/license-MIT-green.svg)
[![Tests](https://github.com/yaitoo/async/actions/workflows/tests.yml/badge.svg)](https://github.com/yaitoo/async/actions/workflows/tests.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/yaitoo/async.svg)](https://pkg.go.dev/github.com/yaitoo/async)
[![Codecov](https://codecov.io/gh/yaitoo/async/branch/main/graph/badge.svg)](https://codecov.io/gh/yaitoo/async)
[![GitHub Release](https://img.shields.io/github/v/release/yaitoo/async)](https://github.com/yaitoo/sqle/blob/main/CHANGELOG.md)
[![Go Report Card](https://goreportcard.com/badge/yaitoo/async)](http://goreportcard.com/report/yaitoo/async)


## Features
- Wait/WaitAny/WaitN
- `context.Context` with `timeout`, `cancel`  support
- Works with generic instead of `interface{}`

## Tutorials
see more examples : [tests](./awaiter_test.go) or [play.go.dev](https://go.dev/play/p/IJ-lbIhTEQS)

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
wait all tasks to completed.

```
t := async.New[int](func(ctx context.Context) (int, error) {
		return 1, nil
	}, func(ctx context.Context) (int, error) {
		return 2, nil
	})

result, err, taskErrs := t.Wait(context.Background())


fmt.Println(result)  //[1,2] or [2,1]
fmt.Println(err) // nil
fmt.Println(taskErrs) //nil


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

result, err, tasksErr := t.WaitAny(context.Background())

fmt.Println(result)  //2
fmt.Println(err) //nil
fmt.Println(taskErrs) //nil

```

### WaitN
wait N tasks to completed. 

```
t := async.New[int](func(ctx context.Context) (int, error) {
    time.Sleep(2 * time.Second)
		return 1, nil
	}, func(ctx context.Context) (int, error) {
		return 2, nil
	}, func(ctx context.Context) (int, error) {
		return 3, nil
	})

result, err,taskErrs := t.WaitN(context.Background(),2)


fmt.Println(result)  //[2,3] or [3,2]
fmt.Println(err) //nil
fmt.Println(taskErrs) //nil

```

### Timeout
cancel all tasks if it is timeout. [playground](https://go.dev/play/p/AY42qZQPQAI)
```
 t := async.New[int](func(ctx context.Context) (int, error) {
		time.Sleep(2 * time.Second)
		return 1, nil
	}, func(ctx context.Context) (int, error) {
		time.Sleep(2 * time.Second)
		return 2, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result, err, tasks := t.WaitAny(ctx)
	//result, err, tasks := t.Wait(ctx)

	
	fmt.Println(result) //nil
	fmt.Println(err) // context.DeadlineExceeded
	fmt.Println(taskErrs) //nil
```

### Cancel
manually cancel all tasks.

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

//result, err, taskErrs := t.WaitAny(ctx)
 result, err, taskErrs := t.Wait(ctx)


fmt.Println(result)  //nil
fmt.Println(err) // context.Cancelled
fmt.Println(taskErrs) // nil


```


## Contributing
Contributions are welcome! If you're interested in contributing, please feel free to [contribute](CONTRIBUTING.md)


## License
[MIT License](LICENSE)