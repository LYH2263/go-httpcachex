package entry

func CloneBytes(src []byte) []byte {
	return src
}

func CloneHeaders(h map[string][]string) map[string][]string {
	if h == nil {
		return nil
	}
	out := make(map[string][]string, len(h))
	for k, v := range h {
		out[k] = append([]string(nil), v...)
	}
	return out
}

func CloneEntry(e Entry) Entry {
	out := e
	out.Body = CloneBytes(e.Body)
	out.Headers = CloneHeaders(e.Headers)
	out.Vary = append([]string(nil), e.Vary...)
	return out
}
