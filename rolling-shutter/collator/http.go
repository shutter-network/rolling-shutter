package collator

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v4"
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
	epoch, _, err := batchhandler.GetNextBatch(req.Context(), db)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(oapi.NextEpoch{
		Id: epoch.Bytes(),
	})
}

func (srv *server) SubmitTransaction(w http.ResponseWriter, r *http.Request) {
	// FIXME undefined
	var x oapi.SubmitTransactionJSONBody
	if err := json.NewDecoder(r.Body).Decode(&x); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid format for SubmitTransaction")
		return
	}
	ctx := r.Context()

	hash := sha3.New256()
	fmt.Fprintf(hash, "%d\n", len(x.Epoch))
	hash.Write(x.Epoch)
	hash.Write(x.EncryptedTx)
	txid := hash.Sum(nil)

	// NOTE: We still have to decide how the caller can query for tx
	// success / failure.

	// there are some conditions where the tx can fail directly
	// this should be checked, and then the request should return

	// if initially valid, it could fail later
	// during inclusion in the batch, e.g. because of nonce mismatch
	// or lack of funds

	err := srv.c.batcher.EnqueueTx(ctx, x.EncryptedTx)
	if err != nil {
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
