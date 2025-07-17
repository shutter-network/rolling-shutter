package shutterservice

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
	"text/scanner"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// ABI encoding word size
const (
	WORD    = 32
	VERSION = 0x1
)

// # Trigger definition
// ## Comparison operators
// "eq", "lt", "lte", "gt", "gte" for number types, "match" for bytes32 and "cmatch" for complex []bytes
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
//	 		"arg": 0,
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
// We need this condensed encoding for registering trigger conditions on the blockchain (most likely as an event......)
//
//
// [0] version byte
// [1:33] address
// [33:65] topic0/raw signature
// [65] OPCODE-MATCH (see event_triggers.py)
// [66:matching_topics_number*32] matching hashes for topics
// [*:end] DATA matches
// Encoding for DATA matches:
// [0] argnumber (note: offset in data ==> argnumber * wordsize; for complex data types, this points to the offset marker in ABI encoding)
// [1] cast-matchtype-size {0: uint256-lt, 1: uint256-lte, 2: uint256-eq, 3: uint256-gte, 4:uint256-gt, 5: byte32-match, 6: []byte-complexmatch}
// [2:2+32] matchdata for 1 word matches OR
// [2:2+X] matchdata for [X]byte-match
// [$repeat for all data field conditions]

type EventTriggerDefinition struct {
	Contract   common.Address
	Signature  EvtSignature
	Conditions []Condition
}

func (e *EventTriggerDefinition) MarshalBytes() [][]byte {
	// write version string to buffer 'work'
	// write common fields to buffer 'work'
	// loop through conditions
	// for TopicData append to work
	// for regular conditions append to buffer 'data'
	// append 'data' to 'work'

	// slice 'work' into [][WORD]array
	var buf []byte
	work := bytes.NewBuffer(buf)
	work.WriteByte(VERSION)
	work.Write(e.Contract[:])
	work.Write(e.Signature.Topic0().Bytes())
	work.WriteByte(e.TopicPattern())
	var d []byte
	data := bytes.NewBuffer(d)
	for _, cond := range e.Conditions {
		switch cond.Location.(type) {
		case TopicData:
			work.Write(cond.Constraint.(MatchConstraint).target)
		case OffsetData:
			data.Write(cond.Bytes())
		}
	}
	work.Write(data.Bytes())
	contents := work.Bytes()
	words := len(contents) / WORD
	if len(contents)%WORD != 0 {
		words++
	}
	target := make([][]byte, words)
	for i := range words {
		target[i] = []byte(contents[i*WORD : (i+1)*WORD])
	}
	return target
}

func (e *EventTriggerDefinition) UnmarshalBytes() error {
	return nil
}

// Topic pattern: we need to specify, how many topics and in which position we reference
/*
	+-+-+-+
	|1|2|3|
	+-+-+-+
	|o|o|o| => 0
	+-+-+-+
	|o|o|x| => 1
	+-+-+-+
	|o|x|o| => 2
	+-+-+-+
	|o|x|x| => 3
	+-+-+-+
	|x|o|o| => 4
	+-+-+-+
	|x|o|x| => 5
	+-+-+-+
	|x|x|o| => 6
	+-+-+-+
	|x|x|x| => 7
	+-+-+-+
*/
func (e *EventTriggerDefinition) TopicPattern() byte {
	v := make([]uint8, 3)
	for _, cond := range e.Conditions {
		switch cond.Location.(type) {
		case TopicData:
			v[cond.Location.(TopicData).number-1] = 1
		default:
			continue
		}
	}
	return v[0] + (v[1] << 1) + (v[2] << 2)
}

func (e EventTriggerDefinition) ToFilterQuery() ethereum.FilterQuery {
	// The Topic list restricts matches to particular event topics. Each event has a list
	// of topics. Topics matches a prefix of that list. An empty element slice matches any
	// topic. Non-empty elements represent an alternative that matches any of the
	// contained topics.
	//
	// Examples:
	// {} or nil          matches any topic list
	// {{A}}              matches topic A in first position
	// {{}, {B}}          matches any topic in first position AND B in second position
	// {{A}, {B}}         matches topic A in first position AND B in second position
	// {{A, B}, {C, D}}   matches topic (A OR B) in first position AND (C OR D) in second position
	// var Topics [][]common.Hash
	topics := [][]common.Hash{
		{e.Signature.Topic0()},
		{},
		{},
		{},
	}
	for _, cond := range e.Conditions {
		switch cond.Location.(type) {
		case TopicData:
			d, ok := cond.Constraint.(MatchConstraint)
			if !ok {
				continue
			}
			topics[cond.Location.(TopicData).number] = []common.Hash{common.Hash(d.target)}
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
	argnumber int
	complex   bool
}

func (o OffsetData) getSliceDef(l types.Log) (offset int64, size int64) {
	start := o.argnumber * WORD
	if o.complex {
		slice := l.Data[start : start+WORD]
		sizeword := big.NewInt(0).SetBytes(slice).Int64()
		offset = sizeword + WORD
		size = big.NewInt(0).SetBytes(l.Data[sizeword : sizeword+WORD]).Int64()
	} else {
		offset = int64(start)
		size = WORD
	}
	return offset, size
}

func (o OffsetData) GetSlice(l types.Log) []byte {
	start, size := o.getSliceDef(l)
	slice := l.Data[start : start+size]
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

// Encoding for DATA matches:
// [0] argnumber (note: offset in data ==> argnumber * wordsize; for complex data types, this points to the offset marker in ABI encoding)
// [1] cast-matchtype-size {0: uint256-lt, 1: uint256-lte, 2: uint256-eq, 3: uint256-gte, 4:uint256-gt, 5: byte32-match, 6: []byte-complexmatch}
// [2:2+32] matchdata for 1 word matches OR
// [3] word count
// [3:3+X] matchdata for [X]byte-match (right padded for [WORD]byte)
func (c *Condition) Bytes() []byte {
	var buf []byte
	data := bytes.NewBuffer(buf)
	switch c.Location.(type) {
	case TopicData:
		data.Write(c.Constraint.(MatchConstraint).target)
	case OffsetData:
		data.WriteByte(byte(c.Location.(OffsetData).argnumber))
		switch c.Constraint.(type) {
		case NumConstraint:
			data.WriteByte(byte(c.Constraint.(NumConstraint).op))
		case MatchConstraint:
			matchBytes := c.Constraint.(MatchConstraint).target
			if c.Location.(OffsetData).complex {
				data.WriteByte(6)
				// align to WORD len and prepend with wordcount
				words := len(matchBytes) / WORD
				if len(matchBytes)%WORD != 0 {
					words++
					data.WriteByte(byte(words))
					data.Write(matchBytes)
					padBytes := make([]byte, len(matchBytes)%WORD)
					fmt.Println(padBytes)
					data.Write(padBytes)
				}
			} else {
				data.WriteByte(5)
				data.Write(matchBytes)
			}
		}
	}
	return data.Bytes()[:]
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
	if len(value) != len(m.target) {
		return false
	}
	for i := range m.target {
		if value[i] != m.target[i] {
			return false
		}
	}
	return true
}

// This kind of padding is only used for topics, which by definition are only 32byte
func TopicPad(data []byte) []byte {
	out := make([]byte, WORD)
	copy(out[WORD-len(data):], data)
	return out
}
