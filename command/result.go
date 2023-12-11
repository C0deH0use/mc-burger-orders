package command

type TypedResult struct {
	Result bool
	Type   string
	Error  error
}

func NewErrorResult(typeName string, err error) TypedResult {
	return TypedResult{
		Result: false,
		Type:   typeName,
		Error:  err,
	}
}

func NewSuccessfulResult(typeName string) TypedResult {
	return TypedResult{
		Result: true,
		Type:   typeName,
	}
}

func NewFailedResult(typeName string) TypedResult {
	return TypedResult{
		Result: false,
		Type:   typeName,
	}
}
