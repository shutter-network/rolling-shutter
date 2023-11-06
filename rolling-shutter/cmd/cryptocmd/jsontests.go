package cryptocmd

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/common/hexutil"
	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"github.com/spf13/cobra"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testkeygen"
)

func GenerateTestdata() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "testdata",
		Short: "Generate testdata in json format to test crypto implementations",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var w io.Writer
			if len(args) == 0 {
				w = os.Stdout
			} else {
				f, err := os.OpenFile(args[0], os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
				if err != nil {
					panic(err)
				}
				defer f.Close()
				w = f
			}

			enc := &testEncoder{
				w: w,
				e: 1,
				d: 1,
				v: 1,
			}
			enc.start()
			CreateJSONTests(*enc)

			enc.flush()
			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&filename, "filename", "f", "", "filename to write result")

	return cmd
}

func RunJSONTests() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jsontests",
		Short: "Use testdata in json format to test crypto implementations",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			errs := ReadTestcases(filename)

			return errs[0]
		},
	}

	cmd.PersistentFlags().StringVarP(&filename, "filename", "f", "", "filename to write result")
	return cmd
}

type testEncoder struct {
	w io.Writer
	i int
	e int
	d int
	v int
}

func (enc *testEncoder) addTest(tc *testCase) {
	tc.ID = fmt.Sprint(enc.i)
	switch tc.TestType {
	case ENCRYPTION:
		tc.Name = fmt.Sprintf("%s %d", ENCRYPTION, enc.e)
		enc.e++
	case DECRYPTION:
		tc.Name = fmt.Sprintf("%s %d", DECRYPTION, enc.d)
		enc.d++
	case VERIFICATION:
		tc.Name = fmt.Sprintf("%s %d", VERIFICATION, enc.v)
		enc.v++
	default:
		panic(fmt.Errorf("unknown test type"))
	}

	const indent = "  "
	encoded, err := json.MarshalIndent(tc, indent, indent)
	if err != nil {
		panic(fmt.Errorf("can't encode test case: %v", err))
	}

	var buf bytes.Buffer
	if enc.i > 0 {
		buf.WriteString(",\n")
	}
	buf.WriteString(indent)
	buf.Write(encoded)
	if _, err := buf.WriteTo(enc.w); err != nil {
		panic(err)
	}

	enc.i++
}

func (enc *testEncoder) start() {
	_, err := io.WriteString(enc.w, "[\n")
	if err != nil {
		panic(err)
	}
}

func (enc *testEncoder) flush() {
	_, err := io.WriteString(enc.w, "\n]\n")
	if err != nil {
		panic(err)
	}
}

type testCase struct {
	testCaseMeta
	Test testData `json:"test_data"`
}

type testCaseMeta struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	Description string `json:"description"`
	TestType    string `json:"type"`
}

type testData interface {
	Run() error
}

func (tc *testCase) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &tc.testCaseMeta); err != nil {
		return err
	}
	switch tc.TestType {
	case ENCRYPTION:
		tc.Test = new(encryptionTest)
	case DECRYPTION:
		tc.Test = new(decryptionTest)
	case VERIFICATION:
		tc.Test = new(verificationTest)
	default:
		return fmt.Errorf("invalid test type %q", tc.Test)
	}

	var testRaw struct {
		Data json.RawMessage `json:"test_data"`
	}
	if err := json.Unmarshal(b, &testRaw); err != nil {
		return err
	}
	return json.Unmarshal(testRaw.Data, tc.Test)
}

const (
	ENCRYPTION   = "encryption"
	DECRYPTION   = "decryption"
	VERIFICATION = "verification"
)

type encryptionTest struct {
	Message      hexutil.Bytes              `json:"message"`
	EonPublicKey *shcrypto.EonPublicKey     `json:"eon_public_key"`
	EpochID      epochid.EpochID            `json:"epoch_id"`
	Sigma        shcrypto.Block             `json:"sigma"`
	Expected     *shcrypto.EncryptedMessage `json:"expected"`
}

type decryptionTest struct {
	Cipher         shcrypto.EncryptedMessage `json:"cipher"`
	EpochSecretKey shcrypto.EpochSecretKey   `json:"epoch_secret_key"`
	Expected       hexutil.Bytes             `json:"expected"`
}

type verificationTest struct {
	EpochSecretKey shcrypto.EpochSecretKey `json:"epoch_secret_key"`
	EonPublicKey   shcrypto.EonPublicKey   `json:"eon_public_key"`
	EpochID        epochid.EpochID         `json:"epoch_id"`
	Expected       bool                    `json:"expected"`
}

func ReadTestcases(filename string) []error {
	var testcases []*testCase

	f, err := os.OpenFile(filename, os.O_RDONLY, 0o600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	decoder := json.NewDecoder(f)

	err = decoder.Decode(&testcases)
	if err != nil {
		panic(err)
	}
	var failed int
	errs := make([]error, len(testcases))
	for i, testcase := range testcases {
		fmt.Printf(
			"[%03d/%03d] '%-14s': %-42s ID: % 3s",
			i+1, len(testcases),
			testcase.Name,
			testcase.Description,
			testcase.ID,
		)
		err = testcase.Test.Run()

		if err != nil {
			failed++
			errs = append(errs, err)
			fmt.Printf(" FAIL (%s)\n", err)
		} else {
			fmt.Printf(" PASS\n")
		}
	}
	fmt.Printf("%03d tests failed.\n", failed)
	return errs
}

const (
	RANDOM       = "random"
	FIXED        = "fixed"
	TAMPERED     = "tampered"
	VERIFYING    = "verifying"
	NONVERIFYING = "nonverifying"
)

var testSpecs = []struct {
	description string
	payload     []byte
	style       string
}{
	{
		"A zero byte message.",
		make([]byte, 0),
		RANDOM,
	},
	{
		"A 1 byte message.",
		make([]byte, 1),
		RANDOM,
	},
	{
		"A 31 byte message.",
		make([]byte, 31),
		RANDOM,
	},
	{
		"A 32 byte message.",
		make([]byte, 32),
		RANDOM,
	},
	{
		"A 33 byte message.",
		make([]byte, 33),
		RANDOM,
	},
	{
		"A 319 byte message.",
		make([]byte, 319),
		RANDOM,
	},
	{
		"A 320 byte message.",
		make([]byte, 320),
		RANDOM,
	},
	{
		"A 321 byte message.",
		make([]byte, 321),
		RANDOM,
	},
	{
		"The message 'A message'",
		[]byte("A message"),
		FIXED,
	}, /*
		{
			"An illegal modification of the encrypted message 'A message'",
			[]byte("A message"),
			TAMPERED,
		},*/
	{
		"Verification of a random 32 byte epochID",
		make([]byte, 32),
		VERIFYING,
	},
	{
		"A failed verification",
		make([]byte, 32),
		NONVERIFYING,
	},
}

func CreateJSONTests(enc testEncoder) {
	keygen := testkeygen.NewKeyGenerator(12, 10)
	var err error
	for i := range testSpecs {
		testSpec := testSpecs[i]

		switch testSpec.style {
		case RANDOM, FIXED:

			if testSpec.style == RANDOM {
				_, err = rand.Read(testSpec.payload)
			}
			if err != nil {
				panic(err)
			}
			et, err := createEncryptionTest(keygen, testSpec.payload)
			if err != nil {
				panic(err)
			}
			testcase := testCase{
				testCaseMeta: testCaseMeta{
					Description: testSpec.description,
					TestType:    ENCRYPTION,
				},
				Test: et,
			}
			if err := verifyTestCase(&testcase); err != nil {
				panic(err)
			}
			enc.addTest(&testcase)

			dt := createDecryptionTest(keygen, *et)
			testcase = testCase{
				testCaseMeta: testCaseMeta{
					Description: testSpec.description,
					TestType:    DECRYPTION,
				},
				Test: &dt,
			}

			if err = verifyTestCase(&testcase); err != nil {
				panic(err)
			}
			enc.addTest(&testcase)

		case TAMPERED:
			et, err := createEncryptionTest(keygen, testSpec.payload)
			if err != nil {
				panic(err)
			}
			tamperedEt := tamperEncryptedMessage(keygen, *et)

			dt := createDecryptionTest(keygen, tamperedEt)
			dt.Expected, _ = hexutil.Decode("0x")
			testcase := testCase{
				testCaseMeta: testCaseMeta{
					Description: testSpec.description,
					TestType:    DECRYPTION,
				},
				Test: &dt,
			}
			if err := verifyTestCase(&testcase); err != nil {
				panic(err)
			}
			enc.addTest(&testcase)
		case VERIFYING, NONVERIFYING:
			var err error
			var vt verificationTest
			if testSpec.style == VERIFYING {
				vt, err = createVerificationTest(keygen, testSpec.payload)
			} else {
				vt, err = createFailedVerificationTest(keygen, testSpec.payload)
			}
			if err != nil {
				panic(err)
			}
			testcase := testCase{
				testCaseMeta: testCaseMeta{
					Description: testSpec.description,
					TestType:    VERIFICATION,
				},
				Test: &vt,
			}
			if err := verifyTestCase(&testcase); err != nil {
				panic(err)
			}
			enc.addTest(&testcase)

		default:
			panic("no test style defined")
		}
	}
}

func verifyTestCase(tc *testCase) error {
	if err := testMarshalingRoundtrip(tc); err != nil {
		return err
	}
	return tc.Test.Run()
}

func createEncryptionTest(keygen *testkeygen.KeyGenerator, message []byte) (*encryptionTest, error) {
	epochID := keygen.RandomEpochID(make([]byte, 32))

	et := encryptionTest{}

	et.Message = message

	et.EonPublicKey = keygen.EonPublicKey(epochID)
	et.EpochID = epochID
	sigma, err := keygen.RandomSigma()
	if err != nil {
		return &et, err
	}
	et.Sigma = sigma

	epochIDPoint := shcrypto.ComputeEpochID(epochID.Bytes())

	encryptedMessage := shcrypto.Encrypt(
		et.Message,
		keygen.EonPublicKey(epochID),
		epochIDPoint,
		sigma,
	)

	et.Expected = encryptedMessage

	if et.Expected.C1 == nil {
		return &et, errors.New("failed to marshal")
	}

	err = et.Run()
	return &et, err
}

// tamperEncryptedMessage changes the C1 value of EncryptedMessage, which allows to test for malleability issues.
func tamperEncryptedMessage(keygen *testkeygen.KeyGenerator, et encryptionTest) encryptionTest {
	decryptionKey := keygen.EpochSecretKey(et.EpochID)
	var c1 *bn256.G2
	var err error

	for i := 1; i <= 10000; i++ {
		c1 = et.Expected.C1
		c1.Add(c1, c1)
		et.Expected.C1 = c1
		sigma := et.Expected.Sigma(decryptionKey)
		decryptedBlocks := shcrypto.DecryptBlocks(et.Expected.C3, sigma)
		_, err = shcrypto.UnpadMessage(decryptedBlocks)

		if err == nil {
			break
		}
	}
	return et
}

func createDecryptionTest(keygen *testkeygen.KeyGenerator, et encryptionTest) decryptionTest {
	dt := decryptionTest{}
	epochSecretKey := keygen.EpochSecretKey(et.EpochID)
	dt.EpochSecretKey = *epochSecretKey

	dt.Cipher = *et.Expected

	dt.Expected = et.Message

	return dt
}

func createVerificationTest(keygen *testkeygen.KeyGenerator, payload []byte) (verificationTest, error) {
	var err error
	vt := verificationTest{}
	epochID := keygen.RandomEpochID(payload)
	vt.EpochID = epochID
	vt.EpochSecretKey = *keygen.EpochSecretKey(epochID)
	vt.EonPublicKey = *keygen.EonPublicKey(epochID)
	vt.Expected, err = shcrypto.VerifyEpochSecretKey(
		&vt.EpochSecretKey,
		&vt.EonPublicKey,
		epochID.Bytes(),
	)
	return vt, err
}

func createFailedVerificationTest(keygen *testkeygen.KeyGenerator, _ []byte) (verificationTest, error) {
	var err error
	vt := verificationTest{}
	epochID := keygen.RandomEpochID(make([]byte, 32))
	mismatch := keygen.RandomEpochID(make([]byte, 32))
	vt.EpochID = epochID
	vt.EpochSecretKey = *keygen.EpochSecretKey(epochID)
	vt.EonPublicKey = *keygen.EonPublicKey(mismatch)
	vt.Expected, err = shcrypto.VerifyEpochSecretKey(
		&vt.EpochSecretKey,
		&vt.EonPublicKey,
		epochID.Bytes(),
	)
	return vt, err
}

func (et *encryptionTest) Run() error {
	result := shcrypto.Encrypt(
		et.Message,
		et.EonPublicKey,
		shcrypto.ComputeEpochID(et.EpochID.Bytes()),
		et.Sigma,
	)

	encoded, err := result.MarshalText()
	if err != nil {
		return fmt.Errorf("failed encryption test on encoding result: %s", err)
	}
	expectation, err := et.Expected.MarshalText()
	if err != nil {
		return fmt.Errorf("failed encryption test on encoding expected: %s", err)
	}
	if !bytes.Equal(encoded, expectation) {
		return fmt.Errorf("failed encryption test on equal results: %s != %s", encoded, et.Expected)
	}
	return nil
}

func (dt *decryptionTest) Run() error {
	expectation, err := dt.Expected.MarshalText()
	if err != nil {
		return err
	}

	result, err := dt.Cipher.Decrypt(
		&dt.EpochSecretKey,
	)
	if err != nil {
		if !bytes.Equal(expectation, []byte("0x")) {
			return fmt.Errorf("failed decryption test with error: %s", err)
		}
		return nil
	}
	encoded := hexutil.Encode(result)

	if !bytes.Equal(result, dt.Expected) {
		return fmt.Errorf("failed decryption test on equal results: %s != %s", encoded, expectation)
	}
	return nil
}

func (vt *verificationTest) Run() error {
	result, err := shcrypto.VerifyEpochSecretKey(
		&vt.EpochSecretKey,
		&vt.EonPublicKey,
		vt.EpochID[:],
	)
	if err != nil {
		return fmt.Errorf("failed verification test with error: %s", err)
	}
	if result != vt.Expected {
		return fmt.Errorf("failed verification test on expected result: expected (%t) != result (%t)", vt.Expected, result)
	}
	return nil
}

func testMarshalingRoundtrip(tc *testCase) error {
	marshaled, err := json.Marshal(tc)
	if err != nil {
		return err
	}
	var unmarshaled testCase
	err = json.Unmarshal(marshaled, &unmarshaled)
	if err != nil {
		return err
	}

	roundtrip, err := json.Marshal(&unmarshaled)
	if err != nil {
		return err
	}
	if !bytes.Equal(marshaled, roundtrip) {
		println(len(marshaled))
		println(len(roundtrip))
		println("m:", string(marshaled))
		println("u:", string(roundtrip))
		return errors.New("roundtrip marshaling failed")
	}
	return nil
}
