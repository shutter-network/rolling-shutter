package syncer

// IndexRange is a closed interval [Start, End] of keyper set indices.
type IndexRange struct {
	Start uint64
	End   uint64
}

// IndicesToRanges converts a sorted slice of int64 indices into contiguous ranges.
// Indices must be non-negative; this matches the int64 type returned by sqlc for bigint columns.
func IndicesToRanges(indices []int64) []IndexRange {
	if len(indices) == 0 {
		return nil
	}
	var ranges []IndexRange
	start := uint64(indices[0]) //nolint:gosec
	end := start
	for _, idx := range indices[1:] {
		v := uint64(idx) //nolint:gosec
		if v == end+1 {
			end = v
		} else {
			ranges = append(ranges, IndexRange{Start: start, End: end})
			start = v
			end = v
		}
	}
	return append(ranges, IndexRange{Start: start, End: end})
}

// complementRanges returns the ranges in [0, total) not covered by known.
// known must be sorted and non-overlapping.
func complementRanges(known []IndexRange, total uint64) []IndexRange {
	var result []IndexRange
	cursor := uint64(0)
	for _, r := range known {
		if r.Start > cursor {
			result = append(result, IndexRange{Start: cursor, End: r.Start - 1})
		}
		if r.End+1 > cursor {
			cursor = r.End + 1
		}
	}
	if cursor < total {
		result = append(result, IndexRange{Start: cursor, End: total - 1})
	}
	return result
}
