package keyper

// EonPublicKey describes a successfully generated eon public key that other
// components in the keyper stack want to consume.
type EonPublicKey struct {
	PublicKey         []byte
	ActivationBlock   uint64
	KeyperConfigIndex uint64
	Eon               uint64
}
