package ptr

func FromStr(value string) *string {
	return &value
}

func FromInt(value int) *int {
	return &value
}

func FromBool(value bool) *bool {
	return &value
}
