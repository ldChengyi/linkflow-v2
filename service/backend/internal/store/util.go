package store

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
