package entry

func CloneBytes(src []byte) []byte {
	if src == nil {
		return nil
	}
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
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
	// Vary 必须深拷贝，否则克隆体与源共享底层数组，外部修改会写穿缓存条目。
	out.Vary = append([]string(nil), e.Vary...)
	return out
}
