package key

import (
	"sort"
	"strings"
)

func VaryNames(h map[string][]string) []string {
	v := h["Vary"]
	if len(v) == 0 {
		v = h["vary"]
	}
	if len(v) == 0 {
		return nil
	}
	parts := strings.Split(v[0], ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func Build(method, url string, headers map[string][]string) string {
	vary := VaryNames(headers)
	var b strings.Builder
	b.WriteString(strings.ToUpper(method))
	b.WriteByte('|')
	b.WriteString(url)
	if len(vary) > 0 {
		names := append([]string(nil), vary...)
		sort.Strings(names)
		for _, n := range names {
			b.WriteByte('|')
			b.WriteString(n)
			b.WriteByte('=')
			vals := headers[n]
			if len(vals) == 0 {
				vals = headers[httpCanonical(n)]
			}
			b.WriteString(strings.Join(vals, ","))
		}
	}
	return b.String()
}

func httpCanonical(s string) string {
	return strings.Title(strings.ToLower(s))
}
