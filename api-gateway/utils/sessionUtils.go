package utils

// Contains checks if a string slice contains a specific target string
func Contains(sessions []string, target string) bool {
	for _, s := range sessions {
		if s == target {
			return true
		}
	}
	return false
} 