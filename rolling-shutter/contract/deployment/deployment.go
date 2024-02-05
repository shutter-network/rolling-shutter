// Package deployment provides mainly two structs: `Deployments` and `Contracts`. `Deployment`
// gathers information about a set of deployed contracts, like addresses and ABIs. It can be
// loaded from a deployment directory filled by hardhat-deploy. `Contracts` enriches the
// deployment data with abigen's contract bindings: For each known contract it has a bound
// contract instance and event types for all events.
package deployment

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"

	obscollator "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/collator"
	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/eventsyncer"
)

const chainIDFileName = ".chainId"

// Contracts groups all contracts the system interacts with as well as related information such as
// addresses, ABIs, and event types.
type Contracts struct {
	Client      *ethclient.Client
	Deployments *Deployments

	KeypersConfigsList           *contract.KeypersConfigsList
	KeypersConfigsListDeployment *Deployment
	KeypersConfigsListNewConfig  *eventsyncer.EventType

	CollatorConfigsList           *contract.CollatorConfigsList
	CollatorConfigsListDeployment *Deployment
	CollatorConfigsListNewConfig  *eventsyncer.EventType

	Keypers                     *contract.AddrsSeq
	KeypersDeployment           *Deployment
	KeypersAdded                *eventsyncer.EventType
	KeypersAppended             *eventsyncer.EventType
	KeypersOwnershipTransferred *eventsyncer.EventType

	Collators                     *contract.AddrsSeq
	CollatorsDeployment           *Deployment
	CollatorsAdded                *eventsyncer.EventType
	CollatorsAppended             *eventsyncer.EventType
	CollatorsOwnershipTransferred *eventsyncer.EventType
}

// Deployments contains information about all deployed contracts loaded from a deployment
// directory.
type Deployments struct {
	ChainID     uint64
	Deployments map[string]*Deployment
}

// Deployment contains information about a single deployed contract.
type Deployment struct {
	ChainID           uint64
	Name              string
	Address           common.Address
	ABI               abi.ABI
	DeployBlockNumber uint64
}

type deploymentJSON struct {
	Address common.Address
	ABI     []interface{}
	Receipt receiptJSON
}

type receiptJSON struct {
	BlockNumber uint64
}

func NewContracts(client *ethclient.Client, deploymentDir string) (*Contracts, error) {
	deployments, err := LoadDeployments(deploymentDir)
	if err != nil {
		return nil, err
	}
	c := &Contracts{
		Client:      client,
		Deployments: deployments,
	}
	if err := c.initKeypersConfigsList(); err != nil {
		return nil, err
	}
	if err := c.initCollatorConfigsList(); err != nil {
		return nil, err
	}
	if err := c.initKeypers(); err != nil {
		return nil, err
	}
	if err := c.initCollator(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Contracts) initKeypersConfigsList() error {
	d, err := c.getDeployment("KeyperConfig")
	if err != nil {
		return err
	}
	c.KeypersConfigsListDeployment = d
	c.KeypersConfigsList, err = contract.NewKeypersConfigsList(d.Address, c.Client)
	if err != nil {
		return err
	}
	boundContract := bind.NewBoundContract(d.Address, d.ABI, c.Client, c.Client, c.Client)
	c.KeypersConfigsListNewConfig = &eventsyncer.EventType{
		Contract:        boundContract,
		Address:         d.Address,
		FromBlockNumber: d.DeployBlockNumber,
		ABI:             d.ABI,
		Name:            "NewConfig",
		Type:            reflect.TypeOf(contract.KeypersConfigsListNewConfig{}),
	}
	return nil
}

func (c *Contracts) initCollatorConfigsList() error {
	d, err := c.getDeployment("CollatorConfig")
	if err != nil {
		return err
	}
	c.CollatorConfigsListDeployment = d
	c.CollatorConfigsList, err = contract.NewCollatorConfigsList(d.Address, c.Client)
	if err != nil {
		return err
	}
	boundContract := bind.NewBoundContract(d.Address, d.ABI, c.Client, c.Client, c.Client)
	c.CollatorConfigsListNewConfig = &eventsyncer.EventType{
		Contract:        boundContract,
		FromBlockNumber: d.DeployBlockNumber,
		Address:         d.Address,
		ABI:             d.ABI,
		Name:            "NewConfig",
		Type:            reflect.TypeOf(contract.CollatorConfigsListNewConfig{}),
	}
	return nil
}

func (c *Contracts) initKeypers() error {
	d, err := c.getDeployment("Keypers")
	if err != nil {
		return err
	}
	c.KeypersDeployment = d
	c.Keypers, err = contract.NewAddrsSeq(d.Address, c.Client)
	if err != nil {
		return err
	}
	kprHandler := &obskeyper.Handler{
		KeyperContract: c.Keypers,
	}
	c.KeypersConfigsListNewConfig.Handler = eventsyncer.MakeHandler(kprHandler.PutDB)

	boundContract := bind.NewBoundContract(d.Address, d.ABI, c.Client, c.Client, c.Client)
	c.KeypersAdded = &eventsyncer.EventType{
		FromBlockNumber: d.DeployBlockNumber,
		Contract:        boundContract,
		Address:         d.Address,
		ABI:             d.ABI,
		Name:            "Added",
		Type:            reflect.TypeOf(contract.AddrsSeqAdded{}),
	}
	c.KeypersAppended = &eventsyncer.EventType{
		FromBlockNumber: d.DeployBlockNumber,
		Contract:        boundContract,
		Address:         d.Address,
		ABI:             d.ABI,
		Name:            "Appended",
		Type:            reflect.TypeOf(contract.AddrsSeqAppended{}),
	}
	c.KeypersOwnershipTransferred = &eventsyncer.EventType{
		FromBlockNumber: d.DeployBlockNumber,
		Contract:        boundContract,
		Address:         d.Address,
		ABI:             d.ABI,
		Name:            "OwnershipTransferred",
		Type:            reflect.TypeOf(contract.AddrsSeqOwnershipTransferred{}),
	}
	return nil
}

func (c *Contracts) initCollator() error {
	d, err := c.getDeployment("Collator")
	if err != nil {
		return err
	}
	c.CollatorsDeployment = d
	c.Collators, err = contract.NewAddrsSeq(d.Address, c.Client)
	if err != nil {
		return err
	}
	cltHandler := &obscollator.Handler{
		CollatorContract: c.Collators,
	}
	c.CollatorConfigsListNewConfig.Handler = eventsyncer.MakeHandler(cltHandler.PutDB)

	boundContract := bind.NewBoundContract(d.Address, d.ABI, c.Client, c.Client, c.Client)
	c.CollatorsAdded = &eventsyncer.EventType{
		FromBlockNumber: d.DeployBlockNumber,
		Contract:        boundContract,
		Address:         d.Address,
		ABI:             d.ABI,
		Name:            "Added",
		Type:            reflect.TypeOf(contract.AddrsSeqAdded{}),
	}
	c.CollatorsAppended = &eventsyncer.EventType{
		FromBlockNumber: d.DeployBlockNumber,
		Contract:        boundContract,
		Address:         d.Address,
		ABI:             d.ABI,
		Name:            "Appended",
		Type:            reflect.TypeOf(contract.AddrsSeqAppended{}),
	}
	c.CollatorsOwnershipTransferred = &eventsyncer.EventType{
		FromBlockNumber: d.DeployBlockNumber,
		Contract:        boundContract,
		Address:         d.Address,
		ABI:             d.ABI,
		Name:            "OwnershipTransferred",
		Type:            reflect.TypeOf(contract.AddrsSeqOwnershipTransferred{}),
	}
	return nil
}

func (c *Contracts) getDeployment(name string) (*Deployment, error) {
	d, ok := c.Deployments.Deployments[name]
	if !ok {
		return nil, errors.Errorf("no deployment of %s contract found", name)
	}
	return d, nil
}

func LoadDeployments(dir string) (*Deployments, error) {
	chainID, err := LoadChainID(dir)
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read deployments directory at %s", dir)
	}
	deploymentFiles := []fs.DirEntry{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if file.Name() == chainIDFileName {
			continue
		}
		if strings.ToLower(filepath.Ext(file.Name())) != ".json" {
			continue
		}
		deploymentFiles = append(deploymentFiles, file)
	}

	deployments := Deployments{
		ChainID:     chainID,
		Deployments: make(map[string]*Deployment),
	}
	for _, file := range deploymentFiles {
		path := filepath.Join(dir, file.Name())
		deployment, err := LoadDeployment(path, chainID)
		if err != nil {
			return nil, err
		}
		deployments.Deployments[deployment.Name] = deployment
	}

	return &deployments, nil
}

func LoadDeployment(path string, chainID uint64) (*Deployment, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open deployment file at %s", path)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load deployment file at %s", path)
	}
	var parsedDeployment deploymentJSON
	if err := json.Unmarshal(data, &parsedDeployment); err != nil {
		return nil, errors.Wrapf(err, "failed to parse deployment file at %s", path)
	}

	encodedABI, err := json.Marshal(parsedDeployment.ABI)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encode ABI in deployment file at %s", path)
	}
	parsedABI, err := abi.JSON(bytes.NewReader(encodedABI))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse ABI in deployment file at %s", path)
	}

	name := contractNameFromPath(path)

	return &Deployment{
		ChainID:           chainID,
		Name:              name,
		Address:           parsedDeployment.Address,
		ABI:               parsedABI,
		DeployBlockNumber: parsedDeployment.Receipt.BlockNumber,
	}, nil
}

func LoadChainID(dir string) (uint64, error) {
	path := filepath.Join(dir, chainIDFileName)
	file, err := os.Open(path)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to open chain id file at %s", path)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to load chain id file at %s", path)
	}

	chainIDStr := strings.TrimSpace(string(data))
	chainID, err := strconv.ParseInt(chainIDStr, 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse chain id in %s", path)
	}
	if chainID < 0 {
		return 0, errors.Wrapf(err, "chain id %d found in %s is invalid", chainID, path)
	}

	return uint64(chainID), nil
}

func contractNameFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(path)
	return strings.TrimSuffix(base, ext)
}
