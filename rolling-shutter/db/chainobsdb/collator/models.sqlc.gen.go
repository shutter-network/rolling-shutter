// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1

package chainobsdb

import ()

type ChainCollator struct {
	ActivationBlockNumber int64  `db:"activation_block_number"`
	Collator              string `db:"collator"`
}
