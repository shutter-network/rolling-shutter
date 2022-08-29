package mocksequencer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog/log"
	txtypes "github.com/shutter-network/txtypes/types"
)

func makeTx(l1BlockNumber, batchIndex, nonce int, gas uint64) []byte {
	privKey, err := ethcrypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	chainID := big.NewInt(1)
	signer := txtypes.NewLondonSigner(chainID)
	// construct a valid transaction
	txData := &txtypes.ShutterTx{
		ChainID:          chainID,
		Nonce:            uint64(nonce),
		GasTipCap:        big.NewInt(2000000),
		GasFeeCap:        big.NewInt(2),
		Gas:              gas,
		EncryptedPayload: []byte("bar"),
		L1BlockNumber:    uint64(l1BlockNumber),
		BatchIndex:       uint64(batchIndex),
	}

	tx, err := txtypes.SignNewTx(privKey, signer, txData)
	if err != nil {
		panic(err)
	}
	// marshal tx to bytes
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return txBytes
}

func logDummyTransaction() {
	chainID := big.NewInt(1)
	l1BlockNumber := 42
	batchIndex := 1
	txData := &txtypes.BatchTx{
		ChainID:       chainID,
		DecryptionKey: []byte("foo"),
		BatchIndex:    uint64(batchIndex),
		L1BlockNumber: uint64(l1BlockNumber),
		Timestamp:     big.NewInt(1231231),
		Transactions:  [][]byte{makeTx(l1BlockNumber, batchIndex, 1, 200000)},
	}
	privKey, err := ethcrypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	signer := txtypes.NewLondonSigner(chainID)

	tx, err := txtypes.SignNewTx(privKey, signer, txData)
	if err != nil {
		panic(err)
	}

	collator, err := signer.Sender(tx)
	if err != nil {
		panic(err)
	}
	txBinary, err := tx.MarshalBinary()
	if err != nil {
		return
	}
	txHex := hexutil.Encode(txBinary)
	log.Debug().Str("transaction", txHex).Str("collator", collator.Hex()).Msg("here is a dummy transaction for you, enjoy")
}
