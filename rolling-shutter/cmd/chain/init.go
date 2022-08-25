package chain

// This has been copied from tendermint's own init command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmtime "github.com/tendermint/tendermint/libs/time"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"

	"github.com/shutter-network/shutter/shuttermint/app"
)

var (
	logger                 = log.MustNewDefaultLogger(log.LogFormatPlain, log.LogLevelInfo, false)
	rootDir                = ""
	devMode                = false
	index                  = 0
	blockTime      float64 = 1.0
	genesisKeypers         = []string{}
	listenAddress          = "tcp://127.0.0.1:26657"
)

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a config file for a Shuttermint node",
		Args:  cobra.NoArgs,
		RunE:  initFiles,
	}
	cmd.PersistentFlags().StringVar(&rootDir, "root", rootDir, "root directory")
	cmd.PersistentFlags().BoolVar(
		&devMode,
		"dev",
		devMode,
		"turn on devmode (disables validator set changes)",
	)
	cmd.PersistentFlags().IntVar(&index, "index", index, "keyper index")
	cmd.PersistentFlags().Float64Var(&blockTime, "blocktime", blockTime, "block time in seconds")
	cmd.PersistentFlags().StringSliceVar(
		&genesisKeypers,
		"genesis-keyper",
		nil,
		"genesis keyper address",
	)
	cmd.PersistentFlags().StringVar(
		&listenAddress, "listen-address", listenAddress, "RCP listen address, "+
			"default: tcp://127.0.0.1:26657",
	)
	cmd.MarkPersistentFlagRequired("genesis-keyper")
	cmd.MarkPersistentFlagRequired("root")
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

func initFiles(_ *cobra.Command, _ []string) error {
	keypers := []common.Address{}

	for _, a := range genesisKeypers {
		if !common.IsHexAddress(a) {
			return errors.Errorf("--genesis-keyper argument '%s' is not an address", a)
		}
		keypers = append(keypers, common.HexToAddress(a))
	}

	config := cfg.DefaultConfig()
	config.LogLevel = log.LogLevelError
	config.RPC.ListenAddress = listenAddress

	scaleToBlockTime(config, blockTime)
	keyper0RPCAddress := config.RPC.ListenAddress
	rpcAddress, err := adjustPort(keyper0RPCAddress, index)
	if err != nil {
		return err
	}
	config.RPC.ListenAddress = rpcAddress

	keyper0P2PAddress := config.P2P.ListenAddress
	p2pAddress, err := adjustPort(keyper0P2PAddress, index)
	if err != nil {
		return err
	}
	config.P2P.ListenAddress = p2pAddress

	config.P2P.AllowDuplicateIP = true
	config.Mode = cfg.ModeValidator
	config.SetRoot(rootDir)
	if err := config.ValidateBasic(); err != nil {
		return errors.Wrap(err, "error in config file")
	}
	cfg.EnsureRoot(config.RootDir)

	// EnsureRoot also write the config file but with the default config. We want our own, so
	// let's overwrite it.
	err = cfg.WriteConfigFile(rootDir, config)
	if err != nil {
		return err
	}
	appState := app.NewGenesisAppState(keypers, (2*len(keypers)+2)/3)

	return initFilesWithConfig(config, appState)
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

func initFilesWithConfig(config *cfg.Config, appState app.GenesisAppState) error {
	// private validator
	privValKeyFile := config.PrivValidator.KeyFile()
	privValStateFile := config.PrivValidator.StateFile()
	var pv *privval.FilePV
	var err error
	if tmos.FileExists(privValKeyFile) {
		pv, err = privval.LoadFilePV(privValKeyFile, privValStateFile)
		if err != nil {
			return err
		}
		logger.Info(
			"Found private validator", "keyFile", privValKeyFile,
			"stateFile", privValStateFile,
		)
	} else {
		pv, err = privval.GenFilePV(privValKeyFile, privValStateFile, types.ABCIPubKeyTypeEd25519)
		if err != nil {
			return err
		}
		pv.Save()
		logger.Info(
			"Generated private validator", "keyFile", privValKeyFile,
			"stateFile", privValStateFile,
		)
	}

	nodeKeyFile := config.NodeKeyFile()
	if tmos.FileExists(nodeKeyFile) {
		logger.Info("Found node key", "path", nodeKeyFile)
	} else {
		nodeid, err := config.LoadOrGenNodeKeyID()
		if err != nil {
			return err
		}
		idpath := nodeKeyFile + ".id"
		err = os.WriteFile(idpath, []byte(nodeid), 0o755)
		if err != nil {
			return errors.Wrapf(err, "Could not write to %s", idpath)
		}

		logger.Info("Generated node key", "path", nodeKeyFile, "id", nodeid)
	}

	// genesis file
	genFile := config.GenesisFile()
	if tmos.FileExists(genFile) {
		logger.Info("Found genesis file", "path", genFile)
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
		genDoc.Validators = []types.GenesisValidator{
			{
				Address: pubKey.Address(),
				PubKey:  pubKey,
				Power:   10,
			},
		}

		if err := genDoc.SaveAs(genFile); err != nil {
			return err
		}
		logger.Info("Generated genesis file", "path", genFile)
	}
	a := app.NewShutterApp()
	a.Gobpath = filepath.Join(config.DBDir(), "shutter.gob")
	a.DevMode = devMode
	err = a.PersistToDisk()
	if err != nil {
		return err
	}

	return nil
}
