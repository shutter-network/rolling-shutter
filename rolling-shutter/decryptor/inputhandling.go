package decryptor

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func handleDecryptionKeyInput(
	ctx context.Context,
	config Config,
	db *dcrdb.Queries,
	key *decryptionKey,
) ([]shmsg.P2PMessage, error) {
	keyBytes, _ := key.key.GobEncode()
	tag, err := db.InsertDecryptionKey(ctx, dcrdb.InsertDecryptionKeyParams{
		EpochID: medley.Uint64EpochIDToBytes(key.epochID),
		Key:     keyBytes,
	})
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		log.Printf("attempted to store multiple keys for same epoch %d", key.epochID)
		return nil, nil
	}
	return handleEpoch(ctx, config, db, key.epochID)
}

func handleCipherBatchInput(
	ctx context.Context,
	config Config,
	db *dcrdb.Queries,
	cipherBatch *cipherBatch,
) ([]shmsg.P2PMessage, error) {
	tag, err := db.InsertCipherBatch(ctx, dcrdb.InsertCipherBatchParams{
		EpochID:      medley.Uint64EpochIDToBytes(cipherBatch.EpochID),
		Transactions: cipherBatch.Transactions,
	})
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		log.Printf("attempted to store multiple cipherbatches for same epoch %d", cipherBatch.EpochID)
		return nil, nil
	}
	return handleEpoch(ctx, config, db, cipherBatch.EpochID)
}

func handleSignatureInput(
	ctx context.Context,
	config Config,
	db *dcrdb.Queries,
	signature *decryptionSignature,
) ([]shmsg.P2PMessage, error) {
	signers := getIndexes(signature.SignerBitfield)
	if len(signers) > 1 {
		// Ignore aggregated signatures
		return nil, nil
	}
	tag, err := db.InsertDecryptionSignature(ctx, dcrdb.InsertDecryptionSignatureParams{
		EpochID:         medley.Uint64EpochIDToBytes(signature.epochID),
		SignedHash:      signature.signedHash.Bytes(),
		SignersBitfield: signature.SignerBitfield,
		Signature:       signature.signature.Marshal(),
	})
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		log.Printf("attempted to store multiple decryption signatures with same epoch and signers")
		return nil, nil
	}

	exists, err := db.ExistsAggregatedSignature(ctx, signature.signedHash.Bytes())
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, nil
	}

	// check if we have enough signatures
	dbSignatures, err := db.GetDecryptionSignatures(ctx, dcrdb.GetDecryptionSignaturesParams{
		EpochID:    medley.Uint64EpochIDToBytes(signature.epochID),
		SignedHash: signature.signedHash.Bytes(),
	})
	if err != nil {
		return nil, err
	}
	if uint(len(dbSignatures)) < config.RequiredSignatures {
		return nil, nil
	}

	signaturesToAggregate := make([]*shbls.Signature, 0, len(dbSignatures))
	publicKeysToAggragate := make([]*shbls.PublicKey, 0, len(dbSignatures))
	bitfield := make([]byte, len(signature.SignerBitfield))
	for _, dbSignature := range dbSignatures {
		unmarshalledSignature := new(shbls.Signature)
		if err := unmarshalledSignature.Unmarshal(dbSignature.Signature); err != nil {
			log.Printf("failed to unmarshal signature from db %s", err)
			continue
		}

		indexes := getIndexes(dbSignature.SignersBitfield)
		if len(indexes) > 1 {
			panic("got signature with multiple signers")
		}
		if len(indexes) == 0 {
			panic("could not retrieve signer index from bitfield")
		}
		pkBytes, err := db.GetDecryptorKey(ctx, dcrdb.GetDecryptorKeyParams{Index: indexes[0], StartEpochID: dbSignature.EpochID})
		if err != nil {
			return nil, err
		}
		pk := new(shbls.PublicKey)
		if err := pk.Unmarshal(pkBytes); err != nil {
			log.Printf("failed to unmarshal public key from db %s", err)
			continue
		}

		signaturesToAggregate = append(signaturesToAggregate, unmarshalledSignature)
		publicKeysToAggragate = append(publicKeysToAggragate, pk)
		bitfield = addBitfields(bitfield, dbSignature.SignersBitfield)
	}

	if uint(len(signaturesToAggregate)) < config.RequiredSignatures {
		return nil, nil
	}

	aggregatedSignature := shbls.AggregateSignatures(signaturesToAggregate)
	aggregatedKey := shbls.AggregatePublicKeys(publicKeysToAggragate)
	if !shbls.Verify(aggregatedSignature, aggregatedKey, signature.signedHash.Bytes()) {
		panic(fmt.Sprintf("could not verify aggregated signature for epochID %d", signature.epochID))
	}

	_, err = db.InsertAggregatedSignature(ctx, dcrdb.InsertAggregatedSignatureParams{
		EpochID:         medley.Uint64EpochIDToBytes(signature.epochID),
		SignedHash:      signature.signedHash.Bytes(),
		SignersBitfield: bitfield,
		Signature:       aggregatedSignature.Marshal(),
	})
	if err != nil {
		return nil, err
	}

	msgs := []shmsg.P2PMessage{
		&shmsg.AggregatedDecryptionSignature{
			InstanceID:          config.InstanceID,
			EpochID:             signature.epochID,
			SignedHash:          signature.signedHash.Bytes(),
			AggregatedSignature: aggregatedSignature.Marshal(),
			SignerBitfield:      bitfield,
		},
	}

	return msgs, nil
}

// handleEpoch produces, store, and output a signature if we have both the cipher batch and key for given epoch.
func handleEpoch(
	ctx context.Context,
	config Config,
	db *dcrdb.Queries,
	epochID uint64,
) ([]shmsg.P2PMessage, error) {
	epochIDBytes := medley.Uint64EpochIDToBytes(epochID)
	cipherBatch, err := db.GetCipherBatch(ctx, epochIDBytes)
	if err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	decryptionKeyDB, err := db.GetDecryptionKey(ctx, epochIDBytes)
	if err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	log.Printf("decrypting batch for epoch %d", epochID)

	decryptionKey := new(shcrypto.EpochSecretKey)
	err = decryptionKey.GobDecode(decryptionKeyDB.Key)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid decryption key for epoch %d in db", epochID)
	}

	decryptedBatch := decryptCipherBatch(cipherBatch.Transactions, decryptionKey)
	signingData := DecryptionSigningData{
		InstanceID:     config.InstanceID,
		EpochID:        epochID,
		CipherBatch:    cipherBatch.Transactions,
		DecryptedBatch: decryptedBatch,
	}
	signedHash := signingData.Hash().Bytes()
	signature := signingData.Sign(config.SigningKey)
	signersBitfield := makeBitfieldFromIndex(config.SignerIndex)

	msgs, err := handleSignatureInput(ctx, config, db, &decryptionSignature{
		instanceID:     config.InstanceID,
		epochID:        epochID,
		signedHash:     common.BytesToHash(signedHash),
		signature:      signature,
		SignerBitfield: signersBitfield,
	})
	if err != nil {
		return nil, err
	}

	signatureMsg := &shmsg.DecryptionSignature{
		InstanceID:     config.InstanceID,
		EpochID:        epochID,
		SignedHash:     signedHash,
		Signature:      signature.Marshal(),
		SignerBitfield: signersBitfield,
	}
	msgs = append(msgs, signatureMsg)
	return msgs, nil
}
