// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package database

import (
	"database/sql"
)

type CurrentDecryptionTrigger struct {
	Eon            int64
	Slot           int64
	TxPointer      int64
	IdentitiesHash []byte
}

type SlotDecryptionSignature struct {
	Eon            int64
	Slot           int64
	KeyperIndex    int64
	TxPointer      int64
	IdentitiesHash []byte
	Signature      []byte
}

type TransactionSubmittedEvent struct {
	Index          int64
	BlockNumber    int64
	BlockHash      []byte
	TxIndex        int64
	LogIndex       int64
	Eon            int64
	IdentityPrefix []byte
	Sender         string
	GasLimit       int64
}

type TransactionSubmittedEventCount struct {
	Eon        int64
	EventCount int64
}

type TransactionSubmittedEventsSyncedUntil struct {
	EnforceOneRow bool
	BlockHash     []byte
	BlockNumber   int64
	Slot          int64
}

type TxPointer struct {
	Eon   int64
	Age   sql.NullInt64
	Value int64
}

type ValidatorRegistration struct {
	BlockNumber    int64
	BlockHash      []byte
	TxIndex        int64
	LogIndex       int64
	ValidatorIndex int64
	Nonce          int64
	IsRegistration bool
}

type ValidatorRegistrationsSyncedUntil struct {
	EnforceOneRow bool
	BlockHash     []byte
	BlockNumber   int64
}
