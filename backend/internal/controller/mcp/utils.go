package mcp

func paginate(items []int64, cursor int, chunkSize int) ([]int64, int) {
	if cursor < 0 {
		cursor = 0
	}

	totalItems := len(items)

	if cursor >= totalItems {
		return []int64{}, -1
	}

	end := cursor + chunkSize

	if end >= totalItems {
		return items[cursor:], -1
	}

	return items[cursor:end], end
}

// stringPtrToString converts a *string to a string, handling nil.
func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
