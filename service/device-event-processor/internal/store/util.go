package store

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func emptyMapAsNil(m map[string]any) any {
	if len(m) == 0 {
		return nil
	}
	return m
}
