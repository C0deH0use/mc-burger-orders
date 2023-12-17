package command

import (
	"fmt"
	"net/http"
)

type TypedResult struct {
	Result bool
	Type   string
	Error  *HttpError
}

type HttpError struct {
	ErrorMessage string
	HttpResponse int
}

func (e *HttpError) Error() error {
	return fmt.Errorf(e.ErrorMessage)
}

func NewSuccessfulResult(typeName string) TypedResult {
	return TypedResult{
		Result: true,
		Type:   typeName,
		Error:  nil,
	}
}

func NewErrorResult(typeName string, err error) TypedResult {
	return TypedResult{
		Result: false,
		Type:   typeName,
		Error:  &HttpError{ErrorMessage: err.Error(), HttpResponse: http.StatusBadRequest},
	}
}

func NewHttpErrorResult(typeName string, errMessage string, httpResponse int) TypedResult {
	return TypedResult{
		Result: false,
		Type:   typeName,
		Error:  &HttpError{ErrorMessage: errMessage, HttpResponse: httpResponse},
	}
}
