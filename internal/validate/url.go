package validate

import "strings"

func URL(u string) bool {
	return strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "mock://")
}

func Method(m string) bool {
	switch strings.ToUpper(m) {
	case "GET", "HEAD", "POST", "PUT", "DELETE", "PATCH":
		return true
	default:
		return false
	}
}
