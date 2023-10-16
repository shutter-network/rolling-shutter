package collator

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/sha3"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/oapi"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/cltrdb"
)

type server struct {
	c *collator
}

func sendError(w http.ResponseWriter, code int, message string) {
	e := oapi.Error{
		Code:    int32(code),
		Message: message,
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(e)
}

func (srv *server) Ping(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("pong"))
}

func (srv *server) GetNextEpoch(w http.ResponseWriter, req *http.Request) {
	db := cltrdb.New(srv.c.dbpool)
	// TODO the get-next batch method should also
	// take a delta, that will infer the epoch (easy)
	// as well as the corresponding l1-block-number
	// by calculating the estimated l1 block at the offset batch
	//  (harder)
	epoch, _, err := batchhandler.GetNextBatch(req.Context(), db)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
	}

	batchOffset := 4
	// inefficient, but for this poc okay.
	// calculate a batch in the near future so that we don't miss it
	for i := 0; i < batchOffset; i++ {
		epoch, err = batchhandler.ComputeNextEpochID(epoch)
		if err != nil {
			sendError(w, http.StatusInternalServerError, err.Error())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(oapi.NextEpoch{
		Id:    epoch.Bytes(),
		Batch: epoch.Uint64(),
		// NOTE: for this proof-of-concept, the check in the mocksequencer
		// is disabled, this means we just have to
		// use the block-number for the current collator.
		// We'll only have 1 collator with start-block at 0,
		// so this is ok.
		L1BlockNumber: 0,
	})
}

// transaction={
// "batchIndex":"0x1c",
// "chainId":"0x1",
// "decryptionKey":"0x0b0ea77e9515f14d1639ea8c0c144b2ce23de11f0e699c7ad9fe6075c81c1a1814ff468980d90a62408befdfc223c234f79c1405b60efe97127991d7a28a0832",
// "from":null,
// "gas":null,
// "gasPrice":null,
// "hash":"0x196c6bf217baab42631c7e88e18cfa1c9224bc9d47fcca1d2ce57d7674ad0fba",
// "input":null,
// "l1BlockNumber":"0x105989c",
// "nonce":null,
// "r":"0x8e41fc3d7e37a5e2d4ff4a9264e568b4202740ce0dc559842716fe8e66d84362",
// "s":"0x243cebd4f25d2decfb3ab396548f5a92f980146ef0f4b6ef7ab9d8af650cbb23",
// "timestamp":"0x651fd38b",
// "to":null,
// "type":"0x5a",
// "v":"0x1",
// "value":null
// }
func (srv *server) SubmitTransaction(w http.ResponseWriter, r *http.Request) {
	// FIXME undefined
	var x oapi.SubmitTransactionJSONBody
	if err := json.NewDecoder(r.Body).Decode(&x); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid format for SubmitTransaction")
		return
	}

	bytesTx, err := hexutil.Decode(x.EncryptedTx)
	if err != nil {
		err = errors.Wrap(err, "failed to decode encrypted-tx")
		log.Info().Err(err).Msg("=========FAILED============")
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()

	epoch, err := hexutil.Decode(x.Epoch)
	if err != nil {
		err = errors.Wrap(err, "failed to decode epoch")
		log.Info().Err(err).Msg("=========FAILED============")
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	hash := sha3.New256()
	fmt.Fprintf(hash, "%d\n", len(epoch))
	hash.Write(epoch)
	hash.Write(bytesTx)
	txid := hash.Sum(nil)

	// NOTE: We still have to decide how the caller can query for tx
	// success / failure.

	// there are some conditions where the tx can fail directly
	// this should be checked, and then the request should return

	// if initially valid, it could fail later
	// during inclusion in the batch, e.g. because of nonce mismatch
	// or lack of funds

	err = srv.c.batcher.EnqueueTx(ctx, bytesTx)
	if err != nil {
		err = errors.Wrap(err, "could not enqueue encrypted transaction")
		log.Error().Err(err).Msg("Error in SubmitTransaction")
		sendError(w, http.StatusConflict, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(oapi.TransactionId{Id: txid})
}

func (srv *server) GetEonPublicKey(
	w http.ResponseWriter,
	r *http.Request,
	params oapi.GetEonPublicKeyParams,
) {
	var (
		eonPub     cltrdb.EonPublicKeyCandidate
		votes      []cltrdb.EonPublicKeyVote
		signatures [][]byte
	)
	ctx := r.Context()

	err := srv.c.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		var err error
		db := cltrdb.New(tx)
		eonPub, err = db.FindEonPublicKeyForBlock(ctx, params.ActivationBlock)
		if err != nil {
			return err
		}

		votes, err = db.FindEonPublicKeyVotes(ctx, eonPub.Hash)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(votes) == 0 {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(make(map[string]string))
		return
	}

	for _, v := range votes {
		signatures = append(signatures, v.Signature)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(oapi.Eon{
		ActivationBlockNumber: eonPub.ActivationBlockNumber,
		Eon:                   eonPub.Eon,
		EonPublicKey:          eonPub.EonPublicKey,
		InstanceId:            int64(srv.c.Config.InstanceID),
		KeyperConfigIndex:     eonPub.KeyperConfigIndex,
		Signatures:            signatures,
	})
}
