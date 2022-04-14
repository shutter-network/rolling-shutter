package keyper

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v4"

	"github.com/shutter-network/shutter/shuttermint/keyper/kprdb"
	"github.com/shutter-network/shutter/shuttermint/keyper/kproapi"
	"github.com/shutter-network/shutter/shuttermint/shdb"
)

type server struct {
	kpr *keyper
}

func sendError(w http.ResponseWriter, code int, message string) {
	e := kproapi.Error{
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

func (srv *server) GetDecryptionKey(w http.ResponseWriter, r *http.Request, epochID kproapi.EpochID) {
	ctx := r.Context()
	db := kprdb.New(srv.kpr.dbpool)

	epochIDBytes, err := hex.DecodeString(strings.TrimPrefix(string(epochID), "0x"))
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	decryptionKey, err := db.GetDecryptionKey(ctx, epochIDBytes)
	if err == pgx.ErrNoRows {
		sendError(w, http.StatusNotFound, "no decryption key found for given epoch")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	res := "0x" + hex.EncodeToString(decryptionKey.DecryptionKey)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (srv *server) GetEons(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := kprdb.New(srv.kpr.dbpool)

	res := kproapi.Eons{}

	eons, err := db.GetAllEons(ctx)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, eon := range eons {
		var eonKey []byte
		var finished bool
		var successful bool

		encodedDKGResult, err := db.GetDKGResult(ctx, eon.Eon)
		if err == pgx.ErrNoRows {
			eonKey = []byte{}
			finished = false
			successful = false
		} else if err != nil {
			log.Println("failed to get dkg result from db")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		} else {
			finished = true
			successful = encodedDKGResult.Success
			if successful {
				dkgResult, err := shdb.DecodePureDKGResult(encodedDKGResult.PureResult)
				if err != nil {
					sendError(w, http.StatusInternalServerError, err.Error())
					return
				}
				eonKey = dkgResult.PublicKey.Marshal()
			} else {
				eonKey = []byte{}
			}
		}
		res = append(res, kproapi.Eon{
			Index:                 int(eon.Eon),
			ActivationBlockNumber: int(eon.ActivationBlockNumber),
			EonKey:                "0x" + hex.EncodeToString(eonKey),
			Finished:              finished,
			Successful:            successful,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (srv *server) SubmitDecryptionTrigger(w http.ResponseWriter, r *http.Request) {
	var requestBody kproapi.SubmitDecryptionTriggerJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request for SubmitDecryptionTrigger")
		return
	}
	epochIDBytes, err := hex.DecodeString(strings.TrimPrefix(string(requestBody), "0x"))
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}
	epochID := binary.BigEndian.Uint64(epochIDBytes)

	ctx := r.Context()
	handler := epochKGHandler{
		config: srv.kpr.config,
		db:     srv.kpr.db,
	}
	msgs, err := handler.sendDecryptionKeyShare(ctx, epochID)
	if err != nil {
		if err != nil {
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	for _, msg := range msgs {
		if err := srv.kpr.p2p.SendMessage(ctx, msg); err != nil {
			log.Printf("error sending message %+v: %s", msg, err)
			continue
		}
	}
}
