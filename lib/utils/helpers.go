package utils

func ValIf(s string, cond bool) string {
	if cond {
		return s
	} else {
		return ""
	}
}
