package p2pmsg

import (
	"fmt"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprtopics"
)

func (c *Commitment) LogInfo() string {
	return fmt.Sprintf("Commitment{commitment_digest=%s}", c.CommitmentDigest)
}

func (c *Commitment) Topic() string {
	return kprtopics.PrimevCommitment
}

func (c *Commitment) Validate() error {
	return nil
}
