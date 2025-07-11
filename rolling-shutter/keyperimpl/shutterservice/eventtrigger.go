package shutterservice

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// # Trigger definition
// ## Comparison operators
// "eq", "lt", "lte", "gt", "gte" for number types, "match" for string/bytes
//
// ## ABI-Event:
//
//	{
//	 "contract": "0xdead..beef",
//		"signature": "Transfer(from address indexed, to address indexed, amount uint256)",
//		"conditions": [
//			{"to": {"match": "0xdead...beef"}},
//			{"amount": {"gte": 1}}
//		],
//	}
//
// Note: fields that are not referenced in "conditions" are not restricted.
//
// ## RAW-Event:
// A user may not have the event-ABI available, or may not want to share it.
//
//	{
//	 "contract": "0xdead..beef",
//	 "rawsig": "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
//	 "rawconditions": [
//	   	{"topic1": "any"},
//	 	{"topic2": {"match": "0xdead..beef"}},
//	 	{"data": {
//	 		"start": 0,
//			"end": 32, // probably unnecessary, can be derived from "cast" type
//			"cast": "uint256",
//			"gte": 1
//			}
//		},
//	 ],
//	}
//
// Note: in order to allow for successful matching/parsing, _all_ "topics" must be referenced -- "any" allows for no restrictions.
//
// ## Condensed encoding (WIP)
//
// We need this condensed encoding for registering trigger conditions on the blockchain (most likely as an event......)
// [-1:0] version byte  // Note: all byte numbers need to shift 1 to the right, to have the version in there...
// [0:32] address
// [33:64] topic0/raw signature
// [65] OPCODE-MATCH (see event_triggers.py)
// [66:matching_topics_number*32] matching hashes for topics
// [*:end] DATA matches
// Encoding for DATA matches:
// [*:2] offset
// [3] cast-matchtype-size {0: bytes32-match, 1: uint256-lt, 2: uint256-lte, 3: uint256-eq, 4: uint256-gte, 5:uint256-gt}
// [4-36] matchdata
// [$repeat for all data field conditions]

type EventTriggerDefition struct {
	Contract   common.Address
	Signature  string
	Conditions []Condition
}

type Condition struct {
	FieldName  string
	Constraint Constraint
}

type Constraint interface {
	Test(any) bool
}

type Operator int

const (
	lt Operator = iota
	lte
	eq
	gte
	gt
)

var operatorSymbol = map[Operator]string{
	lt:  "<",
	lte: "<=",
	eq:  "==",
	gte: "=>",
	gt:  ">",
}

type NumericConstraint struct {
	operator Operator
	value    *big.Int
}

func (n *NumericConstraint) Test(target *big.Int) bool {
	switch n.operator {
	case lt:
		return n.value.Cmp(target) < 0
	case lte:
		return n.value.Cmp(target) < 1
	case eq:
		return n.value.Cmp(target) == 0
	case gte:
		return n.value.Cmp(target) > -1
	case gt:
		return n.value.Cmp(target) > 0
	default:
		return false
	}
}

type MatchConstraint struct {
	value [32]byte
}

func (m *MatchConstraint) Test(target [32]byte) bool {
	for i := range m.value {
		if target[i] != m.value[i] {
			return false
		}
	}
	return true
}
