package utils

func PtrString(str string) *string {
	if str != "" {
		return &str
	}

	return nil
}
