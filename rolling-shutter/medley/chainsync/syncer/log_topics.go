package syncer

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var errTopicMismatch = errors.New("passed in topic does not match handler")

func topics(handler []ContractEventHandler) ([][]common.Hash, error) {
	var query [][]any
	for _, h := range handler {
		// Append the event selector to the query parameters and construct the topic set
		query = append([][]any{{h.Topic()}}, query...)
	}
	topics, err := abi.MakeTopics(query...)
	if err != nil {
		return nil, err
	}
	return topics, nil
}

// UnpackLog unpacks a retrieved log into the provided output structure.
func UnpackLog(a abi.ABI, out interface{}, event string, log types.Log) error {
	// Copy of bind.BoundContract.UnpackLog

	// Anonymous events are not supported.
	if len(log.Topics) == 0 {
		return errTopicMismatch
	}
	if log.Topics[0] != a.Events[event].ID {
		return errTopicMismatch
	}
	if len(log.Data) > 0 {
		if err := a.UnpackIntoInterface(out, event, log.Data); err != nil {
			return fmt.Errorf("error marshaling into `out` value: %w", err)
		}
	}
	var indexed abi.Arguments
	for _, arg := range a.Events[event].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return abi.ParseTopics(out, indexed, log.Topics[1:])
}
