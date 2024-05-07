package beaconapiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type GetProposerDutiesResponse struct {
	ExecutionOptimistic bool           `json:"execution_optimistic"`
	DependentRoot       string         `json:"dependent_root"`
	Data                []ProposerDuty `json:"data"`
}

type ProposerDuty struct {
	// Pubkey         blst.P1Affine `json:"pubkey"`
	Pubkey         string `json:"pubkey"`
	ValidatorIndex uint64 `json:"validator_index,string"`
	Slot           uint64 `json:"slot,string"`
}

func (c *Client) GetProposerDutiesByEpoch(
	ctx context.Context,
	epoch uint64,
) (*GetProposerDutiesResponse, error) {
	path := c.url.JoinPath("/eth/v1/validator/duties/proposer/", fmt.Sprint(epoch))
	req, err := http.NewRequestWithContext(ctx, "GET", path.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get proposer duties for epoch %d from consensus node", epoch)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(err, "failed to get proposer duties for epoch %d from consensus node", epoch)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read consensus client response body")
	}

	response := new(GetProposerDutiesResponse)
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal consensus client response body")
	}

	return response, nil
}

func (r *GetProposerDutiesResponse) GetDutyForSlot(slot uint64) (ProposerDuty, error) {
	for _, duty := range r.Data {
		if duty.Slot == slot {
			return duty, nil
		}
	}
	return ProposerDuty{}, errors.Errorf("consensus client response does not contain duty for slot %d", slot)
}
