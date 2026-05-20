package gnosis

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/pkg/errors"
)

// polyEvalBlobArgs describes the ABI encoding used for the polyEval blob
// emitted in `DealingSubmitted` events: a single `bytes[]` with one entry per
// receiver in ascending receiver-index order, skipping the sender's own index.
var polyEvalBlobArgs = func() abi.Arguments {
	t, err := abi.NewType("bytes[]", "", nil)
	if err != nil {
		panic(err)
	}
	return abi.Arguments{{Type: t}}
}()

// EncodePolyEvalBlob ABI-encodes the per-receiver encrypted evaluations into a
// single blob suitable for `DKGContract.submitDealing`. The slice must be in
// ascending receiver-index order (skipping the sender's own index).
func EncodePolyEvalBlob(encryptedEvals [][]byte) ([]byte, error) {
	return polyEvalBlobArgs.Pack(encryptedEvals)
}

// DecodePolyEvalBlob is the inverse of EncodePolyEvalBlob.
func DecodePolyEvalBlob(blob []byte) ([][]byte, error) {
	values, err := polyEvalBlobArgs.Unpack(blob)
	if err != nil {
		return nil, errors.Wrap(err, "abi-decode polyEval blob")
	}
	if len(values) != 1 {
		return nil, errors.Errorf("polyEval blob: expected 1 ABI value, got %d", len(values))
	}
	evals, ok := values[0].([][]byte)
	if !ok {
		return nil, errors.New("polyEval blob: expected bytes[]")
	}
	return evals, nil
}

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
