package decryptor

import (
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
)

type Decryptor struct {
	db           *dcrdb.Queries
	inputChannel <-chan interface{}
}

func NewDecryptor(db *dcrdb.Queries) *Decryptor {
	return &Decryptor{
		db:           db,
		inputChannel: make(<-chan interface{}),
	}
}
