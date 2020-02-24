package stackerr

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"text/template"
)

// errorStack wraps an error with the stack location where the error occurred.
type errorStack struct {
	Err     error
	trace   []uintptr
	earlier *errorStack
}

// StackTrace returns the call stack frames for the errorStack. If this was the first errorStack on
// the unwrap chain, it captures when the errorStack was instantiated. If there was an earlier errorStack in the chain,
// the se.earlier field is set, and the StackTrace() is returned from it.
//
//  Since *runtime.Frames tracks its own offset and cannot be reused, StackTrace creates a new instance of
// *runtime.Frames every time this method runs.
func (e errorStack) StackTrace() *runtime.Frames {
	if e.earlier != nil {
		return e.earlier.StackTrace()
	}
	return runtime.CallersFrames(e.trace)
}

// Is provides an implementation of the Is method to support the errors.Is() function. This allows two errorStack
// instances to be compared to each other using errors.Is. Both errorStack instances need to be unwrapped because the
// trace field and the earlier field are not relevant for the comparison.
func (e errorStack) Is(err error) bool {
	if err, ok := err.(errorStack); ok {
		return errors.Is(e.Err, err.Err)
	}
	return false
}

// Wrap takes in an error and returns an error wrapped in a errorStack with the location where
// the error was first created or returned from third-party code. If there is already an errorStack
// in the error chain, Wrap returns the passed-in error. Wrap returns nil when a nil error is passed in.
func Wrap(err error) error {
	if err == nil {
		return nil
	}
	var se errorStack
	if errors.As(err, &se) {
		return err
	}
	return errorStack{
		Err:   err,
		trace: buildStackTrace(),
	}
}

func buildStackTrace() []uintptr {
	pc := make([]uintptr, 20)
	n := runtime.Callers(3, pc)
	pc = pc[:n]
	return pc
}

// New builds a errorStack out of a string
func New(msg string) error {
	return errorStack{
		Err:   errors.New(msg),
		trace: buildStackTrace(),
	}
}

// Errorf wraps the error returned by fmt.Errorf in an errorStack. If there is an existing errorStack
// in the unwrap chain, its stack trace is used.
func Errorf(format string, vals ...interface{}) error {
	err := fmt.Errorf(format, vals...)
	out := errorStack{
		Err: err,
	}
	// it's possible that there was already an errorStack in the unwrap chain of the error returned
	// by fmt.Errorf. If so, set the earlier field in the new errorStack to refer to it. Otherwise,
	// create a new stack trace.
	var st errorStack
	if errors.As(err, &st) {
		if st.earlier != nil {
			out.earlier = st.earlier
		} else {
			out.earlier = &st
		}
	} else {
		out.trace = buildStackTrace()
	}
	return out
}

// Unwrap exposes the error wrapped by errorStack
func (e errorStack) Unwrap() error {
	return e.Err
}

// Error returns the error string for the wrapped error.
func (e errorStack) Error() string {
	return e.Err.Error()
}

// Format controls the optional display of the stack trace. Use %+v to output the stack trace, use %v or %s to output
// the wrapped error only, use %q to get a single-quoted character literal safely escaped with Go syntax for the wrapped
// error.
func (e errorStack) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", e.Unwrap())
			trace, _ := Trace(e, StandardFormat)
			fmt.Fprintf(s, "%s", strings.Join(trace, "\n"))
			return
		}
		io.WriteString(s, e.Error()) // nolint: errcheck
	case 's':
		io.WriteString(s, e.Error()) // nolint: errcheck
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}

// StandardFormat is the default template used to convert a *runtime.Frame to a string. Each entry is formatted as
// "FUNCTION_NAME (FILE_NAME:LINE_NUMBER)"
var StandardFormat = template.Must(template.New("standardFormat").Parse("{{.Function}} ({{.File}}:{{.Line}})"))

// Trace returns the stack trace information as a slice of strings formatted using the provided Go template. The valid
// fields in the template are Function, File, and Line. See StandardFormat for an example.
func Trace(e error, t *template.Template) ([]string, error) {
	var se errorStack
	if !errors.As(e, &se) {
		return nil, nil
	}
	s := make([]string, 0, 20)
	frames := se.StackTrace()
	var b bytes.Buffer
	for {
		b.Reset()
		frame, more := frames.Next()
		err := t.Execute(&b, frame)
		if err != nil {
			return nil, Wrap(err)
		}
		s = append(s, b.String())
		if !more {
			break
		}
	}
	return s, nil
}

// HasStack returns true if there is a stack trace in the unwrap chain for the error.
func HasStack(e error) bool {
	var se errorStack
	return errors.As(e, &se)
}
