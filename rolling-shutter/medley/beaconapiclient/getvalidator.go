package beaconapiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
	blst "github.com/supranational/blst/bindings/go"
)

type GetValidatorByIndexResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                ValidatorData
}

type ValidatorData struct {
	Index     uint64 `json:"index,string"`
	Balance   uint64 `json:"balance,string"`
	Status    string `json:"status"`
	Validator Validator
}

type Validator struct {
	Pubkey                     blst.P1Affine `json:"pubkey"`
	WithdrawalCredentials      string        `json:"withdrawal_credentials,string"`
	EffectiveBalance           uint64        `json:"effective_balance,string"`
	Slashed                    bool          `json:"slashed"`
	ActivationEligibilityEpoch uint64        `json:"activation_eligibility_epoch,string"`
	ActivationEpoch            uint64        `json:"activation_epoch,string"`
	ExitEpoch                  uint64        `json:"exit_epoch,string"`
	WithdrawalEpoch            uint64        `json:"withdrawal_epoch,string"`
}

func (c *Client) GetValidatorByIndex(
	ctx context.Context,
	stateID string,
	validatorIndex uint64,
) (*GetValidatorByIndexResponse, error) {
	path := c.url.JoinPath("/eth/v1/beacon/states/", stateID, "/validators/", fmt.Sprint(validatorIndex))
	req, err := http.NewRequestWithContext(ctx, "GET", path.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get validator by index from consensus node")
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.Wrap(err, "failed to get validator by index from consensus node")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read consensus client response body")
	}

	response := new(GetValidatorByIndexResponse)
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal consensus client response body")
	}

	return response, nil
}
