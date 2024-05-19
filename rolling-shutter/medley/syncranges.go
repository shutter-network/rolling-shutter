package medley

func GetSyncRanges(start, end, maxRange uint64) [][2]uint64 {
	ranges := [][2]uint64{}
	for i := start; i <= end; i += maxRange {
		s := i
		e := i + maxRange - 1
		ranges = append(ranges, [2]uint64{s, e})
		if e > end {
			ranges[len(ranges)-1][1] = end
			break
		}
	}
	return ranges
}
