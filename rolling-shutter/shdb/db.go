package shdb

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

// tableNamesQuery returns the names of all user created tables in the database.
const tableNamesQuery = `
	SELECT table_name
	FROM information_schema.tables
	WHERE table_schema = $1
`

// createDecryptorTables creates the tables for the decryptor db.
const createDecryptorTables = `
	CREATE SCHEMA IF NOT EXISTS decryptor;

	CREATE TABLE IF NOT EXISTS decryptor.cipher_batch (
		epoch_id bigint PRIMARY KEY,
		data bytea
	);
	CREATE TABLE IF NOT EXISTS decryptor.decryption_key (
		epoch_id bigint PRIMARY KEY,
		key bytea
	);
	CREATE TABLE IF NOT EXISTS decryptor.decryption_signature (
		epoch_id bigint,
		signed_hash bytea,
		signer_index bigint,
		signature bytea,
		PRIMARY KEY (epoch_id, signer_index)
	);
`

const createKeyperTables = `
	CREATE SCHEMA IF NOT EXISTS keyper;
	CREATE TABLE IF NOT EXISTS keyper.decryption_trigger (
		epoch_id bigint PRIMARY KEY
	);
	CREATE TABLE IF NOT EXISTS keyper.decryption_key_share (
		epoch_id bigint,
		keyper_index bigint,
		decryption_key_share bytea,
		PRIMARY KEY (epoch_id, keyper_index)
	);
	CREATE TABLE IF NOT EXISTS keyper.decryption_key (
		epoch_id bigint PRIMARY KEY,
		keyper_index bigint,
		decryption_key bytea
	);
`

// InitDecryptorDB initializes the database of the decryptor. It is assumed that the db is empty.
func InitDecryptorDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	_, err := dbpool.Exec(ctx, createDecryptorTables)
	if err != nil {
		return errors.Wrap(err, "failed to create decryptor tables")
	}
	return nil
}

// InitKeyperDB initializes the database of the keyper. It is assumed that the db is empty.
func InitKeyperDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	_, err := dbpool.Exec(ctx, createKeyperTables)
	if err != nil {
		return errors.Wrap(err, "failed to create keyper tables")
	}
	return nil
}

// ValidateKeyperDB checks that all expected tables exist in the database. If not, it returns an
// error.
func ValidateKeyperDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	return validateDB(ctx, dbpool, "keyper", []string{
		"decryption_trigger",
		"decryption_key_share",
		"decryption_key",
	})
}

// ValidateDecryptorDB checks that all expected tables exist in the database. If not, it returns an
// error.
func ValidateDecryptorDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	return validateDB(ctx, dbpool, "decryptor", []string{
		"cipher_batch",
		"decryption_key",
		"decryption_signature",
	})
}

func validateDB(ctx context.Context, dbpool *pgxpool.Pool, schema string, requiredTables []string) error {
	requiredTableMap := make(map[string]bool)
	for _, table := range requiredTables {
		requiredTableMap[table] = true
	}

	rows, err := dbpool.Query(ctx, tableNamesQuery, schema)
	if err != nil {
		return errors.Wrap(err, "failed to query table names from db")
	}
	defer rows.Close()

	var tableName string
	for rows.Next() {
		err := rows.Scan(&tableName)
		if err != nil {
			return errors.Wrap(err, "failed to query table names from db")
		}
		delete(requiredTableMap, tableName)
	}
	if rows.Err() != nil {
		return errors.Wrap(rows.Err(), "read table names")
	}

	if len(requiredTableMap) != 0 {
		return errors.New("database misses one or more required table")
	}
	return nil
}
