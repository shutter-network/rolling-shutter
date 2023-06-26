package chain

// This has been copied from tendermint's own init command

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
	cfg "github.com/tendermint/tendermint/config"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmtime "github.com/tendermint/tendermint/libs/time"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/app"
)

type Config struct {
	RootDir       string   `mapstructure:"root"`
	DevMode       bool     `mapstructure:"dev"`
	Index         int      `mapstructure:"index"`
	BlockTime     float64  `mapstructure:"blocktime"`
	GenesisKeyper []string `mapstructure:"genesis-keyper"`
	ListenAddress string   `mapstructure:"listen-address"`
}

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a config file for a Shuttermint node",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &Config{}
			if err := viper.Unmarshal(config); err != nil {
				return err
			}
			if len(config.GenesisKeyper) == 0 {
				return errors.New("required argument `genesis-keyper` not specified")
			}
			if config.RootDir == "" {
				return errors.New("required argument `root` not specified")
			}
			return initFiles(cmd, config, args)
		},
	}
	cmd.PersistentFlags().String("root", "", "root directory")
	cmd.PersistentFlags().Bool("dev", false, "turn on devmode (disables validator set changes)")
	cmd.PersistentFlags().Int("index", 0, "keyper index")
	cmd.PersistentFlags().Float64("blocktime", 1.0, "block time in seconds")
	cmd.PersistentFlags().StringSlice("genesis-keyper", nil, "genesis keyper address")
	cmd.PersistentFlags().String("listen-address", "tcp://127.0.0.1:26657", "tendermint listen address")
	return cmd
}

func scaleToBlockTime(config *cfg.Config, blockTime float64) {
	f := blockTime * float64(time.Second) / float64(config.Consensus.TimeoutCommit)
	scale := func(d *time.Duration) {
		*d = time.Duration(float64(*d) * f)
	}
	scale(&config.Consensus.TimeoutPropose)
	scale(&config.Consensus.TimeoutProposeDelta)
	scale(&config.Consensus.TimeoutPrevote)
	scale(&config.Consensus.TimeoutPrecommit)
	scale(&config.Consensus.TimeoutPrecommitDelta)
	scale(&config.Consensus.TimeoutCommit)
	scale(&config.RPC.TimeoutBroadcastTxCommit)
}

func getArgFromViper[T interface{}](getter func(string) T, name string, required bool) (T, error) {
	if !viper.IsSet(name) && required {
		var nullVal T
		return nullVal, errors.Errorf("required argument `%s` not set", name)
	}
	return getter(name), nil
}

func initFiles(_ *cobra.Command, config *Config, _ []string) error {
	keypers := []common.Address{}

	for _, a := range config.GenesisKeyper {
		if !common.IsHexAddress(a) {
			return errors.Errorf("--genesis-keyper argument '%s' is not an address", a)
		}
		keypers = append(keypers, common.HexToAddress(a))
	}

	tendermintCfg := cfg.DefaultConfig()
	tendermintCfg.LogLevel = tmlog.LogLevelError
	tendermintCfg.RPC.ListenAddress = config.ListenAddress

	scaleToBlockTime(tendermintCfg, config.BlockTime)
	keyper0RPCAddress := tendermintCfg.RPC.ListenAddress
	rpcAddress, err := adjustPort(keyper0RPCAddress, config.Index)
	if err != nil {
		return err
	}
	tendermintCfg.RPC.ListenAddress = rpcAddress

	keyper0P2PAddress := tendermintCfg.P2P.ListenAddress
	p2pAddress, err := adjustPort(keyper0P2PAddress, config.Index)
	if err != nil {
		return err
	}
	tendermintCfg.P2P.ListenAddress = p2pAddress

	tendermintCfg.P2P.AllowDuplicateIP = true
	tendermintCfg.Mode = cfg.ModeValidator

	tendermintCfg.SetRoot(config.RootDir)
	if err := tendermintCfg.ValidateBasic(); err != nil {
		return errors.Wrap(err, "error in config file")
	}
	cfg.EnsureRoot(tendermintCfg.RootDir)

	// EnsureRoot also write the config file but with the default config. We want our own, so
	// let's overwrite it.
	err = cfg.WriteConfigFile(config.RootDir, tendermintCfg)
	if err != nil {
		return err
	}
	appState := app.NewGenesisAppState(keypers, (2*len(keypers)+2)/3)

	return initFilesWithConfig(tendermintCfg, config, appState)
}

func adjustPort(address string, keyperIndex int) (string, error) {
	substrings := strings.Split(address, ":")
	if len(substrings) < 2 {
		return "", errors.Errorf("address %s does not contain port", address)
	}
	portStr := substrings[len(substrings)-1]
	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		return "", errors.Errorf("port %s is not an integer", portStr)
	}
	portIntAdjusted := portInt + keyperIndex*2
	portStrAdjusted := strconv.Itoa(portIntAdjusted)
	return strings.Join(substrings[:len(substrings)-1], ":") + ":" + portStrAdjusted, nil
}

func initFilesWithConfig(tendermintConfig *cfg.Config, config *Config, appState app.GenesisAppState) error {
	// private validator
	privValKeyFile := tendermintConfig.PrivValidator.KeyFile()
	privValStateFile := tendermintConfig.PrivValidator.StateFile()
	var pv *privval.FilePV
	var err error
	if tmos.FileExists(privValKeyFile) {
		pv, err = privval.LoadFilePV(privValKeyFile, privValStateFile)
		if err != nil {
			return err
		}
		log.Info().
			Str("privValKeyFile", privValKeyFile).
			Str("stateFile", privValStateFile).
			Msg("Found private validator")
	} else {
		pv, err = privval.GenFilePV(privValKeyFile, privValStateFile, types.ABCIPubKeyTypeEd25519)
		if err != nil {
			return err
		}
		pv.Save()
		log.Info().
			Str("privValKeyFile", privValKeyFile).
			Str("stateFile", privValStateFile).
			Msg("Generated private validator")
	}

	validatorPubKeyPath := filepath.Join(tendermintConfig.RootDir, "config", "priv_validator_pubkey.hex")
	validatorPublicKeyHex := hex.EncodeToString(pv.Key.PubKey.Bytes())
	err = os.WriteFile(validatorPubKeyPath, []byte(validatorPublicKeyHex), 0o644)
	if err != nil {
		return errors.Wrapf(err, "Could not write to %s", validatorPubKeyPath)
	}
	log.Info().Str("path", validatorPubKeyPath).Str("validatorPublicKey", validatorPublicKeyHex).Msg("Saved private validator publickey")

	nodeKeyFile := tendermintConfig.NodeKeyFile()
	if tmos.FileExists(nodeKeyFile) {
		log.Info().Str("path", nodeKeyFile).Msg("Found node key")
	} else {
		nodeid, err := tendermintConfig.LoadOrGenNodeKeyID()
		if err != nil {
			return err
		}
		idpath := nodeKeyFile + ".id"
		err = os.WriteFile(idpath, []byte(nodeid), 0o755)
		if err != nil {
			return errors.Wrapf(err, "Could not write to %s", idpath)
		}
		log.Info().Str("path", nodeKeyFile).Str("id", string(nodeid)).Msg("Generated node key")
	}

	// genesis file
	genFile := tendermintConfig.GenesisFile()
	if tmos.FileExists(genFile) {
		log.Info().Str("path", genFile).Msg("Found genesis file")
	} else {
		appStateBytes, err := amino.NewCodec().MarshalJSONIndent(appState, "", "    ")
		if err != nil {
			return err
		}
		genDoc := types.GenesisDoc{
			ChainID:         fmt.Sprintf("shutter-test-chain-%v", tmrand.Str(6)),
			GenesisTime:     tmtime.Now(),
			ConsensusParams: types.DefaultConsensusParams(),
			AppState:        appStateBytes,
		}
		pubKey, err := pv.GetPubKey(context.Background())
		if err != nil {
			return errors.Wrap(err, "can't get pubkey")
		}
		genDoc.Validators = []types.GenesisValidator{{
			Address: pubKey.Address(),
			PubKey:  pubKey,
			Power:   10,
		}}

		if err := genDoc.SaveAs(genFile); err != nil {
			return err
		}
		log.Info().Str("path", genFile).Msg("Generated genesis file")
	}
	a := app.NewShutterApp()
	a.Gobpath = filepath.Join(tendermintConfig.DBDir(), "shutter.gob")
	a.DevMode = config.DevMode
	err = a.PersistToDisk()
	if err != nil {
		return err
	}

	return nil
}
