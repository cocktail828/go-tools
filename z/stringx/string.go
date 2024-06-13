package stringx

// report whether 'array' contains string 's'
func Contains(array []string, s string) bool {
	for _, a := range array {
		if a == s {
			return true
		}
	}
	return false
}

// report whether 's' is a member of 'array'
func Oneof(s string, array ...string) bool {
	for _, a := range array {
		if a == s {
			return true
		}
	}
	return false
}
