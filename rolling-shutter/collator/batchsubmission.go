package collator

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shuttermint/medley"
)

const (
	batchSubmissionInterval   = 1 * time.Second
	submitBatchRequestTimeout = 1 * time.Second
)

func newJSONRPCRequest(id int, method string, params []string) medley.RPCRequest {
	return medley.RPCRequest{
		Version: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}
}

func newSubmitShutterBatchRequest(id int, batch []byte) medley.RPCRequest {
	batchHex := "0x" + hex.EncodeToString(batch)
	return newJSONRPCRequest(id, "collator_submitShutterBatch", []string{batchHex})
}

// submitBatches submits encrypted batches with the corresponding decryption keys to the collator.
func (c *collator) submitBatches(ctx context.Context) error {
	for {
		select {
		case <-time.After(batchSubmissionInterval):
			// TODO: query last batch index

			// TODO: get batch from db
			batch := []byte{}

			// submit batch to collator
			err := c.submitBatch(ctx, batch)
			if err != nil {
				log.Printf("error submitting batch to sequencer: %s", err)
				// we don't return and will just try next loop iteration
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *collator) submitBatch(ctx context.Context, batch []byte) error {
	reqBodyStruct := newSubmitShutterBatchRequest(0, batch)
	reqBodyEncoded, err := json.Marshal(reqBodyStruct)
	if err != nil {
		return err
	}
	reqBody := bytes.NewBuffer(reqBodyEncoded)

	reqCtx, cancelReqCtx := context.WithTimeout(ctx, submitBatchRequestTimeout)
	defer cancelReqCtx()

	req, err := http.NewRequestWithContext(reqCtx, "POST", c.Config.SequencerURL, reqBody)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	log.Println("submitting batch to sequecer")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("request failed with status %s", resp.Status)
	}
	return nil
}
