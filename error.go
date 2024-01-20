package erw

import (
	"errors"
	"fmt"
	"strings"
)

type erw struct {
	e     error
	stack *stack
}

type rootErr struct {
	erw   erw
	cause []causeErr
}

type causeErr erw

func (e *rootErr) Error() string {
	return e.erw.e.Error()
}

func (e *causeErr) Error() string {
	return e.e.Error()
}

func (e *rootErr) Unwrap() []error {
	errs := make([]error, len(e.cause))
	for i := range e.cause {
		errs[i] = &e.cause[i]
	}
	return errs
}

//goland:noinspection GoErrorsAs
func (e *rootErr) As(out any) bool {
	if errors.As(e.erw.e, out) {
		return true
	}

	for i := range e.cause {
		if errors.As(e.cause[i].e, out) {
			return true
		}
	}

	return false
}

func (e *rootErr) Is(err error) bool {
	if errors.Is(e.erw.e, err) {
		return true
	}

	for i := range e.cause {
		if errors.Is(e.cause[i].e, err) {
			return true
		}
	}

	return false
}

func New(msg string) error {
	stack := callers(3)
	return &rootErr{
		erw: erw{
			e:     errors.New(msg),
			stack: stack,
		},
		cause: make([]causeErr, 0),
	}
}

//goland:noinspection GoTypeAssertionOnErrors
func S(e error) error {
	stack := callers(3)

	switch err := e.(type) {
	case *rootErr:
		if !err.erw.stack.isGlobal() {
			return err
		}

		return &rootErr{
			erw: erw{
				e:     err.erw.e,
				stack: stack,
			},
			cause: err.cause,
		}
	default:
		return &rootErr{
			erw: erw{
				e:     err,
				stack: stack,
			},
			cause: nil,
		}
	}
}

//goland:noinspection GoTypeAssertionOnErrors
func Wrap(cause error, err error) error {
	if cause == nil {
		return nil
	}

	stack := callers(3)

	var rErr = rootErr{}

	switch e := cause.(type) {
	case *rootErr:
		rErr.cause = append(e.cause, causeErr(e.erw))
	case *causeErr:
		rErr.cause = append(rErr.cause, *e)
	default:
		rErr.cause = append(rErr.cause, causeErr{
			e:     e,
			stack: nil,
		})
	}

	switch e := err.(type) {
	case *rootErr:
		rErr.erw = e.erw
		rErr.cause = append(rErr.cause, e.cause...)
	case *causeErr:
		rErr.erw = erw(*e)
	default:
		rErr.erw = erw{
			e:     e,
			stack: stack,
		}
	}

	return &rErr
}

func Stringify(e error) string {
	var root *rootErr
	if !errors.As(e, &root) {
		return e.Error()
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Error: %v\n", root.erw.e))
	if root.erw.stack != nil && !root.erw.stack.isGlobal() {
		sb.WriteString(fmt.Sprintf("  StackTrace: \n%v", root.erw.stack.String()))
	}

	for i := range root.cause {
		idx := len(root.cause) - 1 - i

		sb.WriteString(fmt.Sprintf("  Cause: %v\n", root.cause[idx].e))

		if root.cause[idx].stack != nil && !root.cause[idx].stack.isGlobal() {
			sb.WriteString(fmt.Sprintf("  StackTrace: \n%v\n", root.erw.stack.String()))
		}
	}

	return strings.TrimSuffix(sb.String(), "\n")
}
