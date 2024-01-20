package erw

import (
	"encoding/json"
	"fmt"
)

type CodeError struct {
	Code        string
	Message     string
	Description string
}

func (e CodeError) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"code":    e.Code,
		"message": e.Message,
	}

	if e.Description != "" {
		data["description"] = e.Description
	}

	return json.Marshal(data)
}

func (e CodeError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewCodeErr(code string, msg string) error {
	return CodeError{
		Code:    code,
		Message: msg,
	}
}

//goland:noinspection GoTypeAssertionOnErrors
func WithDescription(e error, description string) error {
	if ce, ok := e.(*CodeError); ok {
		ce.Description = description

		return ce
	}

	return Wrap(e, &CodeError{
		Code:        "unspecified",
		Message:     "неизвестная ошибка",
		Description: description,
	})
}
