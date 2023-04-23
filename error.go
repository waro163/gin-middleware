package ginmiddleware

import "fmt"

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var _ error = (*Error)(nil)

func (e *Error) Error() string {
	return fmt.Sprintf("Code: %d, Message: %s", e.Code, e.Message)
}
