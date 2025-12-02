package beaconapiclient

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	blst "github.com/supranational/blst/bindings/go"
)

type GetValidatorByIndexResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                []ValidatorData
}

type ValidatorData struct {
	Index     uint64 `json:"index,string"`
	Balance   uint64 `json:"balance,string"`
	Status    string `json:"status"`
	Validator Validator
}

type Validator struct {
	PubkeyHex                  string `json:"pubkey"`
	WithdrawalCredentials      string `json:"withdrawal_credentials"`
	EffectiveBalance           uint64 `json:"effective_balance,string"`
	Slashed                    bool   `json:"slashed"`
	ActivationEligibilityEpoch uint64 `json:"activation_eligibility_epoch,string"`
	ActivationEpoch            uint64 `json:"activation_epoch,string"`
	ExitEpoch                  uint64 `json:"exit_epoch,string"`
	WithdrawalEpoch            uint64 `json:"withdrawal_epoch,string"`
}

func (c *Client) GetValidatorByIndices(
	ctx context.Context,
	stateID string,
	validatorIndices []int64,
) (*GetValidatorByIndexResponse, error) {
	path := c.url.JoinPath("/eth/v1/beacon/states/", stateID, "/validators/")
	query := url.Values{}
	for _, index := range validatorIndices {
		query.Add("id", fmt.Sprint(index))
	}
	path.RawQuery = query.Encode()

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

func (v *Validator) GetPubkey() (*blst.P1Affine, error) {
	pubkeyHex := v.PubkeyHex
	if pubkeyHex[:2] == "0x" {
		pubkeyHex = pubkeyHex[2:]
	}

	pubkeyBytes, err := hex.DecodeString(pubkeyHex)
	if err != nil {
		return nil, errors.Wrap(err, "failed to hex decode validator pubkey")
	}

	pubkey := new(blst.P1Affine)
	pubkey = pubkey.Uncompress(pubkeyBytes)
	if pubkey == nil {
		return nil, errors.New("failed to deserialize pubkey")
	}

	return pubkey, nil
}
