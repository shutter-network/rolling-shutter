package cltrtopics

import (
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrtopics"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprtopics"
	"github.com/shutter-network/shutter/shuttermint/snapshot/snptopics"
)

const (
	CipherBatch       = dcrtopics.CipherBatch
	DecryptionTrigger = kprtopics.DecryptionTrigger
	TimedEpoch        = snptopics.TimedEpoch
)
