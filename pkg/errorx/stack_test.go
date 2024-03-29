package errorx

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// a version of runtime.Caller that returns a Frame, not a uintptr.
func caller() Frame {
	var pcs [3]uintptr
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	frame, _ := frames.Next()
	return Frame(frame.PC)
}

func stackTrace() StackTrace {
	const depth = 8
	var pcs [depth]uintptr
	n := runtime.Callers(1, pcs[:])
	var st stack = pcs[0:n]
	return st.StackTrace()
}

var initpc = caller()

type X struct{}

// val returns a Frame pointing to itself.
func (x X) val() Frame {
	return caller()
}

// ptr returns a Frame pointing to itself.
func (x *X) ptr() Frame {
	return caller()
}

func TestFrameFormat(t *testing.T) {
	var tests = []struct {
		Frame
		format string
		want   string
	}{{
		initpc,
		"%s",
		"stack_test.go",
	}, {
		0,
		"%s",
		"unknown",
	}, {
		0,
		"%+s",
		"unknown",
	}, {
		initpc,
		"%d",
		"30",
	}, {
		0,
		"%d",
		"0",
	}, {
		initpc,
		"%n",
		"init",
	}, {
		func() Frame {
			var x X
			return x.ptr()
		}(),
		"%n",
		`\(\*X\).ptr`,
	}, {
		func() Frame {
			var x X
			return x.val()
		}(),
		"%n",
		"X.val",
	}, {
		0,
		"%n",
		"",
	}, {
		initpc,
		"%v",
		"stack_test.go:30",
	}, {
		initpc,
		"%+v",
		"errors.init\nstack_test.go:30",
	}, {
		0,
		"%v",
		"unknown:0",
	}, {
		0,
		"%+v",
		"unknown\n\tunknown:0",
	}}

	for i, tt := range tests {
		testFormatRegexp(t, i, tt.Frame, tt.format, tt.want)
	}
}

func TestFuncname(t *testing.T) {
	tests := []struct {
		name, want string
	}{
		{"", ""},
		{"runtime.main", "main"},
		{"github.com/ice-coldbell/analyze-server/pkg/errors.funcname", "funcname"},
		{"funcname", "funcname"},
		{"io.copyBuffer", "copyBuffer"},
		{"main.(*R).Write", "(*R).Write"},
	}

	for _, tt := range tests {
		got := funcname(tt.name)
		want := tt.want
		if got != want {
			t.Errorf("funcname(%q): want: %q, got %q", tt.name, want, got)
		}
	}
}

func TestStackTraceFormat(t *testing.T) {
	tests := []struct {
		StackTrace
		format string
		want   string
	}{
		{
			nil,
			"%s",
			`\[\]`,
		},
		{
			nil,
			"%v",
			`\[\]`,
		},
		{
			nil,
			"%+v",
			"",
		},
		{
			nil,
			"%#v",
			`\[\]errors.Frame\(nil\)`,
		},
		{
			make(StackTrace, 0),
			"%s",
			`\[\]`,
		},
		{
			make(StackTrace, 0),
			"%v",
			`\[\]`,
		},
		{
			make(StackTrace, 0),
			"%+v",
			"",
		},
		{
			make(StackTrace, 0),
			"%#v",
			`\[\]errors.Frame{}`,
		},
		{
			stackTrace()[:2],
			"%s",
			`\[stack_test.go stack_test.go\]`,
		},
		{
			stackTrace()[:2],
			"%v",
			`[stack_test.go:153 stack_test.go:200]`,
		},
		{
			stackTrace()[:2],
			"%#v",
			`\[\]errors.Frame{stack_test.go:25, stack_test.go:192}`,
		},
		{
			stackTrace()[:2],
			"%+v",
			"\nerrors.stackTrace\nstack_test.go:25\nerrors.TestStackTraceFormat\nstack_test.go:197",
		},
	}

	for i, tt := range tests {
		testFormatRegexp(t, i, tt.StackTrace, tt.format, tt.want)
	}
}

func TestFrameMarshalText(t *testing.T) {
	var tests = []struct {
		Frame
		format string
		want   string
	}{{
		initpc,
		"%s",
		"stack_test.go",
	}, {
		0,
		"%s",
		"unknown",
	}, {
		0,
		"%+s",
		"unknown",
	}, {
		initpc,
		"%d",
		"30",
	}, {
		initpc,
		"%n",
		"init",
	}, {
		func() Frame {
			var x X
			return x.ptr()
		}(),
		"%n",
		`\(\*X\).ptr`,
	}, {
		func() Frame {
			var x X
			return x.val()
		}(),
		"%n",
		"X.val",
	}, {
		0,
		"%n",
		"",
	}, {
		initpc,
		"%v",
		"stack_test.go:30",
	}, {
		0,
		"%v",
		"unknown",
	}}

	for i, tt := range tests {
		data, err := tt.Frame.MarshalText()
		assert.NoError(t, err)
		testFormatRegexp(t, i, string(data), tt.format, tt.want)
	}
}

func TestStackFormat(t *testing.T) {
	tests := []struct {
		*stack
		format string
		want   string
	}{
		{
			nil,
			"%+v",
			"<nil>",
		},
		{
			func() *stack {
				zero := make(stack, 0)
				return &zero
			}(),
			"%+v",
			"",
		},
		{
			callers(3),
			"%+v",
			"\ntesting.tRunner\n\nruntime.goexit",
		},
	}

	for i, tt := range tests {
		testFormatRegexp(t, i, tt.stack, tt.format, tt.want)
	}
}

func testFormatRegexp(t *testing.T, n int, arg interface{}, format, want string) {
	t.Helper()
	got := fmt.Sprintf(format, arg)
	gotLines := strings.SplitN(got, "\n", -1)
	wantLines := strings.SplitN(want, "\n", -1)

	if len(wantLines) > len(gotLines) {
		t.Errorf("test %d: wantLines(%d) > gotLines(%d):\n got: %q\nwant: %q", n+1, len(wantLines), len(gotLines), got, want)
		return
	}

	for i, w := range wantLines {
		match, err := regexp.MatchString(w, gotLines[i])
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("test %d: line %d: fmt.Sprintf(%q, err):\n got: %q\nwant: %q", n+1, i+1, format, got, want)
		}
	}
}
