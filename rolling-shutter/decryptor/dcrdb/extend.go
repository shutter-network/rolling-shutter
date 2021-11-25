package dcrdb

func SearchDecryptorSetRowsForIndex(rows []GetDecryptorSetRow, index int32) (GetDecryptorSetRow, bool) {
	for _, row := range rows {
		if row.Index == index {
			return row, true
		}
	}
	return GetDecryptorSetRow{}, false
}
