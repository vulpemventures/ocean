package appconfig

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/ocean/internal/config"
	"github.com/vulpemventures/ocean/internal/core/application"
	"github.com/vulpemventures/ocean/internal/core/ports"
	electrum_scanner "github.com/vulpemventures/ocean/internal/infrastructure/blockchain-scanner/electrum"
	elements_scanner "github.com/vulpemventures/ocean/internal/infrastructure/blockchain-scanner/elements"
	neutrino_scanner "github.com/vulpemventures/ocean/internal/infrastructure/blockchain-scanner/neutrino"
	dbbadger "github.com/vulpemventures/ocean/internal/infrastructure/storage/db/badger"
	"github.com/vulpemventures/ocean/internal/infrastructure/storage/db/inmemory"
	postgresdb "github.com/vulpemventures/ocean/internal/infrastructure/storage/db/postgres"
	path "github.com/vulpemventures/ocean/pkg/wallet/derivation-path"
)

// AppConfig is the struct holding all configuration options for
// every application service (wallet, account, transaction and notification).
// This data structure acts also as a factory of the mentioned application
// services and the portable services used by them.
// Public config args:
//   - RootPath - (optional) Wallet root HD path (defaults to m/84'/0').
//   - Network - (required) The Liquid network (mainnet, testnet, regtest).
//   - UtxoExpiryDuration - (required) The duration in seconds for the app service to wait until unlocking one or more previously locked utxo.
//   - RepoManagerType - (required) One of the supported repository manager types.
//   - BlockchainScannerType - (required) One of the supported blockchain scanner types.
//   - RepoManagerConfig - (optional) Custom config args for the repository manager based on its type.
//   - BlockchainScannerConfig - (optional) Custom config args for the blockchain scanner based on its type.
type AppConfig struct {
	Version string
	Commit  string
	Date    string

	RootPath           string
	Network            *network.Network
	UtxoExpiryDuration time.Duration
	DustAmount         uint64

	RepoManagerType         string
	BlockchainScannerType   string
	RepoManagerConfig       interface{}
	BlockchainScannerConfig interface{}

	rm         ports.RepoManager
	bcs        ports.BlockchainScanner
	walletSvc  *application.WalletService
	accountSvc *application.AccountService
	txSvc      *application.TransactionService
	notifySvc  *application.NotificationService
}

func (c *AppConfig) Validate() error {
	if c.Network == nil {
		return fmt.Errorf("missing network")
	}
	if c.UtxoExpiryDuration == 0 {
		return fmt.Errorf("missing utxo expiry duration")
	}
	if c.DustAmount == 0 {
		return fmt.Errorf("missing dust amount threshold")
	}
	if len(c.RepoManagerType) == 0 {
		return fmt.Errorf("missing repo manager type")
	}
	if _, ok := config.SupportedDbs[c.RepoManagerType]; !ok {
		return fmt.Errorf(
			"repo manager type not supported, must be one of: %s",
			config.SupportedDbs,
		)
	}
	if len(c.BlockchainScannerType) == 0 {
		return fmt.Errorf("missing blockchain scanner type")
	}
	if _, ok := config.SupportedBcScanners[c.BlockchainScannerType]; !ok {
		return fmt.Errorf(
			"blockchain scanner type not supported, must be one of: %s",
			config.SupportedBcScanners,
		)
	}
	if _, err := c.repoManager(); err != nil {
		return err
	}
	if _, err := c.bcScanner(); err != nil {
		return err
	}
	if c.RootPath == "" {
		return fmt.Errorf("missing root path")
	}
	if _, err := path.ParseRootDerivationPath(c.RootPath); err != nil {
		return err
	}

	return nil
}

func (c *AppConfig) RepoManager() ports.RepoManager {
	return c.rm
}

func (c *AppConfig) BlockchainScanner() ports.BlockchainScanner {
	return c.bcs
}

func (c *AppConfig) WalletService() *application.WalletService {
	return c.walletService()
}

func (c *AppConfig) AccountService() *application.AccountService {
	return c.accountService()
}

func (c *AppConfig) TransactionService() *application.TransactionService {
	return c.transactionService()
}

func (c *AppConfig) NotificationService() *application.NotificationService {
	return c.notificationService()
}

func (c *AppConfig) repoManager() (ports.RepoManager, error) {
	if c.rm != nil {
		return c.rm, nil
	}

	switch c.RepoManagerType {
	case "inmemory":
		c.rm = inmemory.NewRepoManager()
		return c.rm, nil
	case "badger":
		if c.RepoManagerConfig == nil {
			return nil, fmt.Errorf("missing repo manager config args")
		}
		datadir, ok := c.RepoManagerConfig.(string)
		if !ok {
			return nil, fmt.Errorf("invalid repo manager config type, must be string")
		}
		rm, err := dbbadger.NewRepoManager(datadir, log.New())
		if err != nil {
			return nil, err
		}
		c.rm = rm
		return c.rm, nil
	case "postgres":
		dbConfig, ok := c.RepoManagerConfig.(postgresdb.DbConfig)
		if !ok {
			return nil, fmt.Errorf("invalid repo manager config type, must be postgresdb.DbConfig")
		}

		rm, err := postgresdb.NewRepoManager(dbConfig)
		if err != nil {
			return nil, err
		}

		c.rm = rm
		return c.rm, nil
	default:
		return nil, fmt.Errorf("unknown repo manager type")
	}
}

func (c *AppConfig) bcScanner() (ports.BlockchainScanner, error) {
	if c.bcs != nil {
		return c.bcs, nil
	}

	switch c.BlockchainScannerType {
	case "neutrino":
		if c.BlockchainScannerConfig == nil {
			return nil, fmt.Errorf("missing blockchain scanner config args")
		}
		args, ok := c.BlockchainScannerConfig.(neutrino_scanner.NodeServiceArgs)
		if !ok {
			return nil, fmt.Errorf(
				"invalid blockchain scanner config type, must be " +
					"neutrino_scanner.NodeServiceArgs",
			)
		}
		bcs, err := neutrino_scanner.NewNeutrinoScanner(args)
		if err != nil {
			return nil, err
		}
		c.bcs = bcs
		return c.bcs, nil
	case "elements":
		if c.BlockchainScannerConfig == nil {
			return nil, fmt.Errorf("missing blockchain scanner config args")
		}
		args, ok := c.BlockchainScannerConfig.(elements_scanner.ServiceArgs)
		if !ok {
			return nil, fmt.Errorf(
				"invalid blockchain scanner config type, must be " +
					"elements_scanner.ServiceArgs",
			)
		}
		bcs, err := elements_scanner.NewElementsScanner(args)
		if err != nil {
			return nil, err
		}
		c.bcs = bcs
		return c.bcs, nil
	case "electrum":
		if c.BlockchainScannerConfig == nil {
			return nil, fmt.Errorf("missing blockchain scanner config args")
		}
		args, ok := c.BlockchainScannerConfig.(electrum_scanner.ServiceArgs)
		if !ok {
			return nil, fmt.Errorf(
				"invalid blockchain scanner config type, must be " +
					"elements_scanner.ServiceArgs",
			)
		}
		bcs, err := electrum_scanner.NewService(args)
		if err != nil {
			return nil, err
		}
		c.bcs = bcs
		return c.bcs, nil
	default:
		return nil, fmt.Errorf("unknown blockchain scanner type")
	}
}

func (c *AppConfig) walletService() *application.WalletService {
	if c.walletSvc != nil {
		return c.walletSvc
	}

	rm, _ := c.repoManager()
	bcs, _ := c.bcScanner()
	c.walletSvc = application.NewWalletService(
		rm, bcs, c.RootPath, c.Network, c.buildInfo(),
	)
	return c.walletSvc
}

func (c *AppConfig) accountService() *application.AccountService {
	if c.accountSvc != nil {
		return c.accountSvc
	}

	rm, _ := c.repoManager()
	bcs, _ := c.bcScanner()
	c.accountSvc = application.NewAccountService(rm, bcs)
	return c.accountSvc
}

func (c *AppConfig) transactionService() *application.TransactionService {
	if c.txSvc != nil {
		return c.txSvc
	}

	rm, _ := c.repoManager()
	bcs, _ := c.bcScanner()
	c.txSvc = application.NewTransactionService(
		rm, bcs, c.Network, c.UtxoExpiryDuration, c.DustAmount,
	)
	return c.txSvc
}

func (c *AppConfig) notificationService() *application.NotificationService {
	if c.notifySvc != nil {
		return c.notifySvc
	}

	rm, _ := c.repoManager()
	bcs, _ := c.bcScanner()
	c.notifySvc = application.NewNotificationService(rm, bcs)
	return c.notifySvc
}

func (c *AppConfig) buildInfo() application.BuildInfo {
	version := "dev"
	if c.Version != "" {
		version = c.Version
	}
	commit := "none"
	if c.Commit != "" {
		commit = c.Commit
	}
	date := "unknown"
	if c.Date != "" {
		date = c.Date
	}
	return application.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
}
