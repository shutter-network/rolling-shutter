package shutterservice

import (
	"fmt"
	"math/big"
	"strings"
	"text/scanner"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// # Trigger definition
// ## Comparison operators
// "eq", "lt", "lte", "gt", "gte" for number types, "match" for string/bytes
//
// ## ABI-Event:
//
//	{
//	 "contract": "0xdead..beef",
//		"signature": "Transfer(address from indexed, address to indexed, uint256 amount)",
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
// [0] version byte
// [1:33] address
// [33:65] topic0/raw signature
// [65] OPCODE-MATCH (see event_triggers.py)
// [66:matching_topics_number*32] matching hashes for topics
// [*:end] DATA matches
// Encoding for DATA matches:
// [*:2] offset
// [3] cast-matchtype-size {0: uint256-lt, 1: uint256-lte, 2: uint256-eq, 3: uint256-gte, 4:uint256-gt, 5: byte32-match, X with X>32: [X]byte-match}
// [4:4+32] matchdata for 1 word matches OR
// [4:4+X] matchdata for [X]byte-match
// [$repeat for all data field conditions]

type EventTriggerDefinition struct {
	Contract   common.Address
	Signature  EvtSignature
	Conditions []Condition
}

func (e EventTriggerDefinition) ToFilterQuery() ethereum.FilterQuery {
	topics := [][]common.Hash{
		{e.Signature.Topic0()},
	}
	for _, cond := range e.Conditions {
		switch cond.Location.(type) {
		case TopicData:
			d, ok := cond.Constraint.(MatchConstraint)
			if !ok {
				continue
			}
			topics = append(topics, []common.Hash{common.Hash(d.target)})
		default:
			continue
		}
	}

	query := ethereum.FilterQuery{
		BlockHash: nil,
		FromBlock: nil,
		ToBlock:   nil,
		Addresses: []common.Address{e.Contract},
		Topics:    topics,
	}
	return query
}

func (e *EventTriggerDefinition) Match(elog types.Log, testTopics bool) bool {
	for _, c := range e.Conditions {
		switch c.Location.(type) {
		case TopicData:
			continue
		default:
			if !c.Fullfilled(elog) {
				return false
			}
		}
	}
	return true
}

type EvtSignature string

func (e EvtSignature) ToHashableSig() string {
	var name string
	var prev string
	i := 0
	args := make([]string, 1+strings.Count(string(e), ","))
	var s scanner.Scanner
	s.Init(strings.NewReader(string(e)))
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		if s.Position.Offset == 0 {
			name = s.TokenText()
		} else {
			if prev == "(" || prev == "," {
				args[i] = s.TokenText()
				i++
			}
			prev = s.TokenText()
		}
	}
	result := fmt.Sprintf("%v(%v)", name, strings.Join(args, ","))
	return result
}

func (e EvtSignature) Topic0() common.Hash {
	shortSig := e.ToHashableSig()

	return common.Hash(crypto.Keccak256([]byte(shortSig)))
}

type LogField interface {
	String() string
	GetSlice(l types.Log) []byte
}

type TopicData struct {
	number int
}

func (t TopicData) String() string {
	return fmt.Sprintf("topic%v", t.number)
}

func (t TopicData) GetSlice(l types.Log) []byte {
	return l.Topics[t.number].Bytes()
}

type OffsetData struct {
	start int
	len   int
}

func (o OffsetData) String() string {
	return fmt.Sprintf("[%v:%v]", o.start, o.start+o.len)
}

func (o OffsetData) GetSlice(l types.Log) []byte {
	// if size > 32:
	// find offset at described position
	// return slice from complex data
	// see https://learnevm.com/chapters/abi-encoding/anatomy#data-type-breakdown
	// else
	// return slice by offset/start + len
	if o.len > 32 {
		slice := l.Data[o.start : o.start+32]
		offset := big.NewInt(0).SetBytes(slice).Int64()
		size := big.NewInt(0).SetBytes(l.Data[offset : offset+32]).Int64()
		return l.Data[offset+32 : offset+32+size]
	}
	slice := l.Data[o.start : o.start+o.len]
	return slice
}

// TODO: TopicData can ONLY allow for MatchConstraint (due to values getting hashed into topics, so no numeric comparison possible)
type Condition struct {
	Location   LogField
	Constraint Constraint
}

func (c *Condition) Fullfilled(elog types.Log) bool {
	val := c.Location.GetSlice(elog)
	switch c.Constraint.(type) {
	case NumConstraint:
		num := c.Constraint.(NumConstraint)
		return num.Test(num.GetValue(val))
	case MatchConstraint:
		match := c.Constraint.(MatchConstraint)
		return match.Test(match.GetValue(val))
	default:
		return false
	}
}

type Constraint interface {
	Test(t any) bool
}

type Op int

const (
	LT Op = iota
	LTE
	EQ
	GTE
	GT
)

var operatorSymbol = map[Op]string{
	LT:  "<",
	LTE: "<=",
	EQ:  "==",
	GTE: "=>",
	GT:  ">",
}

type NumConstraint struct {
	op     Op
	target *big.Int
}

func (n NumConstraint) Test(v any) bool {
	value, ok := v.(*big.Int)
	if !ok {
		return false
	}
	switch n.op {
	case LT:
		return n.target.Cmp(value) > 0
	case LTE:
		return n.target.Cmp(value) >= 0
	case EQ:
		return n.target.Cmp(value) == 0
	case GTE:
		return n.target.Cmp(value) <= 0
	case GT:
		return n.target.Cmp(value) < 0
	default:
		return false
	}
}

func (n NumConstraint) GetValue(slice []byte) *big.Int {
	r := big.NewInt(0)
	result := r.SetBytes(slice)
	return result
}

type MatchConstraint struct {
	target []byte
}

func (m MatchConstraint) GetValue(slice []byte) []byte {
	return slice
}

func (m MatchConstraint) Test(v any) bool {
	value, ok := v.([]byte)
	if !ok {
		return false
	}
	for i := range m.target {
		if value[i] != m.target[i] {
			return false
		}
	}
	return true
}

func TopicPad(data []byte) []byte {
	out := make([]byte, 32)
	copy(out[32-len(data):], data)
	return out
}
