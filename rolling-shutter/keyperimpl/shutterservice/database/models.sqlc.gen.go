// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package database

type CurrentDecryptionTrigger struct {
	Eon                  int64
	TriggeredBlockNumber int64
	IdentitiesHash       []byte
}

type DecryptionSignature struct {
	Eon            int64
	KeyperIndex    int64
	IdentitiesHash []byte
	Signature      []byte
}

type IdentityRegisteredEvent struct {
	BlockNumber    int64
	BlockHash      []byte
	TxIndex        int64
	LogIndex       int64
	Eon            int64
	IdentityPrefix []byte
	Sender         string
	Timestamp      int64
	Decrypted      bool
	Identity       []byte
}

type IdentityRegisteredEventsSyncedUntil struct {
	EnforceOneRow bool
	BlockHash     []byte
	BlockNumber   int64
}
