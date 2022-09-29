package hubapi

import (
	"encoding/hex"
	"log"
	"strconv"

	"github.com/AdamSLevy/jsonrpc2/v14"
)

type HubAPI struct {
	BaseURL string
	Client  jsonrpc2.Client
}

func New(hubURL string) *HubAPI {
	return &HubAPI{
		BaseURL: hubURL,
		Client:  jsonrpc2.Client{},
	}
}

func (hub *HubAPI) SubmitEonKey(eonId uint64, key []byte) error {
	params := []string{strconv.FormatUint(eonId, 10), hex.EncodeToString(key)}
	var result bool
	err := hub.Client.Request(nil, hub.BaseURL, "shutter_set_eon_pubkey", params, &result)
	if err != nil {
		log.Printf("Error posting to HUB: %v", err)
		return err
	}
	return nil
}

func (hub *HubAPI) SubmitProposalKey(proposalId []byte, key []byte) error {
	params := []string{hex.EncodeToString(proposalId), hex.EncodeToString(key)}
	var result bool
	err := hub.Client.Request(nil, hub.BaseURL, "shutter_set_proposal_key", params, &result)
	if err != nil {
		return err
	}
	return nil
}
