package mocksequencer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog/log"
	txtypes "github.com/shutter-network/txtypes/types"
)

func makeTx(batchIndex, nonce int, gas uint64) []byte {
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
		EncryptedPayload: []byte("foo"),
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
	txData := &txtypes.BatchTx{
		ChainID:       big.NewInt(1),
		DecryptionKey: []byte("foo"),
		BatchIndex:    1,
		L1BlockNumber: 42,
		Timestamp:     big.NewInt(1231231),
		Transactions:  [][]byte{makeTx(1, 1, 200000)},
	}
	tx := txtypes.NewTx(txData)
	txBinary, err := tx.MarshalBinary()
	if err != nil {
		return
	}
	txHex := hexutil.Encode(txBinary)
	log.Debug().Str("transaction", txHex).Msg("here is a dummy transaction for you, enjoy")
}
