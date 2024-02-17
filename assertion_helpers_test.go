package cproject_test

// StringSlicesEqual returns true if two string slices have identical strings in the exact same order.
func StringSlicesEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}
