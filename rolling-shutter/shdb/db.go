package shdb

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shlib/puredkg"
	"github.com/shutter-network/shutter/shlib/shcrypto"
)

const SchemaVersionKey = "schema-version"

// ConnectionInfo returns a string describing the current database connection.
func ConnectionInfo(dbpool *pgxpool.Pool) string {
	cc := dbpool.Config().ConnConfig
	return fmt.Sprintf("host=%s, user=%s, db=%s", cc.Host, cc.User, cc.Database)
}

// MustFindSchemaVersion extracts the expected schema version from schema.sql files that should
// contain a header like:
// -- schema-version: 23 --
// It aborts the program, if the version cannot be determined.
func MustFindSchemaVersion(schema, path string) string {
	rx := "-- schema-version: ([0-9]+) --"
	matches := regexp.MustCompile(rx).FindStringSubmatch(schema)
	if len(matches) != 2 {
		log.Fatalf("malformed schema in %s, cannot find regular expression %s", path, rx)
	}
	return matches[1]
}

func EncodeEciesPublicKey(key *ecies.PublicKey) []byte {
	return ethcrypto.FromECDSAPub(key.ExportECDSA())
}

func DecodeEciesPublicKey(data []byte) (*ecies.PublicKey, error) {
	k, err := ethcrypto.UnmarshalPubkey(data)
	if err != nil {
		return nil, err
	}
	return ecies.ImportECDSAPublic(k), nil
}

func EncodeAddress(addr common.Address) string {
	return addr.Hex()
}

func DecodeAddress(data string) (common.Address, error) {
	if !common.IsHexAddress(data) {
		return common.Address{}, errors.Errorf("not an address: %s", data)
	}
	return common.HexToAddress(data), nil
}

func EncodeAddresses(addrs []common.Address) []string {
	encodedAddrs := []string{}
	for _, addr := range addrs {
		encodedAddr := EncodeAddress(addr)
		encodedAddrs = append(encodedAddrs, encodedAddr)
	}
	return encodedAddrs
}

func DecodeAddresses(data []string) ([]common.Address, error) {
	addrs := []common.Address{}
	for _, d := range data {
		addr, err := DecodeAddress(d)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

func EncodePureDKG(p *puredkg.PureDKG) ([]byte, error) {
	buff := bytes.Buffer{}
	err := gob.NewEncoder(&buff).Encode(p)
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func DecodePureDKG(data []byte) (*puredkg.PureDKG, error) {
	buf := bytes.NewBuffer(data)
	p := &puredkg.PureDKG{}
	err := gob.NewDecoder(buf).Decode(p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func EncodeUint64(n uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, n)
	return b
}

func DecodeUint64(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}

func EncodeBigint(n *big.Int) []byte {
	return n.Bytes()
}

func DecodeBigint(b []byte) *big.Int {
	return new(big.Int).SetBytes(b)
}

func EncodePureDKGResult(result *puredkg.Result) ([]byte, error) {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(result)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodePureDKGResult(b []byte) (*puredkg.Result, error) {
	res := puredkg.Result{}
	err := gob.NewDecoder(bytes.NewBuffer(b)).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func EncodeEpochSecretKeyShare(share *shcrypto.EpochSecretKeyShare) []byte {
	return share.Marshal()
}

func DecodeEpochSecretKeyShare(b []byte) (*shcrypto.EpochSecretKeyShare, error) {
	share := new(shcrypto.EpochSecretKeyShare)
	err := share.Unmarshal(b)
	if err != nil {
		return nil, err
	}
	return share, nil
}
