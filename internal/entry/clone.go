package entry

// CloneBytes returns an independent copy of src so callers cannot mutate the
// cached body through the returned slice (and vice versa). Returning src
// directly would alias the stored entry's backing array: a caller writing to
// Response.Body[0] would corrupt the in-cache body that other tenants read.
func CloneBytes(src []byte) []byte {
	if src == nil {
		return nil
	}
	out := make([]byte, len(src))
	copy(out, src)
	return out
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
