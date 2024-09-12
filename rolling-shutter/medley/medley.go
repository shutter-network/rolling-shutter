// Package medley provides some functions that may be useful in various parts of shutter
package medley

import (
	"context"
	"errors"
	"math"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	pkgErrors "github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const receiptPollInterval = 500 * time.Millisecond

var errAddressNotFound = errors.New("address not found")

var ErrShutdownRequested = errors.New("shutdown requested from user")

// FindAddressIndex returns the index of the given address inside the slice of addresses or returns
// an error, if the slice does not contain the given address.
func FindAddressIndex(addresses []common.Address, addr common.Address) (int, error) {
	for i, a := range addresses {
		if a == addr {
			return i, nil
		}
	}
	return -1, pkgErrors.WithStack(errAddressNotFound)
}

// Sleep pauses the current goroutine for the given duration.
func Sleep(ctx context.Context, d time.Duration) {
	if d <= 0 {
		return
	}
	select {
	case <-ctx.Done():
		return
	case <-time.After(d):
	}
}

// EnsureUniqueAddresses makes sure the slice of addresses doesn't contain duplicate addresses.
func EnsureUniqueAddresses(addrs []common.Address) error {
	seen := make(map[common.Address]struct{})
	for _, a := range addrs {
		if _, ok := seen[a]; ok {
			return pkgErrors.Errorf("duplicate address: %s", a.Hex())
		}
		seen[a] = struct{}{}
	}
	return nil
}

func normName(s string) string {
	return strings.ToUpper(strings.ReplaceAll(s, "-", "_"))
}

const depthToFoldChildPrefixes = 2

func bindFlagsToRootCommand(cmd *cobra.Command) (string, int, error) {
	var (
		prefix string
		depth  int
		err    error
	)

	postfix := normName(cmd.Name())
	if cmd.HasParent() {
		prefix, depth, err = bindFlagsToRootCommand(cmd.Parent())
		if err != nil {
			return prefix, depth, err
		}
		depth++
		// If the current depth is at or above the folding depth,
		// all children flags will be "folded" under the same prefix.
		// E.g. with depthToFoldChildPrefix = 1, all args will be
		// settable with `ROLLING_SHUTTER_<ARG>` env vars,
		// while a depth of 2 would result in
		// `ROLLING_SHUTTER_<ARG>` and `ROLLING_SHUTTER_<SUBCMD>_<ARG>`
		// vars and so on
		if depth < depthToFoldChildPrefixes {
			prefix = prefix + "_" + postfix
		}
	} else {
		prefix = postfix
		depth = 0
	}

	lflg := cmd.LocalFlags()
	if lflg == nil {
		return prefix, depth, nil
	}
	flgs := pflag.NewFlagSet(prefix, pflag.ContinueOnError)
	lflg.VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" {
			return
		}
		envName := prefix + "_" + normName(f.Name)
		viper.BindEnv(f.Name, envName)
		if f.DefValue != "" {
			viper.SetDefault(f.Name, f.DefValue)
		}
		flgs.AddFlag(f)
	})
	// bind only the filtered flagset
	// (e.g. without "help" flag) to viper
	err = viper.BindPFlags(flgs)
	return prefix, depth, err
}

func BindFlags(cmd *cobra.Command) error {
	_, _, err := bindFlagsToRootCommand(cmd)
	return err
}

// ShowHelpAndExit shows the commands help message and exits the program with status 1.
func ShowHelpAndExit(cmd *cobra.Command, args []string) {
	_ = args
	_ = cmd.Help()
	os.Exit(1)
}

func Uint64ToInt64Safe(u uint64) (int64, error) {
	if u > math.MaxInt64 {
		return math.MaxInt64, errors.New("int64 overflow")
	}
	return int64(u), nil
}

func Int64ToUint64Safe(i int64) (uint64, error) {
	if i < 0 {
		return 0, errors.New("uint64 can't be negative")
	}
	return uint64(i), nil
}

func Int32ToUint64Safe(i int32) (uint64, error) {
	if i < 0 {
		return 0, errors.New("uint32 can't be negative")
	}
	return uint64(i), nil
}
