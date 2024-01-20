package erw

import (
	"fmt"
	"runtime"
	"strings"
)

type StackFrame struct {
	Name string
	File string
	Line int
}

type stack []uintptr

func callers(skip int) *stack {
	const depth = 64
	var pcs [depth]uintptr
	n := runtime.Callers(skip, pcs[:])
	var st stack = pcs[0 : n-2] // todo: change this to filtering out runtime instead of hardcoding n-2
	return &st
}

func (s *stack) insertPC(wrapPCs stack) {
	if len(wrapPCs) == 0 {
		return
	} else if len(wrapPCs) == 1 {
		// append the pc to the end if there's only one
		*s = append(*s, wrapPCs[0])
		return
	}
	for at, f := range *s {
		if f == wrapPCs[0] {
			// break if the stack already contains the pc
			break
		} else if f == wrapPCs[1] {
			// insert the first pc into the stack if the second pc is found
			*s = insert(*s, wrapPCs[0], at)
			break
		}
	}
}

// get returns a human readable stack trace.
func (s *stack) get() []StackFrame {
	var stackFrames []StackFrame

	frames := runtime.CallersFrames(*s)
	for {
		frame, more := frames.Next()
		i := strings.LastIndex(frame.Function, "/")
		name := frame.Function[i+1:]
		stackFrames = append(stackFrames, StackFrame{
			Name: name,
			File: frame.File,
			Line: frame.Line,
		})
		if !more {
			break
		}
	}

	return stackFrames
}

func (s *stack) String() string {
	frames := s.get()

	var sb strings.Builder
	for _, f := range frames {
		sb.WriteString(fmt.Sprintf("    %s:%d: %v\n", f.File, f.Line, f.Name))
	}

	return sb.String()
}

// isGlobal determines if the stack trace represents a global error
func (s *stack) isGlobal() bool {
	frames := s.get()
	for _, f := range frames {
		if strings.ToLower(f.Name) == "runtime.doinit" {
			return true
		}
	}
	return false
}

func insert(s stack, u uintptr, at int) stack {
	// this inserts the pc by breaking the stack into two slices (s[:at] and s[at:])
	return append(s[:at], append([]uintptr{u}, s[at:]...)...)
}
