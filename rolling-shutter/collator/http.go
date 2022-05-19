package collator

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/sha3"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/oapi"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
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
	epoch, err := getNextEpochID(req.Context(), db)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(oapi.NextEpoch{
		Id: shdb.EncodeUint64(epoch),
	})
}

func (srv *server) SubmitTransaction(w http.ResponseWriter, r *http.Request) {
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

	err := insertTx(ctx, srv.c.dbpool, cltrdb.InsertTxParams{
		TxID:        txid,
		EpochID:     x.Epoch,
		EncryptedTx: x.EncryptedTx,
	})
	if err != nil {
		log.Printf("Error in SubmitTransaction: %s", err)
		sendError(w, http.StatusConflict, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(oapi.TransactionId{Id: txid})
}

func (srv *server) GetEonPublicKey(w http.ResponseWriter, r *http.Request, params oapi.GetEonPublicKeyParams) {
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
