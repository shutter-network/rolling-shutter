// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1

package chainobsdb

import ()

type KeyperSet struct {
	KeyperConfigIndex     int64    `db:"keyper_config_index"`
	ActivationBlockNumber int64    `db:"activation_block_number"`
	Keypers               []string `db:"keypers"`
	Threshold             int32    `db:"threshold"`
}
