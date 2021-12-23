package cryptocmd

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/shutter-network/shutter/shlib/shcrypto"
)

var (
	eonKeyFlag        string
	decryptionKeyFlag string
	epochIDFlag       string
	sigmaFlag         string
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crypto",
		Short: "CLI tool to access crypto functions",
		Long: `This command provides utility functions to manually encrypt messages with an eon
key, decrypt them with a decryption key, and check that a decryption key is correct.`,
	}
	cmd.AddCommand(encryptCmd())
	cmd.AddCommand(decryptCmd())
	cmd.AddCommand(verifyKeyCmd())
	return cmd
}

func encryptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt the message given as positional argument",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return encrypt(args[0])
		},
	}

	cmd.PersistentFlags().StringVarP(&eonKeyFlag, "eon-key", "k", "", "eon public key (hex encoded)")
	cmd.PersistentFlags().StringVarP(&epochIDFlag, "epoch-id", "e", "", "epoch id (hex encoded)")
	cmd.PersistentFlags().StringVarP(&sigmaFlag, "sigma", "s", "", "sigma (optional, hex encoded)")

	cmd.MarkPersistentFlagRequired("eon-key")
	cmd.MarkPersistentFlagRequired("epoch-id")

	return cmd
}

func decryptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt the message given as positional argument",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return decrypt(args[0])
		},
	}

	cmd.PersistentFlags().StringVarP(&decryptionKeyFlag, "decryption-key", "k", "", "decryption key (hex encoded)")
	cmd.MarkPersistentFlagRequired("decryption-key")

	return cmd
}

func verifyKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-key",
		Short: "Check that the decryption key given as positional argument is correct",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return verifyKey(args[0])
		},
	}

	cmd.PersistentFlags().StringVarP(&eonKeyFlag, "eon-key", "k", "", "eon public key (hex encoded)")
	cmd.PersistentFlags().StringVarP(&epochIDFlag, "epoch-id", "e", "", "epoch id (hex encoded)")

	cmd.MarkPersistentFlagRequired("eon-key")
	cmd.MarkPersistentFlagRequired("epoch-id")

	return cmd
}

func encrypt(msg string) error {
	eonKey, err := parseEonKey(eonKeyFlag)
	if err != nil {
		return err
	}
	epochIDInt, err := parseEpochID(epochIDFlag)
	if err != nil {
		return err
	}
	epochID := shcrypto.ComputeEpochID(epochIDInt)
	sigma, err := parseSigma(sigmaFlag)
	if err != nil {
		return err
	}

	msgBytes := []byte(msg)
	encryptedMsg := shcrypto.Encrypt(msgBytes, eonKey, epochID, sigma)
	fmt.Println("0x" + hex.EncodeToString(encryptedMsg.Marshal()))
	return nil
}

func decrypt(msg string) error {
	msgBytes, err := parseHex(msg)
	if err != nil {
		return err
	}
	decryptionKey, err := parseDecryptionKey(decryptionKeyFlag)
	if err != nil {
		return err
	}

	encryptedMsg := new(shcrypto.EncryptedMessage)
	err = encryptedMsg.Unmarshal(msgBytes)
	if err != nil {
		return errors.Wrap(err, "invalid encrypted message")
	}
	decryptedMsg, err := encryptedMsg.Decrypt(decryptionKey)
	if err != nil {
		return errors.Wrapf(err, "failed to decrypt message")
	}
	fmt.Println(string(decryptedMsg))
	return nil
}

func verifyKey(key string) error {
	decryptionKey, err := parseDecryptionKey(key)
	if err != nil {
		return err
	}
	eonKey, err := parseEonKey(eonKeyFlag)
	if err != nil {
		return err
	}
	epochID, err := parseEpochID(epochIDFlag)
	if err != nil {
		return err
	}
	ok, err := shcrypto.VerifyEpochSecretKey(decryptionKey, eonKey, epochID)
	if err != nil {
		return errors.Wrapf(err, "failed to verify decryption key")
	}
	if ok {
		fmt.Println("the given decryption key is valid")
		return nil
	}
	return errors.Errorf("the given decryption key is invalid")
}

func parseEonKey(f string) (*shcrypto.EonPublicKey, error) {
	eonKeyBytes, err := parseHex(f)
	if err != nil {
		return nil, err
	}
	eonKey := new(shcrypto.EonPublicKey)
	err = eonKey.Unmarshal(eonKeyBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid eon public key")
	}
	return eonKey, err
}

func parseEpochID(f string) (uint64, error) {
	epochIDBytes, err := parseHex(f)
	if err != nil {
		return 0, err
	}
	if len(epochIDBytes) != 8 {
		return 0, errors.Errorf("epoch id must be 8 bytes, got %d", len(epochIDBytes))
	}
	epochID := binary.BigEndian.Uint64(epochIDBytes)
	return epochID, nil
}

func parseSigma(f string) (shcrypto.Block, error) {
	if f == "" {
		return shcrypto.RandomSigma(rand.Reader)
	}
	sigmaBytes, err := parseHex(f)
	if err != nil {
		return shcrypto.Block{}, err
	}
	if len(sigmaBytes) != 32 {
		return shcrypto.Block{}, errors.Errorf("if given, sigma must be 32 bytes, got %d", len(sigmaBytes))
	}
	sigma := shcrypto.Block{}
	copy(sigma[:], sigmaBytes)
	return sigma, nil
}

func parseDecryptionKey(f string) (*shcrypto.EpochSecretKey, error) {
	decryptionKeyBytes, err := parseHex(f)
	if err != nil {
		return nil, err
	}
	decryptionKey := new(shcrypto.EpochSecretKey)
	err = decryptionKey.Unmarshal(decryptionKeyBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid decryption key")
	}
	return decryptionKey, nil
}

func parseHex(f string) ([]byte, error) {
	withoutPrefix := strings.TrimPrefix(f, "0x")
	b, err := hex.DecodeString(withoutPrefix)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid hex encoded argument")
	}
	return b, nil
}
