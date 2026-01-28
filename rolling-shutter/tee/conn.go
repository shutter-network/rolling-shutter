package tee

import (
	"bytes"
	"crypto/ecdh"
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net"
	"strings"

	"github.com/edgelesssys/ego/attestation"
	"github.com/edgelesssys/ego/enclave"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

// Communication and guarantees
// Currently there are two different messages we can get (if the "shutter" feature is enabled for the ethereum-enclave):
//
// **VerifiedHeadData** contains information about the latest and finalized (processed) header.
// it has the most up-to-date state but is currently still lacking behind the real chain head
// during sync. The reason for this is simplicity. The ethereum-enclave needs this data to
// verify events. If needed we could change it to sync to the head and store a block_hash
// for each Epoch (~25h) that is then used for event verification. These messages should
// be sorted based on Finalized.Number, but there is no mechanism to ensure we receiver every
// message. Nor are there any guarantees about ordering (unless explicitly enforced here) or
// gaps.
//
// **FinalizedEventData** contains verified and filtered contract events. It provides much
// stronger guarantees: Ordering is guaranteed/enforced by checking the block numbers. This
// additionally guarantees that no block was missed. Processing starts at (or up to 25h before)
// `HandshakeMsg1.StartBlocknum`

type VerifiedHeadData struct {
	Finalized  ExecutionPayloadHeader `json:"finalized"`
	Optimistic ExecutionPayloadHeader `json:"optimistic"`
}
type ExecutionPayloadHeader struct {
	// There are more fields that we currently don't decode into this struct.
	ParentHash common.Hash `json:"parent_hash"`
	Number     uint64      `json:"block_number"`
	Timestamp  uint64      `json:"timestamp"` // Seconds
	BlockHash  common.Hash `json:"block_hash"`
}

type FinalizedEventData struct {
	Number    uint64  `json:"number"`
	Timestamp uint64  `json:"timestamp"` // Milliseconds
	Events    []Event `json:"events"`
}

// For now this code only supports RawEthereum events, as that's all we need in Shutter.
type Event struct {
	ContractId uint   `json:"contract_id"`
	Raw        RawEth `json:"raweth"`
}

type RawEth struct {
	Topics []string `json:"topics"` // hex
	Data   string   `json:"data"`   // hex
}

type Attestation struct {
	// The raw local attestation (already verified + checked according to config)
	Report *attestation.Report
	// Additional Data from the chain enclave (configuration). Should be checked against the application configuration.
	// Current format (for Ethereum): JSON
	Data []byte
}
type Connection struct {
	conn        net.Conn
	attestation Attestation

	// Channel for communication/use by other systems
	headers chan VerifiedHeadData
	events  chan FinalizedEventData
	errors  chan error
}
type Config struct {
	// SGX related
	SameSigner bool
	SignerID   []byte
	MrEnclave  []byte
	ProductID  *uint16
	MinISVSVN  uint16

	// Verifier settings
	EventExtractionStartBlocknum uint64 // Set to all ones to disable event extraction `^uint64(0)`
	Contracts                    []common.Address
}

func DialVerifiedChainDataChannel(address string, config Config) (*Connection, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("Could not connect: %w", err)
	}

	sk, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("Could not generate ECDH secret: %w", err)
	}

	selfreport_bytes, err := enclave.GetLocalReport(nil, nil)
	var target string
	var selfreport *attestation.Report
	if err == nil {
		target = "0x" + hex.EncodeToString(extract_target_info(selfreport_bytes))
		// The selfreport_bytes we have above is not signed and I'm not sure if we can fully trust it.
		// Ego gets another report for itself and verifies that, so we're doing the same to be on the safe side.
		r, err := enclave.GetLocalReport(nil, selfreport_bytes)
		if err != nil {
			return nil, fmt.Errorf("Could not get second selfreport: %w", err)
		}
		r2, err := enclave.VerifyLocalReport(r)
		if err != nil {
			return nil, fmt.Errorf("Could not verify selfreport: %w", err)
		}
		selfreport = &r2
	} else if err.Error() == "OE_UNSUPPORTED" {
		target = ""
	} else {
		return nil, fmt.Errorf("Could not get SelfReport: %w", err)
	}

	// Convert easy-to-configure format to the communication format.
	contracts := make([]contract, len(config.Contracts))
	for i, c := range config.Contracts {
		contracts[i] = contract{
			Addr: c.Hex(),
			Typ:  DECODE_TYPE_RAW,
		}
	}

	h1 := handshakeMsg1{
		Version:           1,
		AttestationTarget: target,
		ECDH_Pubkey:       "0x" + hex.EncodeToString(sk.PublicKey().Bytes()),
		// Current ethereum-runner expects us to always provide some data here.
		Data: erdstallParams{
			StartBlocknum: config.EventExtractionStartBlocknum,
			Contracts:     contracts,
		},
	}
	err = sendJSON(conn, h1)
	if err != nil {
		return nil, fmt.Errorf("Could not send HandshakeMsg1: %w", err)
	}
	var h2 handshakeMsg2
	err = recvJSON(conn, &h2)
	if err != nil {
		return nil, fmt.Errorf("Could not receive HandshakeMsg2: %w", err)
	}

	other_pk_bytes, err := hex.DecodeString(strings.TrimPrefix(h2.ECDH_Pubkey, "0x"))
	if err != nil {
		return nil, fmt.Errorf("Could not hex decode ECDH pubkey: %w", err)
	}
	other_pk, err := ecdh.X25519().NewPublicKey(other_pk_bytes)
	if err != nil {
		return nil, fmt.Errorf("Could not convert to ECDH pubkey: %w", err)
	}
	sharedSecret, err := sk.ECDH(other_pk)
	if err != nil {
		return nil, fmt.Errorf("Could not finish ECDH: %w", err)
	}
	mac := hmac.New(sha3.New256, sharedSecret)

	attestation_data, err := base64.StdEncoding.DecodeString(h2.AttestationData)
	if err != nil {
		return nil, fmt.Errorf("could not base64 decode attestation_data")
	}

	// Checking the report only makes sense if we've requested one (which we only do if we're running in a TEE, too)
	var report *attestation.Report
	if selfreport != nil {
		report_bytes, err := hex.DecodeString(strings.TrimPrefix(h2.Report, "0x"))
		if err != nil {
			return nil, fmt.Errorf("Could not hex decode report: %w", err)
		}
		r, err := enclave.VerifyLocalReport(decorate_report(report_bytes))
		if err != nil {
			return nil, fmt.Errorf("Invalid chain-enclave attestation report: %w", err)
		}

		err = verify_report(r, *selfreport, config, other_pk_bytes, attestation_data)
		if err != nil {
			return nil, err
		}
		report = &r
	}

	channel := make(chan VerifiedHeadData)
	errchannel := make(chan error)
	events := make(chan FinalizedEventData)
	attestation := Attestation{Report: report, Data: attestation_data}

	c := Connection{conn, attestation, channel, events, errchannel}
	go func() {
		errchannel <- c.run(mac)
	}()

	return &c, nil
}

func allZero(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}

// Caller should verify the signature itself (to get an attestation.Report). This function just checks the report data.
func verify_report(report attestation.Report, selfreport attestation.Report, config Config, other_pk_bytes []byte, attestation_data []byte) error {
	// Debug attribute
	if !selfreport.Debug && report.Debug {
		return fmt.Errorf("chain-enclave is in Debug mode while we are not")
	}

	// Product ID
	prodid := binary.LittleEndian.Uint16(report.ProductID[:2])
	if config.ProductID != nil && prodid != *config.ProductID {
		return fmt.Errorf("unexpected ProductID: %v (%v)", prodid, report.ProductID)
	}
	if config.ProductID != nil && !allZero(report.ProductID[2:]) {
		return fmt.Errorf("unexpected ProductID padding: %v", report.ProductID)
	}

	// SVN
	if config.MinISVSVN > 0 && report.SecurityVersion < uint(config.MinISVSVN) {
		return fmt.Errorf("SVN too low: %d < %d", report.SecurityVersion, config.MinISVSVN)
	}

	// Signer
	if config.SameSigner && !bytes.Equal(report.SignerID, selfreport.SignerID) {
		return fmt.Errorf("chain-enclave signer differs from our own")
	}
	if config.SignerID != nil && !bytes.Equal(report.SignerID, config.SignerID) {
		return fmt.Errorf("unexpected mrsigner")
	}

	// MRENCLAVE
	if config.MrEnclave != nil && !bytes.Equal(report.UniqueID, config.MrEnclave) {
		return fmt.Errorf("unexpected mrenclave")
	}

	// report.TCBStatus is always tcbstatus.Unknown for local attestations => No need to check

	// AttestationData (ecdh_pk)
	if !bytes.Equal(other_pk_bytes, report.Data[:32]) {
		return fmt.Errorf("ECDH PublicKey doesn't match report_data")
	}

	// AttestationData (extra)
	// See get_attestation_ecdh in lib/enclave/src/lib.rs
	// Contrary to Erdstall we have (and output) the complete attestation_data and should thus check its integrity.
	// The Erdstall enclave doesn't have this and only processes/returns its hash.
	hasher := sha3.New256()
	_, err := hasher.Write(attestation_data)
	if err != nil {
		return fmt.Errorf("Could not hash attestation_data: %w", err)
	}
	attestation_data_hash := hasher.Sum(nil)
	if !bytes.Equal(attestation_data_hash, report.Data[32:]) {
		return fmt.Errorf("Attestation data doesn't match report_data")
	}

	return nil
}

func (c *Connection) run(mac hash.Hash) error {
	for {
		var msg message
		err := recvJSON(c.conn, &msg)
		if err != nil {
			return fmt.Errorf("Could not receive message: %w", err)
		}

		data, err := base64.StdEncoding.DecodeString(msg.Data)
		if err != nil {
			return fmt.Errorf("Could not base64 decode message.Data: %w", err)
		}

		received_mac, err := base64.StdEncoding.DecodeString(msg.Mac)
		if err != nil {
			return fmt.Errorf("Could not hex decode message.Mac: %w", err)
		}

		mac.Reset()
		_, err = mac.Write(data)
		if err != nil {
			return fmt.Errorf("Could not hash message.Data: %w", err)
		}
		computed_mac := mac.Sum(nil)
		if !hmac.Equal(computed_mac, received_mac) {
			return fmt.Errorf("Invalid message MAC")
		}

		// First, try to decode as FinalizedEventData
		var d FinalizedEventData
		err1 := json.Unmarshal(data, &d)
		if err1 == nil && d.Number > 0 {
			c.events <- d
			continue
		}

		// If that fails: Try VerifiedHeadData
		var v VerifiedHeadData
		err2 := json.Unmarshal(data, &v)
		if err2 == nil {
			c.headers <- v
			continue
		}

		// Could not decode into either type, log as much as we can and fail loudly
		fmt.Println(string(data))
		fmt.Printf("Failed decoding: %v, %v", err1, err2)
		panic(err2)
	}
}

func (c *Connection) Close() {
	c.conn.Close()
}
func (c *Connection) Headers() <-chan VerifiedHeadData {
	return c.headers
}
func (c *Connection) Events() <-chan FinalizedEventData {
	return c.events
}
func (c *Connection) Errors() <-chan error {
	return c.errors
}
func (c *Connection) Attestation() Attestation {
	return c.attestation
}

func sendJSON(conn net.Conn, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	// 4-byte big-endian length prefixed json encoding
	_, err = conn.Write(binary.BigEndian.AppendUint32(nil, uint32(len(data))))
	if err != nil {
		return err
	}
	_, err = conn.Write(data)
	return err
}
func recvJSON(conn net.Conn, v any) error {
	var len_bytes [4]byte
	_, err := io.ReadFull(conn, len_bytes[:])
	if err != nil {
		return err
	}
	l := binary.BigEndian.Uint32(len_bytes[:])
	data := make([]byte, l)
	_, err = io.ReadFull(conn, data[:])
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

type handshakeMsg1 struct {
	Version           uint16         `json:"version"`
	AttestationTarget string         `json:"attestation_target,omitempty"` // hex
	ECDH_Pubkey       string         `json:"ecdh_pk"`                      // hex
	Data              erdstallParams `json:"data"`
}
type handshakeMsg2 struct {
	// Local attestation report (not a remote attestation quote)
	Report          string `json:"report"`           // hex | null
	ECDH_Pubkey     string `json:"ecdh_pk"`          // hex
	AttestationData string `json:"attestation_data"` // base64 encoded json
}
type message struct {
	Data string `json:"data"` // base64
	Mac  string `json:"mac"`  // base64 of [u8; 32]
}
type erdstallParams struct {
	StartBlocknum uint64     `json:"extraction_start_blocknum"`
	Contracts     []contract `json:"contracts"`
}
type contract struct {
	Addr string `json:"addr"`
	Typ  string `json:"typ"` // Decode Type
}

const TARGET_INFO_SIZE = 512
const REPORT_SIZE = 432
const OE_REPORT_HEADER_SIZE = 16

const DECODE_TYPE_RAW = "raw"

func extract_target_info(oe_report []byte) []byte {
	// EGo unfortunately has no way to directly get a TargetInfo object, but the
	// report contains everything we need, so we can build one ourselves.
	// Alternatively we could have sent the report to the rust side, but I think
	// it is better to use the types closer to the cpu instructions for
	// communicating between Rust and Go.

	// Details of report:
	// - https://github.com/openenclave/openenclave/blob/416ce1964ec1041e7f8fcd9ecedf44f951c92489/include/openenclave/bits/sgx/sgxtypes.h#L599-L660
	// - https://github.com/fortanix/rust-sgx/blob/c16c8aa29f1daf85159ff25f1290553bd334f879/intel-sgx/sgx-isa/src/lib.rs#L640-L660

	// Details of target info:
	// - https://github.com/openenclave/openenclave/blob/416ce1964ec1041e7f8fcd9ecedf44f951c92489/include/openenclave/bits/sgx/sgxtypes.h#L541-L557
	// - https://github.com/fortanix/rust-sgx/blob/c16c8aa29f1daf85159ff25f1290553bd334f879/intel-sgx/sgx-isa/src/lib.rs#L732-L756

	// Details of the report header in OpenEnclave:
	// - https://github.com/openenclave/openenclave/blob/cdeb95c1ec163117de409295333b6b2702013e08/enclave/core/sgx/report.c#L349-L361
	// - https://github.com/openenclave/openenclave/blob/cdeb95c1ec163117de409295333b6b2702013e08/include/openenclave/internal/report.h#L71-L77

	if len(oe_report) != OE_REPORT_HEADER_SIZE+REPORT_SIZE {
		panic(fmt.Sprintf("Unexpected OE report size, got %v (%d bytes)", oe_report, len(oe_report)))
	}
	report := oe_report[16:]

	target_info := make([]byte, TARGET_INFO_SIZE)
	copy(target_info[0:32], report[64:96])  // mrenclave
	copy(target_info[32:48], report[48:64]) // attributes
	copy(target_info[52:56], report[16:20]) // miscselect

	return target_info
}

const OE_REPORT_TYPE_SGX_LOCAL = 1
const OE_REPORT_TYPE_SGX_REMOTE = 2

// prepends relevant header data to a raw report
func decorate_report(report []byte) []byte {
	// For checking the report we need to add the header.
	if len(report) != REPORT_SIZE {
		panic(fmt.Sprintf("Unexpected report size, got %d (%d bytes)", report, len(report)))
	}
	oe_report := make([]byte, OE_REPORT_HEADER_SIZE+REPORT_SIZE)
	oe_report[0] = 0x01                     // version (uint32)
	oe_report[4] = OE_REPORT_TYPE_SGX_LOCAL // report_type (uint32)
	binary.LittleEndian.PutUint64(oe_report[8:16], REPORT_SIZE)
	copy(oe_report[16:], report)

	return oe_report
}
