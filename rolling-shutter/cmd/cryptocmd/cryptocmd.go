package cryptocmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
)

var (
	eonKeyFlag        string
	decryptionKeyFlag string
	epochIDFlag       string
	sigmaFlag         string
	threshold         uint64
	filename          string
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crypto",
		Short: "CLI tool to access crypto functions",
		Long: `This command provides utility functions to manually encrypt messages with an eon
key, decrypt them with a decryption key, and check that a decryption key is correct. It also hosts
a tool to generate and run crypto tests in a JSON formatted collection.`,
	}
	cmd.AddCommand(encryptCmd())
	cmd.AddCommand(decryptCmd())
	cmd.AddCommand(verifyKeyCmd())
	cmd.AddCommand(aggregateCmd())
	cmd.AddCommand(GenerateTestdata())
	cmd.AddCommand(RunJSONTests())
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

func aggregateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aggregate",
		Short: "Aggregate key shares to construct a decryption key",
		Long: `Aggregate key shares to construct a decryption key.

Pass the shares as the first and only positional argument in the form of hex
values separated by commas. The shares must be ordered by keyper index and
missing shares must be denoted by empty strings. Exactly "threshold" shares
must be provided.

Example: rolling-shutter crypto aggregate -t 2 0BC5CDC5778D473B881E73297AFB830
1D35830786C6A80CD289672536655470A0149BA7394DF240C96F7D60BAF94D0FD2A39B4314088E
AF94E3D1EB52106E718,,03646AE08A8EF00D0AE04294529466C0F7AC65C4D9B0ADEAD1461964A
6F784202B61B2EBD5DE8B80E787E9FD4DE4899880C2263B67EC478D88D3558B0C22DA66`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return aggregate(args[0])
		},
	}

	cmd.PersistentFlags().Uint64VarP(&threshold, "threshold", "t", 0, "threshold parameter")
	cmd.MarkPersistentFlagRequired("threshold")

	return cmd
}

func encrypt(msg string) error {
	eonKey, err := parseEonKey(eonKeyFlag)
	if err != nil {
		return err
	}
	epochID, err := parseEpochID(epochIDFlag)
	if err != nil {
		return err
	}
	epochIDPoint := shcrypto.ComputeEpochID(epochID.Bytes())
	sigma, err := parseSigma(sigmaFlag)
	if err != nil {
		return err
	}

	msgBytes := []byte(msg)
	encryptedMsg := shcrypto.Encrypt(msgBytes, eonKey, epochIDPoint, sigma)
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
	ok, err := shcrypto.VerifyEpochSecretKey(decryptionKey, eonKey, epochID.Bytes())
	if err != nil {
		return errors.Wrapf(err, "failed to verify decryption key")
	}
	if ok {
		fmt.Println("the given decryption key is valid")
		return nil
	}
	return errors.Errorf("the given decryption key is invalid")
}

func aggregate(commaSeparatedHexShares string) error {
	hexShares := strings.Split(commaSeparatedHexShares, ",")
	indices := []int{}
	shares := []*shcrypto.EpochSecretKeyShare{}
	for i, hexShare := range hexShares {
		if hexShare == "" {
			continue
		}
		share, err := parseDecryptionKeyShare(hexShare)
		if err != nil {
			return errors.Wrapf(err, "key share %d is neither empty nor a valid hex encoded secret key share", i)
		}
		indices = append(indices, i)
		shares = append(shares, share)
	}

	key, err := shcrypto.ComputeEpochSecretKey(indices, shares, threshold)
	if err != nil {
		return err
	}
	fmt.Println("0x" + hex.EncodeToString(key.Marshal()))
	return nil
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

func parseEpochID(f string) (identitypreimage.IdentityPreimage, error) {
	epochIDBytes, err := parseHex(f)
	if err != nil {
		return identitypreimage.IdentityPreimage{}, err
	}
	return identitypreimage.IdentityPreimage(epochIDBytes), nil
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

func parseDecryptionKeyShare(f string) (*shcrypto.EpochSecretKeyShare, error) {
	decryptionKeyShareBytes, err := parseHex(f)
	if err != nil {
		return nil, err
	}
	decryptionKeyShare := new(shcrypto.EpochSecretKeyShare)
	err = decryptionKeyShare.Unmarshal(decryptionKeyShareBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid decryption key share")
	}
	return decryptionKeyShare, nil
}

func parseHex(f string) ([]byte, error) {
	withoutPrefix := strings.TrimPrefix(f, "0x")
	b, err := hex.DecodeString(withoutPrefix)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid hex encoded argument")
	}
	return b, nil
}
