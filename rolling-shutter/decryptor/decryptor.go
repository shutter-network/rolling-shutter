package decryptor

import (
	"context"

	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
)

type Decryptor struct {
	db            *dcrdb.Queries
	inputChannel  <-chan interface{}
	outputChannel chan interface{}
}

func NewDecryptor(db *dcrdb.Queries) *Decryptor {
	return &Decryptor{
		db:            db,
		inputChannel:  make(<-chan interface{}),
		outputChannel: make(chan interface{}),
	}
}

func (d *Decryptor) Run(ctx context.Context) error {
	return d.handleInputs(ctx)
}
