package utils

func ErrorPayload(err string) map[string]string {
	return map[string]string{
		"errorMessage": err,
	}
}
