package model

func getString(m map[string]interface{}, field string) (string, bool) {
	r, ok := m[field]
	if !ok {
		return "", ok
	}

	rr, ok := r.(string)
	if !ok {
		return "", ok
	}

	return rr, true
}

// func getBool(m map[string]interface{}, field string) (bool, bool) {
// 	r, ok := m[field]
// 	if !ok {
// 		return false, ok
// 	}

// 	rr, ok := r.(bool)
// 	if !ok {
// 		return false, ok
// 	}

// 	return rr, true
// }
