// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package database

type DecryptionKey struct {
	EpochID []byte
	Key     []byte
}

type EonPublicKey struct {
	EonID        int64
	EonPublicKey []byte
}
