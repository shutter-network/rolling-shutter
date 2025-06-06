// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v5.29.3
// source: shmsg.proto

package shmsg

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type BatchConfig struct {
	state                 protoimpl.MessageState `protogen:"open.v1"`
	ActivationBlockNumber uint64                 `protobuf:"varint,1,opt,name=activation_block_number,json=activationBlockNumber,proto3" json:"activation_block_number,omitempty"`
	Keypers               [][]byte               `protobuf:"bytes,2,rep,name=keypers,proto3" json:"keypers,omitempty"`
	Threshold             uint64                 `protobuf:"varint,3,opt,name=threshold,proto3" json:"threshold,omitempty"`
	KeyperConfigIndex     uint64                 `protobuf:"varint,5,opt,name=keyper_config_index,json=keyperConfigIndex,proto3" json:"keyper_config_index,omitempty"`
	unknownFields         protoimpl.UnknownFields
	sizeCache             protoimpl.SizeCache
}

func (x *BatchConfig) Reset() {
	*x = BatchConfig{}
	mi := &file_shmsg_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BatchConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BatchConfig) ProtoMessage() {}

func (x *BatchConfig) ProtoReflect() protoreflect.Message {
	mi := &file_shmsg_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BatchConfig.ProtoReflect.Descriptor instead.
func (*BatchConfig) Descriptor() ([]byte, []int) {
	return file_shmsg_proto_rawDescGZIP(), []int{0}
}

func (x *BatchConfig) GetActivationBlockNumber() uint64 {
	if x != nil {
		return x.ActivationBlockNumber
	}
	return 0
}

func (x *BatchConfig) GetKeypers() [][]byte {
	if x != nil {
		return x.Keypers
	}
	return nil
}

func (x *BatchConfig) GetThreshold() uint64 {
	if x != nil {
		return x.Threshold
	}
	return 0
}

func (x *BatchConfig) GetKeyperConfigIndex() uint64 {
	if x != nil {
		return x.KeyperConfigIndex
	}
	return 0
}

type BlockSeen struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	BlockNumber   uint64                 `protobuf:"varint,1,opt,name=block_number,json=blockNumber,proto3" json:"block_number,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *BlockSeen) Reset() {
	*x = BlockSeen{}
	mi := &file_shmsg_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BlockSeen) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockSeen) ProtoMessage() {}

func (x *BlockSeen) ProtoReflect() protoreflect.Message {
	mi := &file_shmsg_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockSeen.ProtoReflect.Descriptor instead.
func (*BlockSeen) Descriptor() ([]byte, []int) {
	return file_shmsg_proto_rawDescGZIP(), []int{1}
}

func (x *BlockSeen) GetBlockNumber() uint64 {
	if x != nil {
		return x.BlockNumber
	}
	return 0
}

type CheckIn struct {
	state               protoimpl.MessageState `protogen:"open.v1"`
	ValidatorPublicKey  []byte                 `protobuf:"bytes,1,opt,name=validator_public_key,json=validatorPublicKey,proto3" json:"validator_public_key,omitempty"`    // 32 byte ed25519 public key
	EncryptionPublicKey []byte                 `protobuf:"bytes,2,opt,name=encryption_public_key,json=encryptionPublicKey,proto3" json:"encryption_public_key,omitempty"` // compressed ecies public key
	unknownFields       protoimpl.UnknownFields
	sizeCache           protoimpl.SizeCache
}

func (x *CheckIn) Reset() {
	*x = CheckIn{}
	mi := &file_shmsg_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CheckIn) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckIn) ProtoMessage() {}

func (x *CheckIn) ProtoReflect() protoreflect.Message {
	mi := &file_shmsg_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckIn.ProtoReflect.Descriptor instead.
func (*CheckIn) Descriptor() ([]byte, []int) {
	return file_shmsg_proto_rawDescGZIP(), []int{2}
}

func (x *CheckIn) GetValidatorPublicKey() []byte {
	if x != nil {
		return x.ValidatorPublicKey
	}
	return nil
}

func (x *CheckIn) GetEncryptionPublicKey() []byte {
	if x != nil {
		return x.EncryptionPublicKey
	}
	return nil
}

type PolyEval struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	Eon            uint64                 `protobuf:"varint,1,opt,name=eon,proto3" json:"eon,omitempty"`
	Receivers      [][]byte               `protobuf:"bytes,2,rep,name=receivers,proto3" json:"receivers,omitempty"`
	EncryptedEvals [][]byte               `protobuf:"bytes,3,rep,name=encrypted_evals,json=encryptedEvals,proto3" json:"encrypted_evals,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *PolyEval) Reset() {
	*x = PolyEval{}
	mi := &file_shmsg_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PolyEval) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PolyEval) ProtoMessage() {}

func (x *PolyEval) ProtoReflect() protoreflect.Message {
	mi := &file_shmsg_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PolyEval.ProtoReflect.Descriptor instead.
func (*PolyEval) Descriptor() ([]byte, []int) {
	return file_shmsg_proto_rawDescGZIP(), []int{3}
}

func (x *PolyEval) GetEon() uint64 {
	if x != nil {
		return x.Eon
	}
	return 0
}

func (x *PolyEval) GetReceivers() [][]byte {
	if x != nil {
		return x.Receivers
	}
	return nil
}

func (x *PolyEval) GetEncryptedEvals() [][]byte {
	if x != nil {
		return x.EncryptedEvals
	}
	return nil
}

type PolyCommitment struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Eon           uint64                 `protobuf:"varint,1,opt,name=eon,proto3" json:"eon,omitempty"`
	Gammas        [][]byte               `protobuf:"bytes,2,rep,name=gammas,proto3" json:"gammas,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PolyCommitment) Reset() {
	*x = PolyCommitment{}
	mi := &file_shmsg_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PolyCommitment) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PolyCommitment) ProtoMessage() {}

func (x *PolyCommitment) ProtoReflect() protoreflect.Message {
	mi := &file_shmsg_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PolyCommitment.ProtoReflect.Descriptor instead.
func (*PolyCommitment) Descriptor() ([]byte, []int) {
	return file_shmsg_proto_rawDescGZIP(), []int{4}
}

func (x *PolyCommitment) GetEon() uint64 {
	if x != nil {
		return x.Eon
	}
	return 0
}

func (x *PolyCommitment) GetGammas() [][]byte {
	if x != nil {
		return x.Gammas
	}
	return nil
}

type Accusation struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Eon           uint64                 `protobuf:"varint,1,opt,name=eon,proto3" json:"eon,omitempty"`
	Accused       [][]byte               `protobuf:"bytes,2,rep,name=accused,proto3" json:"accused,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Accusation) Reset() {
	*x = Accusation{}
	mi := &file_shmsg_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Accusation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Accusation) ProtoMessage() {}

func (x *Accusation) ProtoReflect() protoreflect.Message {
	mi := &file_shmsg_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Accusation.ProtoReflect.Descriptor instead.
func (*Accusation) Descriptor() ([]byte, []int) {
	return file_shmsg_proto_rawDescGZIP(), []int{5}
}

func (x *Accusation) GetEon() uint64 {
	if x != nil {
		return x.Eon
	}
	return 0
}

func (x *Accusation) GetAccused() [][]byte {
	if x != nil {
		return x.Accused
	}
	return nil
}

type Apology struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Eon           uint64                 `protobuf:"varint,1,opt,name=eon,proto3" json:"eon,omitempty"`
	Accusers      [][]byte               `protobuf:"bytes,2,rep,name=accusers,proto3" json:"accusers,omitempty"`
	PolyEvals     [][]byte               `protobuf:"bytes,3,rep,name=poly_evals,json=polyEvals,proto3" json:"poly_evals,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Apology) Reset() {
	*x = Apology{}
	mi := &file_shmsg_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Apology) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Apology) ProtoMessage() {}

func (x *Apology) ProtoReflect() protoreflect.Message {
	mi := &file_shmsg_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Apology.ProtoReflect.Descriptor instead.
func (*Apology) Descriptor() ([]byte, []int) {
	return file_shmsg_proto_rawDescGZIP(), []int{6}
}

func (x *Apology) GetEon() uint64 {
	if x != nil {
		return x.Eon
	}
	return 0
}

func (x *Apology) GetAccusers() [][]byte {
	if x != nil {
		return x.Accusers
	}
	return nil
}

func (x *Apology) GetPolyEvals() [][]byte {
	if x != nil {
		return x.PolyEvals
	}
	return nil
}

// DKGResult is sent by the keyper if the DKG process for an eon has
// finished. The field 'success' is used to signal whether the DKG process was
// successful. If the DKG process fails for a majority of keypers, the
// shuttermint app will restart the DKG process. This replaces the EonStartVote
// the keypers sent previously when the DKG process failed.
type DKGResult struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Eon           uint64                 `protobuf:"varint,2,opt,name=eon,proto3" json:"eon,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DKGResult) Reset() {
	*x = DKGResult{}
	mi := &file_shmsg_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DKGResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DKGResult) ProtoMessage() {}

func (x *DKGResult) ProtoReflect() protoreflect.Message {
	mi := &file_shmsg_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DKGResult.ProtoReflect.Descriptor instead.
func (*DKGResult) Descriptor() ([]byte, []int) {
	return file_shmsg_proto_rawDescGZIP(), []int{7}
}

func (x *DKGResult) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *DKGResult) GetEon() uint64 {
	if x != nil {
		return x.Eon
	}
	return 0
}

type Message struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Payload:
	//
	//	*Message_BatchConfig
	//	*Message_BlockSeen
	//	*Message_CheckIn
	//	*Message_PolyEval
	//	*Message_PolyCommitment
	//	*Message_Accusation
	//	*Message_Apology
	//	*Message_DkgResult
	Payload       isMessage_Payload `protobuf_oneof:"payload"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Message) Reset() {
	*x = Message{}
	mi := &file_shmsg_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_shmsg_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_shmsg_proto_rawDescGZIP(), []int{8}
}

func (x *Message) GetPayload() isMessage_Payload {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *Message) GetBatchConfig() *BatchConfig {
	if x != nil {
		if x, ok := x.Payload.(*Message_BatchConfig); ok {
			return x.BatchConfig
		}
	}
	return nil
}

func (x *Message) GetBlockSeen() *BlockSeen {
	if x != nil {
		if x, ok := x.Payload.(*Message_BlockSeen); ok {
			return x.BlockSeen
		}
	}
	return nil
}

func (x *Message) GetCheckIn() *CheckIn {
	if x != nil {
		if x, ok := x.Payload.(*Message_CheckIn); ok {
			return x.CheckIn
		}
	}
	return nil
}

func (x *Message) GetPolyEval() *PolyEval {
	if x != nil {
		if x, ok := x.Payload.(*Message_PolyEval); ok {
			return x.PolyEval
		}
	}
	return nil
}

func (x *Message) GetPolyCommitment() *PolyCommitment {
	if x != nil {
		if x, ok := x.Payload.(*Message_PolyCommitment); ok {
			return x.PolyCommitment
		}
	}
	return nil
}

func (x *Message) GetAccusation() *Accusation {
	if x != nil {
		if x, ok := x.Payload.(*Message_Accusation); ok {
			return x.Accusation
		}
	}
	return nil
}

func (x *Message) GetApology() *Apology {
	if x != nil {
		if x, ok := x.Payload.(*Message_Apology); ok {
			return x.Apology
		}
	}
	return nil
}

func (x *Message) GetDkgResult() *DKGResult {
	if x != nil {
		if x, ok := x.Payload.(*Message_DkgResult); ok {
			return x.DkgResult
		}
	}
	return nil
}

type isMessage_Payload interface {
	isMessage_Payload()
}

type Message_BatchConfig struct {
	BatchConfig *BatchConfig `protobuf:"bytes,4,opt,name=batch_config,json=batchConfig,proto3,oneof"`
}

type Message_BlockSeen struct {
	// BatchConfigStarted batch_config_started = 6;
	BlockSeen *BlockSeen `protobuf:"bytes,14,opt,name=block_seen,json=blockSeen,proto3,oneof"`
}

type Message_CheckIn struct {
	CheckIn *CheckIn `protobuf:"bytes,7,opt,name=check_in,json=checkIn,proto3,oneof"`
}

type Message_PolyEval struct {
	// DKG messages
	PolyEval *PolyEval `protobuf:"bytes,9,opt,name=poly_eval,json=polyEval,proto3,oneof"`
}

type Message_PolyCommitment struct {
	PolyCommitment *PolyCommitment `protobuf:"bytes,10,opt,name=poly_commitment,json=polyCommitment,proto3,oneof"`
}

type Message_Accusation struct {
	Accusation *Accusation `protobuf:"bytes,11,opt,name=accusation,proto3,oneof"`
}

type Message_Apology struct {
	Apology *Apology `protobuf:"bytes,12,opt,name=apology,proto3,oneof"`
}

type Message_DkgResult struct {
	DkgResult *DKGResult `protobuf:"bytes,15,opt,name=dkg_result,json=dkgResult,proto3,oneof"` // EonStartVote eon_start_vote = 13;
}

func (*Message_BatchConfig) isMessage_Payload() {}

func (*Message_BlockSeen) isMessage_Payload() {}

func (*Message_CheckIn) isMessage_Payload() {}

func (*Message_PolyEval) isMessage_Payload() {}

func (*Message_PolyCommitment) isMessage_Payload() {}

func (*Message_Accusation) isMessage_Payload() {}

func (*Message_Apology) isMessage_Payload() {}

func (*Message_DkgResult) isMessage_Payload() {}

type MessageWithNonce struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Msg           *Message               `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
	ChainId       []byte                 `protobuf:"bytes,2,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	RandomNonce   uint64                 `protobuf:"varint,3,opt,name=random_nonce,json=randomNonce,proto3" json:"random_nonce,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *MessageWithNonce) Reset() {
	*x = MessageWithNonce{}
	mi := &file_shmsg_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MessageWithNonce) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageWithNonce) ProtoMessage() {}

func (x *MessageWithNonce) ProtoReflect() protoreflect.Message {
	mi := &file_shmsg_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageWithNonce.ProtoReflect.Descriptor instead.
func (*MessageWithNonce) Descriptor() ([]byte, []int) {
	return file_shmsg_proto_rawDescGZIP(), []int{9}
}

func (x *MessageWithNonce) GetMsg() *Message {
	if x != nil {
		return x.Msg
	}
	return nil
}

func (x *MessageWithNonce) GetChainId() []byte {
	if x != nil {
		return x.ChainId
	}
	return nil
}

func (x *MessageWithNonce) GetRandomNonce() uint64 {
	if x != nil {
		return x.RandomNonce
	}
	return 0
}

var File_shmsg_proto protoreflect.FileDescriptor

var file_shmsg_proto_rawDesc = string([]byte{
	0x0a, 0x0b, 0x73, 0x68, 0x6d, 0x73, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x73,
	0x68, 0x6d, 0x73, 0x67, 0x22, 0xad, 0x01, 0x0a, 0x0b, 0x42, 0x61, 0x74, 0x63, 0x68, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x12, 0x36, 0x0a, 0x17, 0x61, 0x63, 0x74, 0x69, 0x76, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x15, 0x61, 0x63, 0x74, 0x69, 0x76, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07,
	0x6b, 0x65, 0x79, 0x70, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x07, 0x6b,
	0x65, 0x79, 0x70, 0x65, 0x72, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x68, 0x72, 0x65, 0x73, 0x68,
	0x6f, 0x6c, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x74, 0x68, 0x72, 0x65, 0x73,
	0x68, 0x6f, 0x6c, 0x64, 0x12, 0x2e, 0x0a, 0x13, 0x6b, 0x65, 0x79, 0x70, 0x65, 0x72, 0x5f, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x11, 0x6b, 0x65, 0x79, 0x70, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x49,
	0x6e, 0x64, 0x65, 0x78, 0x22, 0x2e, 0x0a, 0x09, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x53, 0x65, 0x65,
	0x6e, 0x12, 0x21, 0x0a, 0x0c, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65,
	0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75,
	0x6d, 0x62, 0x65, 0x72, 0x22, 0x6f, 0x0a, 0x07, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x49, 0x6e, 0x12,
	0x30, 0x0a, 0x14, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x70, 0x75, 0x62,
	0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x12, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65,
	0x79, 0x12, 0x32, 0x0a, 0x15, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x13, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x75, 0x62, 0x6c,
	0x69, 0x63, 0x4b, 0x65, 0x79, 0x22, 0x63, 0x0a, 0x08, 0x50, 0x6f, 0x6c, 0x79, 0x45, 0x76, 0x61,
	0x6c, 0x12, 0x10, 0x0a, 0x03, 0x65, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03,
	0x65, 0x6f, 0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x73,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x09, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72,
	0x73, 0x12, 0x27, 0x0a, 0x0f, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x65, 0x64, 0x5f, 0x65,
	0x76, 0x61, 0x6c, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x0e, 0x65, 0x6e, 0x63, 0x72,
	0x79, 0x70, 0x74, 0x65, 0x64, 0x45, 0x76, 0x61, 0x6c, 0x73, 0x22, 0x3a, 0x0a, 0x0e, 0x50, 0x6f,
	0x6c, 0x79, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x10, 0x0a, 0x03,
	0x65, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x65, 0x6f, 0x6e, 0x12, 0x16,
	0x0a, 0x06, 0x67, 0x61, 0x6d, 0x6d, 0x61, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x06,
	0x67, 0x61, 0x6d, 0x6d, 0x61, 0x73, 0x22, 0x38, 0x0a, 0x0a, 0x41, 0x63, 0x63, 0x75, 0x73, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x10, 0x0a, 0x03, 0x65, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x03, 0x65, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x63, 0x63, 0x75, 0x73, 0x65,
	0x64, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x07, 0x61, 0x63, 0x63, 0x75, 0x73, 0x65, 0x64,
	0x22, 0x56, 0x0a, 0x07, 0x41, 0x70, 0x6f, 0x6c, 0x6f, 0x67, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x65,
	0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x65, 0x6f, 0x6e, 0x12, 0x1a, 0x0a,
	0x08, 0x61, 0x63, 0x63, 0x75, 0x73, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52,
	0x08, 0x61, 0x63, 0x63, 0x75, 0x73, 0x65, 0x72, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x6f, 0x6c,
	0x79, 0x5f, 0x65, 0x76, 0x61, 0x6c, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x09, 0x70,
	0x6f, 0x6c, 0x79, 0x45, 0x76, 0x61, 0x6c, 0x73, 0x22, 0x37, 0x0a, 0x09, 0x44, 0x4b, 0x47, 0x52,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12,
	0x10, 0x0a, 0x03, 0x65, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x65, 0x6f,
	0x6e, 0x22, 0xb3, 0x03, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x37, 0x0a,
	0x0c, 0x62, 0x61, 0x74, 0x63, 0x68, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x73, 0x68, 0x6d, 0x73, 0x67, 0x2e, 0x42, 0x61, 0x74, 0x63,
	0x68, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x48, 0x00, 0x52, 0x0b, 0x62, 0x61, 0x74, 0x63, 0x68,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x31, 0x0a, 0x0a, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f,
	0x73, 0x65, 0x65, 0x6e, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x73, 0x68, 0x6d,
	0x73, 0x67, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x53, 0x65, 0x65, 0x6e, 0x48, 0x00, 0x52, 0x09,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x53, 0x65, 0x65, 0x6e, 0x12, 0x2b, 0x0a, 0x08, 0x63, 0x68, 0x65,
	0x63, 0x6b, 0x5f, 0x69, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x73, 0x68,
	0x6d, 0x73, 0x67, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x49, 0x6e, 0x48, 0x00, 0x52, 0x07, 0x63,
	0x68, 0x65, 0x63, 0x6b, 0x49, 0x6e, 0x12, 0x2e, 0x0a, 0x09, 0x70, 0x6f, 0x6c, 0x79, 0x5f, 0x65,
	0x76, 0x61, 0x6c, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x73, 0x68, 0x6d, 0x73,
	0x67, 0x2e, 0x50, 0x6f, 0x6c, 0x79, 0x45, 0x76, 0x61, 0x6c, 0x48, 0x00, 0x52, 0x08, 0x70, 0x6f,
	0x6c, 0x79, 0x45, 0x76, 0x61, 0x6c, 0x12, 0x40, 0x0a, 0x0f, 0x70, 0x6f, 0x6c, 0x79, 0x5f, 0x63,
	0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x15, 0x2e, 0x73, 0x68, 0x6d, 0x73, 0x67, 0x2e, 0x50, 0x6f, 0x6c, 0x79, 0x43, 0x6f, 0x6d, 0x6d,
	0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x0e, 0x70, 0x6f, 0x6c, 0x79, 0x43, 0x6f,
	0x6d, 0x6d, 0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x33, 0x0a, 0x0a, 0x61, 0x63, 0x63, 0x75,
	0x73, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x73,
	0x68, 0x6d, 0x73, 0x67, 0x2e, 0x41, 0x63, 0x63, 0x75, 0x73, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x48,
	0x00, 0x52, 0x0a, 0x61, 0x63, 0x63, 0x75, 0x73, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2a, 0x0a,
	0x07, 0x61, 0x70, 0x6f, 0x6c, 0x6f, 0x67, 0x79, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e,
	0x2e, 0x73, 0x68, 0x6d, 0x73, 0x67, 0x2e, 0x41, 0x70, 0x6f, 0x6c, 0x6f, 0x67, 0x79, 0x48, 0x00,
	0x52, 0x07, 0x61, 0x70, 0x6f, 0x6c, 0x6f, 0x67, 0x79, 0x12, 0x31, 0x0a, 0x0a, 0x64, 0x6b, 0x67,
	0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e,
	0x73, 0x68, 0x6d, 0x73, 0x67, 0x2e, 0x44, 0x4b, 0x47, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x48,
	0x00, 0x52, 0x09, 0x64, 0x6b, 0x67, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x42, 0x09, 0x0a, 0x07,
	0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x22, 0x72, 0x0a, 0x10, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x57, 0x69, 0x74, 0x68, 0x4e, 0x6f, 0x6e, 0x63, 0x65, 0x12, 0x20, 0x0a, 0x03, 0x6d,
	0x73, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x73, 0x68, 0x6d, 0x73, 0x67,
	0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x12, 0x19, 0x0a,
	0x08, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x07, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x72, 0x61, 0x6e, 0x64,
	0x6f, 0x6d, 0x5f, 0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b,
	0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x4e, 0x6f, 0x6e, 0x63, 0x65, 0x42, 0x0a, 0x5a, 0x08, 0x2e,
	0x2f, 0x3b, 0x73, 0x68, 0x6d, 0x73, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_shmsg_proto_rawDescOnce sync.Once
	file_shmsg_proto_rawDescData []byte
)

func file_shmsg_proto_rawDescGZIP() []byte {
	file_shmsg_proto_rawDescOnce.Do(func() {
		file_shmsg_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_shmsg_proto_rawDesc), len(file_shmsg_proto_rawDesc)))
	})
	return file_shmsg_proto_rawDescData
}

var file_shmsg_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_shmsg_proto_goTypes = []any{
	(*BatchConfig)(nil),      // 0: shmsg.BatchConfig
	(*BlockSeen)(nil),        // 1: shmsg.BlockSeen
	(*CheckIn)(nil),          // 2: shmsg.CheckIn
	(*PolyEval)(nil),         // 3: shmsg.PolyEval
	(*PolyCommitment)(nil),   // 4: shmsg.PolyCommitment
	(*Accusation)(nil),       // 5: shmsg.Accusation
	(*Apology)(nil),          // 6: shmsg.Apology
	(*DKGResult)(nil),        // 7: shmsg.DKGResult
	(*Message)(nil),          // 8: shmsg.Message
	(*MessageWithNonce)(nil), // 9: shmsg.MessageWithNonce
}
var file_shmsg_proto_depIdxs = []int32{
	0, // 0: shmsg.Message.batch_config:type_name -> shmsg.BatchConfig
	1, // 1: shmsg.Message.block_seen:type_name -> shmsg.BlockSeen
	2, // 2: shmsg.Message.check_in:type_name -> shmsg.CheckIn
	3, // 3: shmsg.Message.poly_eval:type_name -> shmsg.PolyEval
	4, // 4: shmsg.Message.poly_commitment:type_name -> shmsg.PolyCommitment
	5, // 5: shmsg.Message.accusation:type_name -> shmsg.Accusation
	6, // 6: shmsg.Message.apology:type_name -> shmsg.Apology
	7, // 7: shmsg.Message.dkg_result:type_name -> shmsg.DKGResult
	8, // 8: shmsg.MessageWithNonce.msg:type_name -> shmsg.Message
	9, // [9:9] is the sub-list for method output_type
	9, // [9:9] is the sub-list for method input_type
	9, // [9:9] is the sub-list for extension type_name
	9, // [9:9] is the sub-list for extension extendee
	0, // [0:9] is the sub-list for field type_name
}

func init() { file_shmsg_proto_init() }
func file_shmsg_proto_init() {
	if File_shmsg_proto != nil {
		return
	}
	file_shmsg_proto_msgTypes[8].OneofWrappers = []any{
		(*Message_BatchConfig)(nil),
		(*Message_BlockSeen)(nil),
		(*Message_CheckIn)(nil),
		(*Message_PolyEval)(nil),
		(*Message_PolyCommitment)(nil),
		(*Message_Accusation)(nil),
		(*Message_Apology)(nil),
		(*Message_DkgResult)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_shmsg_proto_rawDesc), len(file_shmsg_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_shmsg_proto_goTypes,
		DependencyIndexes: file_shmsg_proto_depIdxs,
		MessageInfos:      file_shmsg_proto_msgTypes,
	}.Build()
	File_shmsg_proto = out.File
	file_shmsg_proto_goTypes = nil
	file_shmsg_proto_depIdxs = nil
}
