package stackerr_test

import (
	"errors"
	"fmt"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"

	"github.com/jonbodner/stackerr"
)

func TestWithStack(t *testing.T) {
	e := errors.New("new err")
	data := []struct {
		name         string
		formatString string
		expected     string
	}{
		{
			name:         "string",
			formatString: "%s",
			expected:     "new err",
		},
		{
			name:         "quoted",
			formatString: "%q",
			expected:     `"new err"`,
		},
		{
			name:         "value",
			formatString: "%v",
			expected:     "new err",
		},
		{
			name:         "detailed value",
			formatString: "%+v",
			expected: `new err
github.com/jonbodner/stackerr_test.TestWithStack (github.com/jonbodner/stackerr_test/stackerr_test.go:45)
testing.tRunner (testing/testing.go:909)
runtime.goexit (runtime/asm_amd64.s:1357)`,
		},
	}
	se := stackerr.Wrap(e)
	for _, v := range data {
		t.Run(v.name, func(t *testing.T) {
			result := fmt.Sprintf(v.formatString, se)
			if result != v.expected {
				t.Errorf("Expected `%s`, got `%s`", v.expected, result)
			}
		})
	}
	expectedTrace := `["github.com/jonbodner/stackerr_test.TestWithStack (github.com/jonbodner/stackerr_test/stackerr_test.go:45)" "testing.tRunner (testing/testing.go:909)" "runtime.goexit (runtime/asm_amd64.s:1357)"]`
	out, err := stackerr.Trace(se, stackerr.StandardFormat)
	if err != nil {
		t.Fatal(err)
	}
	actualTrace := fmt.Sprintf("%q", out)
	if expectedTrace != actualTrace {
		t.Errorf("Expected `%s`, got `%s`", expectedTrace, actualTrace)
	}

	// re-wrap does nothing
	se2 := stackerr.Wrap(se)
	for _, v := range data {
		t.Run(v.name, func(t *testing.T) {
			result := fmt.Sprintf(v.formatString, se2)
			if result != v.expected {
				t.Errorf("Expected `%s`, got `%s`", v.expected, result)
			}
		})
	}

	if se2.Error() != "new err" {
		t.Errorf("Expected `%s`, got `%s`", "new err", se2.Error())
	}
}

func TestNew(t *testing.T) {
	err := stackerr.New("test message")
	expected := `test message
github.com/jonbodner/stackerr_test.TestNew (github.com/jonbodner/stackerr_test/stackerr_test.go:81)
testing.tRunner (testing/testing.go:909)
runtime.goexit (runtime/asm_amd64.s:1357)`
	result := fmt.Sprintf("%+v", err)
	if expected != result {
		t.Errorf("expected `%s`, got `%s`", expected, result)
	}
}

func TestErrorf(t *testing.T) {
	data := []struct {
		name         string
		formatString string
		values       []interface{}
		expected     string
	}{
		{
			"wrapped non-stack trace error",
			"This is an %s: %w",
			[]interface{}{"error", errors.New("inner error")},
			`This is an error: inner error
github.com/jonbodner/stackerr_test.TestErrorf.func1 (github.com/jonbodner/stackerr_test/stackerr_test.go:138)
testing.tRunner (testing/testing.go:909)
runtime.goexit (runtime/asm_amd64.s:1357)`,
		},
		{
			"wrapped stack trace error",
			"This is an %s: %w",
			[]interface{}{"error", stackerr.New("inner error")},
			`This is an error: inner error
github.com/jonbodner/stackerr_test.TestErrorf (github.com/jonbodner/stackerr_test/stackerr_test.go:111)
testing.tRunner (testing/testing.go:909)
runtime.goexit (runtime/asm_amd64.s:1357)`,
		},
		{
			"double wrapped stack trace error",
			"This is an %s: %w",
			[]interface{}{"error", stackerr.Errorf("double-wrapped: %w", stackerr.New("inner error"))},
			`This is an error: double-wrapped: inner error
github.com/jonbodner/stackerr_test.TestErrorf (github.com/jonbodner/stackerr_test/stackerr_test.go:120)
testing.tRunner (testing/testing.go:909)
runtime.goexit (runtime/asm_amd64.s:1357)`,
		},
		{
			"no error",
			"This is an %s",
			[]interface{}{"error"},
			`This is an error
github.com/jonbodner/stackerr_test.TestErrorf.func1 (github.com/jonbodner/stackerr_test/stackerr_test.go:138)
testing.tRunner (testing/testing.go:909)
runtime.goexit (runtime/asm_amd64.s:1357)`,
		},
	}
	for _, v := range data {
		t.Run(v.name, func(t *testing.T) {
			errOuter := stackerr.Errorf(v.formatString, v.values...)
			result := fmt.Sprintf("%+v", errOuter)
			if v.expected != result {
				t.Errorf("expected `%s`, got `%s`", v.expected, result)
			}
		})
	}
}

func TestTrace(t *testing.T) {
	data := []struct {
		name     string
		inErr    error
		expected []string
	}{
		{
			"no trace",
			errors.New("error"),
			nil,
		},
		{
			"trace",
			stackerr.New("error"),
			[]string{
				"github.com/jonbodner/stackerr_test.TestTrace (github.com/jonbodner/stackerr_test/stackerr_test.go:160)",
				"testing.tRunner (testing/testing.go:909)",
				"runtime.goexit (runtime/asm_amd64.s:1357)",
			},
		},
		{
			"wrapped trace",
			fmt.Errorf("outer: %w", stackerr.New("inner")),
			[]string{
				"github.com/jonbodner/stackerr_test.TestTrace (github.com/jonbodner/stackerr_test/stackerr_test.go:169)",
				"testing.tRunner (testing/testing.go:909)",
				"runtime.goexit (runtime/asm_amd64.s:1357)",
			},
		},
	}
	for _, v := range data {
		t.Run(v.name, func(t *testing.T) {
			lines, err := stackerr.Trace(v.inErr, stackerr.StandardFormat)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(v.expected, lines); diff != "" {
				t.Error(diff)
			}
		})
	}

	// invalid format
	invalidFormat := template.Must(template.New("standardFormat").Parse("{{.Function}} ({{.File}}:{{.Foobar}})"))
	x := stackerr.New("bad")
	lines, err := stackerr.Trace(x, invalidFormat)
	if len(lines) != 0 {
		t.Errorf("Expected no lines ,got `%q`", lines)
	}
	expectedErr := `template: standardFormat:1:27: executing "standardFormat" at <.Foobar>: can't evaluate field Foobar in type runtime.Frame`
	var resultErr string
	if err != nil {
		resultErr = err.Error()
	}
	if expectedErr != resultErr {
		t.Errorf("expected `%s`, got `%s`", expectedErr, resultErr)
	}
}

func TestErrorStack_Is(t *testing.T) {
	err := stackerr.New("foo")
	if !errors.Is(err, err) {
		t.Error("oops")
	}
	err2 := errors.New("bar")
	if errors.Is(err, err2) {
		t.Error("shouldn't match")
	}
}

func TestErrorPrinting(t *testing.T) {
	err := stackerr.New("error message")
	err2 := stackerr.Errorf("wrapped %w", err)
	data := []struct {
		name     string
		err      error
		format   string
		expected string
	}{
		{
			name:     "v",
			err:      err,
			format:   "%v",
			expected: `error message`,
		},
		{
			name:   "plus_v",
			err:    err,
			format: "%+v",
			expected: `error message
github.com/jonbodner/stackerr_test.TestErrorPrinting (github.com/jonbodner/stackerr_test/stackerr_test.go:218)
testing.tRunner (testing/testing.go:909)
runtime.goexit (runtime/asm_amd64.s:1357)`,
		},
		{
			name:     "s",
			err:      err,
			format:   "%s",
			expected: `error message`,
		},
		{
			name:     "q",
			err:      err,
			format:   "%q",
			expected: `"error message"`,
		},
		{
			name:     "proxy_v",
			err:      err2,
			format:   "%v",
			expected: `wrapped error message`,
		},
		{
			name:   "proxy_plus_v",
			err:    err2,
			format: "%+v",
			expected: `wrapped error message
github.com/jonbodner/stackerr_test.TestErrorPrinting (github.com/jonbodner/stackerr_test/stackerr_test.go:218)
testing.tRunner (testing/testing.go:909)
runtime.goexit (runtime/asm_amd64.s:1357)`,
		},
		{
			name:     "proxy_s",
			err:      err2,
			format:   "%s",
			expected: `wrapped error message`,
		},
		{
			name:     "proxy_q",
			err:      err2,
			format:   "%q",
			expected: `"wrapped error message"`,
		},
	}
	for _, v := range data {
		t.Run(v.name, func(t *testing.T) {
			result := fmt.Sprintf(v.format, v.err)
			if result != v.expected {
				t.Errorf("Expected `%s`, got `%s`", v.expected, result)
			}
		})
	}
}

func TestWithStackNil(t *testing.T) {
	if stackerr.Wrap(nil) != nil {
		t.Error("Got non-nil for nil passed to Wrap")
	}
}

func TestHasStack(t *testing.T) {
	e := errors.New("innermost error")
	s := stackerr.Wrap(e)
	f := fmt.Errorf("wrapped: %w", s)

	if stackerr.HasStack(e) {
		t.Error("e doesn't have a stack trace")
	}
	if !stackerr.HasStack(s) {
		t.Error("s does have a stack trace")
	}
	if !stackerr.HasStack(f) {
		t.Error("f does have a stack trace")
	}
}
