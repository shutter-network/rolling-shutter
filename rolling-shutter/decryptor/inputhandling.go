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
	"github.com/shutter-network/shutter/shuttermint/medley/bitfield"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func handleDecryptionKeyInput(
	ctx context.Context,
	config Config,
	db *dcrdb.Queries,
	key *decryptionKey,
) ([]shmsg.P2PMessage, error) {
	keyBytes := key.key.Marshal()
	tag, err := db.InsertDecryptionKey(ctx, dcrdb.InsertDecryptionKeyParams{
		EpochID: shdb.EncodeUint64(key.epochID),
		Key:     keyBytes,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to insert decryption key for epoch %d into db", key.epochID)
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
		EpochID:      shdb.EncodeUint64(cipherBatch.EpochID),
		Transactions: cipherBatch.Transactions,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to insert cipher batch for epoch %d into db", cipherBatch.EpochID)
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
	signers := bitfield.GetIndexes(signature.SignerBitfield)
	if len(signers) > 1 {
		// Ignore aggregated signatures
		return nil, nil
	}
	tag, err := db.InsertDecryptionSignature(ctx, dcrdb.InsertDecryptionSignatureParams{
		EpochID:         shdb.EncodeUint64(signature.epochID),
		SignedHash:      signature.signedHash.Bytes(),
		SignersBitfield: signature.SignerBitfield,
		Signature:       signature.signature.Marshal(),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to insert decryption signature for epoch %d into db", signature.epochID)
	}
	if tag.RowsAffected() == 0 {
		log.Printf("attempted to store multiple decryption signatures with same epoch and signers")
		return nil, nil
	}

	exists, err := db.ExistsAggregatedSignature(ctx, signature.signedHash.Bytes())
	if err != nil {
		return nil, errors.Wrap(err, "failed to check if db contains aggregated signature")
	}
	if exists {
		return nil, nil
	}

	// check if we have enough signatures
	dbSignatures, err := db.GetDecryptionSignatures(ctx, dcrdb.GetDecryptionSignaturesParams{
		EpochID:    shdb.EncodeUint64(signature.epochID),
		SignedHash: signature.signedHash.Bytes(),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query decryption signatures for epoch %d from db", signature.epochID)
	}
	if uint(len(dbSignatures)) < config.RequiredSignatures {
		return nil, nil
	}

	signaturesToAggregate := make([]*shbls.Signature, 0, len(dbSignatures))
	publicKeysToAggragate := make([]*shbls.PublicKey, 0, len(dbSignatures))
	signerBitfield := make([]byte, len(signature.SignerBitfield))
	for _, dbSignature := range dbSignatures {
		unmarshalledSignature := new(shbls.Signature)
		if err := unmarshalledSignature.Unmarshal(dbSignature.Signature); err != nil {
			log.Printf("failed to unmarshal signature from db %s", err)
			continue
		}

		indexes := bitfield.GetIndexes(dbSignature.SignersBitfield)
		if len(indexes) > 1 {
			panic("got signature with multiple signers")
		}
		if len(indexes) == 0 {
			panic("could not retrieve signer index from bitfield")
		}
		epochID := shdb.DecodeUint64(dbSignature.EpochID)
		activationBlockNumber := medley.ActivationBlockNumberFromEpochID(epochID)
		pkBytes, err := db.GetDecryptorKey(ctx, dcrdb.GetDecryptorKeyParams{
			Index:                 indexes[0],
			ActivationBlockNumber: int64(activationBlockNumber),
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to query public key of decryptor #%d from db", indexes[0])
		}
		pk := new(shbls.PublicKey)
		if err := pk.Unmarshal(pkBytes); err != nil {
			log.Printf("failed to unmarshal public key from db %s", err)
			continue
		}

		signaturesToAggregate = append(signaturesToAggregate, unmarshalledSignature)
		publicKeysToAggragate = append(publicKeysToAggragate, pk)
		signerBitfield = bitfield.AddBitfields(signerBitfield, dbSignature.SignersBitfield)
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
		EpochID:         shdb.EncodeUint64(signature.epochID),
		SignedHash:      signature.signedHash.Bytes(),
		SignersBitfield: signerBitfield,
		Signature:       aggregatedSignature.Marshal(),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error inserting aggregated signature for epoch %d into db", signature.epochID)
	}

	msgs := []shmsg.P2PMessage{
		&shmsg.AggregatedDecryptionSignature{
			InstanceID:          config.InstanceID,
			EpochID:             signature.epochID,
			SignedHash:          signature.signedHash.Bytes(),
			AggregatedSignature: aggregatedSignature.Marshal(),
			SignerBitfield:      signerBitfield,
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
	epochIDBytes := shdb.EncodeUint64(epochID)
	cipherBatch, err := db.GetCipherBatch(ctx, epochIDBytes)
	if err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "failed to get cipher batch for epoch %d from db", epochID)
	}

	decryptionKeyDB, err := db.GetDecryptionKey(ctx, epochIDBytes)
	if err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "failed to get decryption key for epoch %d from db", epochID)
	}

	log.Printf("decrypting batch for epoch %d", epochID)

	decryptionKey := new(shcrypto.EpochSecretKey)
	err = decryptionKey.Unmarshal(decryptionKeyDB.Key)
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
	signersBitfield := bitfield.MakeBitfieldFromIndex(config.SignerIndex)

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
