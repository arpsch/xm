package utils

// Check if string is presnt in an array. Would use interface{} but
// whatever.
func ContainsString(val string, vals []string) bool {

	for _, v := range vals {
		if val == v {
			return true
		}
	}
	return false
}
