[![Go Report Card](https://goreportcard.com/badge/github.com/jonbodner/stackerr)](https://goreportcard.com/report/github.com/jonbodner/stackerr)

# stackerr

A simple Go error library that provides stack traces. 

## Motivation

One of the biggest differences between errors in Go and exceptions in other languages
is that you don't get a stack trace with a Go error. The `stackerr` package fixes this
limitation.

## Creating an error with a stack trace

There are three different functions in the `stackerr` package for creating a stack trace:

### Wrap

Use the `stackerr.Wrap` function to wrap errors returned by third-party libraries. It takes an existing error and wraps it in an error that contains a stack trace.  In order to simplify your error handling code, `Wrap` has two special cases:

- If a `nil` error is passed to `stackerr.Wrap`, `nil` is returned. 
- If an error with a stack trace error somewhere in its unwrap chain is passed to
`stackerr.Wrap`, it returns the passed-in error. 

These two rules make it possible to write the following code and 
not worry if there's already a stack trace (or no error) stored in `err`:

```go
func DoSomething(input string) (string, error) {
    result, err := ThingToCall(input)
    return result, stackerr.Wrap(err)
}
```

### Errorf

If you want to wrap an existing error with your own contextual information, use 
`stackerr.Errorf`. This works exactly like `fmt.Errorf`:

```go
func DoSomething(input string) (string, error) {
    result, err := ThingToCall(input)
    if err != nil {
        err = stackerr.Errorf("DoSomething failed on call to ThingToCall: %w", err)
    }
    return result, err
}
```

If there's an error in the unwrap chain that provides a stack trace, 
`stackerr.Errorf` preserves the existing trace information.

### New

If you are creating a new error that's only a `string`, use `stackerr.New`. Just as `stackerr.Errorf` is
 a replacement for `fmt.Errorf`, this function is a replacement for `errors.New`:

```go
func DoSomething(input string) (string, error) {
    if input == "" {
        return "", stackerr.New("cannot supply an empty string to DoSomething")
    }
    result, err := ThingToCall(input)
    return result, stackerr.Wrap(err)
}
```

## Retrieving the stack trace

Once you have an error in your unwrap chain with a stack trace, there are two ways to get the trace back.

### Trace

Use the `stackerr.Trace` function to get a `[]string` that contains each line of
the stack trace:

```go
s := stackerr.New("This is a stack trace error")
callStack, err := stackerr.Trace(s, stackerr.StandardFormat)
fmt.Println(callStack)
```

`stackerr.Trace` takes two parameters. The first is an error and the second is a
`text.Template`. There's a default template defined, `stackerr.StandardFormat`.
For each line, it produces output that looks like:

```txt
FUNCTION_NAME (FILE_PATH_AND_NAME:LINE_NUMBER)
```

If you want to write your own template, there are three valid variables:

- .Function (for the function name),
- .File (for the file path and name)
- .Line (for the line number).

There are three possible outputs from `stackerr.Trace`:

- If you supply an error that doesn't have stack trace in its unwrap chain, `nil` is returned for both the slice of strings and the error. 
- If an invalid template is supplied, `nil` is returned for the slice and the error is returned. (with a stack trace!)
- Otherwise, the stack trace is returned as a slice of strings along with a `nil` error.

Note that by default, the File path will include the absolute path to the file on the
machine that built the code. If you want to hide this path, build using the
`-trimpath` flag.

### fmt Formatting and %+v

Use the `%+v` formatting directive with `fmt.Printf` and variants to get the stack trace as a string. 

```go
s := stackerr.New("This is a stack trace error")
fmt.Printf("%+v\n",s)
```

This prints the stack trace using the `stackerr.StandardFormat`, with each level of the call stack separated by newlines (`\n`).

Note that this will not print out the stack trace if there is a `fmt.Errorf` wrapping the error with a stack trace. In those situations, you need to use `stackerr.Trace`.

## HasStack

Use `stackerr.HasStack` to determine if there is a stack trace in the unwrap chain for an error.

# Testing

The tests for `stackerr` require you to run `go test` with the `-trimpath` flag:

```bash
go test -trimpath ./...
``` 
