package stringx

func toMap(s []string) map[string]struct{} {
	tmp := make(map[string]struct{}, len(s))
	for _, _s := range s {
		tmp[_s] = struct{}{}
	}
	return tmp
}

func fromMap(m map[string]struct{}) []string {
	strs := make([]string, 0, len(m))
	for s := range m {
		strs = append(strs, s)
	}
	return strs
}

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

// report whether 's1', 's2' as equal(ignore the sort)
func Equal(s1, s2 []string) bool {
	return len(s1) == len(s2) && len(Elimate(s1, s2)) == 0
}

// report the common part of 's1' and 's2'
func Overlap(s1, s2 []string) []string {
	tmp := toMap(s1)
	rlt := []string{}
	for _, s := range s2 {
		if _, ok := tmp[s]; ok {
			rlt = append(rlt, s)
		}
	}
	return rlt
}

// report the incommon part of 's1' with 's2'
func Elimate(s1, s2 []string) []string {
	tmp := toMap(s1)
	for _, s := range s2 {
		delete(tmp, s)
	}
	return fromMap(tmp)
}
