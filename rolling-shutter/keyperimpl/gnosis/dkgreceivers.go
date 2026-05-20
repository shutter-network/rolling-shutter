package gnosis

// ReceiverIndicesForSender returns the keyper indices that should be receivers
// of polyEvals submitted by `senderIndex` within a keyper set of size `n`:
// all indices in [0, n) except `senderIndex`, in ascending order.
func ReceiverIndicesForSender(n, senderIndex uint64) []uint64 {
	out := make([]uint64, 0, n-1)
	for i := uint64(0); i < n; i++ {
		if i == senderIndex {
			continue
		}
		out = append(out, i)
	}
	return out
}
