package hubapi

import (
	"encoding/hex"
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

//func (hub *HubAPI) doGet(endpoint string) ([]byte, error) {
//	resp, err := hub.Client.Get(hub.BaseURL + endpoint)
//	if err != nil {
//		return nil, err
//	}
//	defer func(Body io.ReadCloser) {
//		_ = Body.Close()
//	}(resp.Body)
//
//	body, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return nil, err
//	}
//	return body, nil
//}
//
//func (hub *HubAPI) doPost(endpoint string, qs string) ([]byte, error) {
//	resp, err := hub.Client.Post(fmt.Sprintf("%s%s?%s", hub.BaseURL, endpoint, qs), "application/x-www-form-urlencoded", nil)
//	if err != nil {
//		return nil, err
//	}
//	defer func(Body io.ReadCloser) {
//		_ = Body.Close()
//	}(resp.Body)
//
//	body, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return nil, err
//	}
//	return body, nil
//}

//func (hub *HubAPI) pollProposals() ([]Proposal, error) {
//	body, err := hub.doGet("/proposals")
//	if err != nil {
//		return nil, err
//	}
//
//	var proposals []Proposal
//	err = json.Unmarshal(body, &proposals)
//	return proposals, nil
//}

//func (hub *HubAPI) GetFinishedProposals() ([]Proposal, error) {
//	proposals, err := hub.pollProposals()
//	if err != nil {
//		return nil, err
//	}
//	now := time.Now().UTC()
//	var finishedProposals []Proposal
//	for _, proposal := range proposals {
//		if now.After(proposal.Closes.Time) {
//			finishedProposals = append(finishedProposals, proposal)
//		}
//	}
//	return finishedProposals, nil
//}

func (hub *HubAPI) SubmitEonKey(eonId uint64, key []byte) error {
	params := []string{strconv.FormatUint(eonId, 10), hex.EncodeToString(key)}
	var result bool
	err := hub.Client.Request(nil, hub.BaseURL, "shutter_set_eon_pubkey", params, &result)
	if err != nil {
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
