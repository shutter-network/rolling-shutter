package shutterservice

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

const (
	Word    = 32
	Version = 0x1
)

// EventTriggerDefinition specifies an event-based trigger.
type EventTriggerDefinition struct {
	Contract      common.Address
	LogPredicates []LogPredicate
}

// LogPredicate defines a condition on the events emitted by a contract that must be satisfied for a
// corresponding event trigger to fire.
type LogPredicate struct {
	LogValueRef    LogValueRef
	ValuePredicate ValuePredicate
}

// LogValueRef references a value contained in an event log.
//   - If 0 <= Offset < 4, it refers to the topic of the log at index Offset. In this case, Length
//     must be 1.
//   - If Offset >= 4, it refers to a slice of 32-byte words from the log's data. Its start index is
//     Offset - 4 and the length is Length. E.g., for offset 5 and length 2, the slice starts at
//     byte 32, ends at byte 96 (exclusive), and is 64 bytes long.
type LogValueRef struct {
	Offset uint64
	Length uint64
}

// ValuePredicate defines a condition on a value contained in an event log that must be satisfied
// for a corresponding event trigger to fire. It consists of an operation and a set of arguments,
// e.g. `<` and `100` for a predicate that checks if a value is less than 100. The type and number
// of arguments that are required depend on the operation.
type ValuePredicate struct {
	Op       Op
	IntArgs  []*big.Int
	ByteArgs [][]byte
}

// Op enumerates the operation to be performed when evaluating a constraint.
type Op uint64

func (d *EventTriggerDefinition) UnmarshalBytes(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("data is empty")
	}
	version := data[0]
	if version != Version {
		return fmt.Errorf("unsupported version %d, expected %d", version, Version)
	}

	if err := rlp.DecodeBytes(data[1:], d); err != nil {
		return fmt.Errorf("failed to decode EventTriggerDefinitionRLP: %w", err)
	}
	if err := d.Validate(); err != nil {
		return fmt.Errorf("invalid EventTriggerDefinitionRLP: %w", err)
	}
	return nil
}

func (d *EventTriggerDefinition) MarshalBytes() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(Version)
	if err := rlp.Encode(&buf, d); err != nil {
		return nil, fmt.Errorf("failed to encode EventTriggerDefinitionRLP: %w", err)
	}
	return buf.Bytes(), nil
}

// Validate checks if the event trigger definition is valid.
//
// A trigger definition is valid if
//   - all log predicates are valid and
//   - there are no two log BytesEq predicates for the same topic
func (d *EventTriggerDefinition) Validate() error {
	for i, lp := range d.LogPredicates {
		if err := lp.Validate(); err != nil {
			return fmt.Errorf("invalid log predicate at index %d: %w", i, err)
		}
	}

	topicMap := make(map[uint64]struct{})
	for i, lp := range d.LogPredicates {
		if !lp.LogValueRef.IsTopic() || lp.ValuePredicate.Op != BytesEq {
			continue
		}
		if _, exists := topicMap[lp.LogValueRef.Offset]; exists {
			return fmt.Errorf("duplicate BytesEq log predicate for topic %d at index %d", lp.LogValueRef.Offset, i)
		}
		topicMap[lp.LogValueRef.Offset] = struct{}{}
	}

	return nil
}

// ToFilterQuery creates an Ethereum filter query based on the event trigger definition.
//
// The returned filter includes:
//   - Contract address filtering: Only events from the specified contract are matched
//   - Topic filtering: BytesEq operations on topics are converted to topic filters
//
// Any other operation is not included in the filter and must be checked by the caller.
//
// The method returns an error if
//   - there are multiple BytesEq log predicates for the same topic
//   - the argument for a topic BytesEq log predicate is not a 32-byte value
//
// These errors do not occur if Validate passes.
func (d *EventTriggerDefinition) ToFilterQuery() (ethereum.FilterQuery, error) {
	topics := [][]common.Hash{}
	for _, logPredicate := range d.LogPredicates {
		if !logPredicate.LogValueRef.IsTopic() {
			continue
		}
		if logPredicate.ValuePredicate.Op != BytesEq {
			continue
		}

		topicIndex := logPredicate.LogValueRef.Offset
		for uint64(len(topics)) <= topicIndex {
			topics = append(topics, []common.Hash{})
		}
		if len(topics[topicIndex]) != 0 {
			return ethereum.FilterQuery{}, fmt.Errorf("multiple log predicates for topic %d", topicIndex)
		}
		topic := logPredicate.ValuePredicate.ByteArgs[0]
		if len(topic) != Word {
			return ethereum.FilterQuery{}, fmt.Errorf("log predicate for topic %d must have a 32-byte value, got %d bytes", topicIndex, len(topic))
		}
		topics[logPredicate.LogValueRef.Offset] = []common.Hash{common.BytesToHash(topic)}
	}
	return ethereum.FilterQuery{
		BlockHash: nil,
		FromBlock: nil,
		ToBlock:   nil,
		Addresses: []common.Address{d.Contract},
		Topics:    topics,
	}, nil
}

// Match checks if the log matches the event trigger definition by checking all log predicates.
//
// This may panic if Validate does not pass.
func (d *EventTriggerDefinition) Match(log *types.Log) (bool, error) {
	if log.Address != d.Contract {
		return false, nil
	}
	for _, logPredicate := range d.LogPredicates {
		match, err := logPredicate.Match(log)
		if err != nil {
			return false, err
		}
		if !match {
			return false, nil
		}
	}
	return true, nil
}

func (p *LogPredicate) Validate() error {
	if err := p.LogValueRef.Validate(); err != nil {
		return err
	}
	if err := p.ValuePredicate.Validate(p.LogValueRef.Length); err != nil {
		return err
	}
	return nil
}

func (p *LogPredicate) Match(log *types.Log) (bool, error) {
	value := p.LogValueRef.GetValue(log)
	return p.ValuePredicate.Match(value)
}

func (r *LogValueRef) Validate() error {
	if r.Length == 0 {
		return fmt.Errorf("log value reference length must be positive, got %d", r.Length)
	}
	if r.Offset < 4 && r.Length != 1 {
		return fmt.Errorf("log value reference offset < 4 requires length to be 1, got %d", r.Length)
	}
	// Check that the offset and length are within reasonable bounds so that we can convert them
	// to bytes and bits without worrying.
	if r.Offset > math.MaxUint32 {
		return fmt.Errorf("log value reference offset must be less than 2^32, got %d", r.Offset)
	}
	if r.Length > math.MaxUint32 {
		return fmt.Errorf("log value reference length must be less than 2^32, got %d", r.Length)
	}
	return nil
}

func (r *LogValueRef) EncodeRLP(w io.Writer) error {
	buf := rlp.NewEncoderBuffer(w)
	if r.Length == 1 {
		buf.WriteUint64(r.Offset)
	} else {
		listIndex := buf.List()
		buf.WriteUint64(r.Offset)
		buf.WriteUint64(r.Length)
		buf.ListEnd(listIndex)
	}
	return buf.Flush()
}

func (r *LogValueRef) DecodeRLP(s *rlp.Stream) error {
	var offset, length uint64
	kind, _, err := s.Kind()
	if err != nil {
		return fmt.Errorf("failed to decode LogValueRef: %w", err)
	}
	switch kind {
	case rlp.Byte, rlp.String:
		offset, err = s.Uint64()
		if err != nil {
			return fmt.Errorf("failed to read offset from LogValueRef: %w", err)
		}
		length = 1
	case rlp.List:
		_, err = s.List()
		if err != nil {
			return fmt.Errorf("failed to read LogValueRef list: %w", err)
		}
		offset, err = s.Uint64()
		if err != nil {
			return fmt.Errorf("failed to read offset from LogValueRef: %w", err)
		}
		length, err = s.Uint64()
		if err != nil {
			return fmt.Errorf("failed to read length from LogValueRef: %w", err)
		}
		err = s.ListEnd()
		if err != nil {
			return fmt.Errorf("failed to decode LogValueRef: %w", err)
		}
	default:
		panic(fmt.Sprintf("unexpected kind %d for LogValueRef", kind))
	}
	r.Offset = offset
	r.Length = length
	if err := r.Validate(); err != nil {
		return fmt.Errorf("invalid LogValueRef: %w", err)
	}
	return nil
}

func (r *LogValueRef) IsTopic() bool {
	return r.Offset < 4
}

// GetValue retrieves the value from the log based on the LogValueRef.
//
// In case a slice of log data is referenced and the slice exceeds the log's data length, the
// result will be zero-padded on the right to the expected length.
func (r *LogValueRef) GetValue(log *types.Log) []byte {
	if r.IsTopic() {
		if uint64(len(log.Topics)) <= r.Offset {
			return nil
		}
		return log.Topics[r.Offset].Bytes()
	}

	dataOffset := r.Offset - 4
	value := make([]byte, r.Length*Word)

	startByte := dataOffset * Word
	endByte := (dataOffset + r.Length) * Word

	if startByte < uint64(len(log.Data)) {
		availableEnd := uint64(len(log.Data))
		if endByte < availableEnd {
			availableEnd = endByte
		}
		copy(value, log.Data[startByte:availableEnd])
	}

	return value
}

const (
	UintLt Op = iota
	UintLte
	UintEq
	UintGt
	UintGte
	BytesEq
)

func (op Op) Validate() error {
	switch op {
	case UintLt, UintLte, UintEq, UintGt, UintGte, BytesEq:
		return nil
	default:
		return fmt.Errorf("invalid operation: %d", op)
	}
}

func (op Op) NumIntArgs() int {
	switch op {
	case UintLt, UintLte, UintEq, UintGt, UintGte:
		return 1
	case BytesEq:
		return 0
	default:
		return 0
	}
}

func (op Op) NumByteArgs() int {
	switch op {
	case UintLt, UintLte, UintEq, UintGt, UintGte:
		return 0
	case BytesEq:
		return 1
	default:
		return 0
	}
}

func (p *ValuePredicate) EncodeRLP(w io.Writer) error {
	var elements []interface{}
	elements = append(elements, uint64(p.Op))
	for _, intArg := range p.IntArgs {
		elements = append(elements, intArg)
	}
	for _, byteArg := range p.ByteArgs {
		elements = append(elements, byteArg)
	}
	return rlp.Encode(w, elements)
}

func (p *ValuePredicate) DecodeRLP(s *rlp.Stream) error {
	_, err := s.List()
	if err != nil {
		return fmt.Errorf("failed to decode ValuePredicate: %w", err)
	}
	opInt, err := s.Uint64()
	if err != nil {
		return fmt.Errorf("failed to read operation from ValuePredicate: %w", err)
	}
	op := Op(opInt)
	if err := op.Validate(); err != nil {
		return fmt.Errorf("invalid operation: %w", err)
	}

	intArgs := []*big.Int{}
	for i := 0; i < op.NumIntArgs(); i++ {
		intArg, err := s.BigInt()
		if err != nil {
			return fmt.Errorf("failed to read integer argument %d: %w", i, err)
		}
		intArgs = append(intArgs, intArg)
	}

	byteArgs := [][]byte{}
	for i := 0; i < op.NumByteArgs(); i++ {
		byteArg, err := s.Bytes()
		if err != nil {
			return fmt.Errorf("failed to read byte argument %d: %w", i, err)
		}
		byteArgs = append(byteArgs, byteArg)
	}

	err = s.ListEnd()
	if err != nil {
		return fmt.Errorf("failed to decode ValuePredicate: %w", err)
	}

	p.Op = op
	p.IntArgs = intArgs
	p.ByteArgs = byteArgs

	return nil
}

func (p *ValuePredicate) Validate(numWords uint64) error {
	if err := p.Op.Validate(); err != nil {
		return err
	}
	if err := p.validateArgNums(); err != nil {
		return err
	}
	if err := p.validateArgValues(numWords); err != nil {
		return err
	}
	return nil
}

func (p *ValuePredicate) validateArgNums() error {
	requiredIntArgs := p.Op.NumIntArgs()
	requiredByteArgs := p.Op.NumByteArgs()

	if len(p.IntArgs) != requiredIntArgs {
		return fmt.Errorf("operation %d requires exactly %d integer argument(s), got %d", p.Op, requiredIntArgs, len(p.IntArgs))
	}
	if len(p.ByteArgs) != requiredByteArgs {
		return fmt.Errorf("operation %d requires exactly %d bytes argument(s), got %d", p.Op, requiredByteArgs, len(p.ByteArgs))
	}
	return nil
}

func (p *ValuePredicate) validateArgValues(numWords uint64) error {
	for i, arg := range p.IntArgs {
		if arg == nil {
			return fmt.Errorf("integer argument %d cannot be nil for operation %d", i, p.Op)
		}
		if arg.Sign() < 0 {
			return fmt.Errorf("integer argument %d cannot be negative for operation %d", i, p.Op)
		}
		if uint64(arg.BitLen()) > numWords*Word*8 {
			return fmt.Errorf(
				"bit length of integer argument %d cannot exceed value bit length %d for operation %d, got %d bits",
				i, numWords*Word*8, p.Op, arg.BitLen(),
			)
		}
	}
	for i, arg := range p.ByteArgs {
		if uint64(len(arg)) != numWords*Word {
			return fmt.Errorf(
				"size of byte argument %d must match size of value (%d bytes) for operation %d, got %d bytes",
				i, numWords*Word, p.Op, len(arg),
			)
		}
	}
	return nil
}

func (p *ValuePredicate) Match(value []byte) (bool, error) {
	n := new(big.Int).SetBytes(value)
	switch p.Op {
	case UintLt:
		return n.Cmp(p.IntArgs[0]) < 0, nil
	case UintLte:
		return n.Cmp(p.IntArgs[0]) <= 0, nil
	case UintEq:
		return n.Cmp(p.IntArgs[0]) == 0, nil
	case UintGt:
		return n.Cmp(p.IntArgs[0]) > 0, nil
	case UintGte:
		return n.Cmp(p.IntArgs[0]) >= 0, nil
	case BytesEq:
		return bytes.Equal(value, p.ByteArgs[0]), nil
	}
	return false, fmt.Errorf("unknown operation %d", p.Op)
}
