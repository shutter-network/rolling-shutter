package shutterservice

import (
	"fmt"
	"math/big"
	"strings"
	"text/scanner"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
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

type EventTriggerDefinition struct {
	Contract   common.Address
	Signature  EvtSignature
	Conditions []Condition
}

func (e EventTriggerDefinition) ToFilterQuery() ethereum.FilterQuery {
	topics := [][]common.Hash{
		{e.Signature.Topic0()},
		// {common.Hash(topic1)},
		// {common.Hash(topic2)},
	}

	query := ethereum.FilterQuery{
		BlockHash: nil,
		FromBlock: nil,
		ToBlock:   nil,
		Addresses: []common.Address{e.Contract},
		Topics:    topics,
	}
	return query
	/*
		topics := [][]common.Hash{
			{common.Hash(topic0)},
		}

		query := ethereum.FilterQuery{
			BlockHash: nil,
			FromBlock: big.NewInt(int64(latest)),
			ToBlock:   nil,
			Addresses: []common.Address{setup.contractAddress},
			Topics:    topics,
		}
	*/
}

type EvtSignature string

func (e EvtSignature) ToHashableSig() string {
	var name string
	var prev string
	i := 0
	args := make([]string, 3)
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
	fmt.Printf(result)
	return result
}

func (e EvtSignature) Topic0() common.Hash {
	shortSig := e.ToHashableSig()

	return common.Hash(crypto.Keccak256([]byte(shortSig)))
}

type Position interface {
	String() string
}

type Topic struct {
	number int
}

func (t Topic) String() string {
	return fmt.Sprintf("topic%v", t.number)
}

type OffsetData struct {
	start int
	len   int
}

func (o OffsetData) String() string {
	return fmt.Sprintf("[%v:%v]", o.start, o.start+o.len)
}

type Condition struct {
	Location   Position // FieldNames do NOT exist in the log context! only topic{N} and data[offset]
	Constraint Constraint
}

type Constraint interface {
	Test(T any) bool
}

type Op int

const (
	lt Op = iota
	lte
	eq
	gte
	gt
)

var operatorSymbol = map[Op]string{
	lt:  "<",
	lte: "<=",
	eq:  "==",
	gte: "=>",
	gt:  ">",
}

type NumConstraint struct {
	op  Op
	val *big.Int
}

func (n NumConstraint) Test(t any) bool {
	target, ok := t.(*big.Int)
	if !ok {
		return false
	}
	switch n.op {
	case lt:
		return n.val.Cmp(target) < 0
	case lte:
		return n.val.Cmp(target) < 1
	case eq:
		return n.val.Cmp(target) == 0
	case gte:
		return n.val.Cmp(target) > -1
	case gt:
		return n.val.Cmp(target) > 0
	default:
		return false
	}
}

type MatchConstraint struct {
	val [32]byte
}

func (m MatchConstraint) Test(t any) bool {
	target, ok := t.([32]byte)
	if !ok {
		return false
	}
	for i := range m.val {
		if target[i] != m.val[i] {
			return false
		}
	}
	return true
}
