package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	appconfig "github.com/vulpemventures/ocean/internal/app-config"
	"github.com/vulpemventures/ocean/internal/config"
	neutrino_scanner "github.com/vulpemventures/ocean/internal/infrastructure/blockchain-scanner/neutrino"
	"github.com/vulpemventures/ocean/internal/interfaces"
	grpc_interface "github.com/vulpemventures/ocean/internal/interfaces/grpc"
	"github.com/vulpemventures/ocean/pkg/profiler"
)

var (
	// Build info.
	version string
	commit  string
	date    string

	// Config from env vars.
	dbType             = config.GetString(config.DatabaseTypeKey)
	bcScannerType      = config.GetString(config.BlockchainScannerTypeKey)
	logLevel           = config.GetInt(config.LogLevelKey)
	datadir            = config.GetDatadir()
	port               = config.GetInt(config.PortKey)
	profilerPort       = config.GetInt(config.ProfilerPortKey)
	network            = config.GetNetwork()
	noTLS              = config.GetBool(config.NoTLSKey)
	noProfiler         = config.GetBool(config.NoProfilerKey)
	dbDir              = filepath.Join(datadir, config.DbLocation)
	tlsDir             = filepath.Join(datadir, config.TLSLocation)
	scannerDir         = filepath.Join(datadir, config.ScannerLocation)
	profilerDir        = filepath.Join(datadir, config.ProfilerLocation)
	filtersDir         = filepath.Join(scannerDir, "filters")
	blockHeadersDir    = filepath.Join(scannerDir, "headers")
	tlsExtraIPs        = config.GetStringSlice(config.TLSExtraIPKey)
	tlsExtraDomains    = config.GetStringSlice(config.TLSExtraDomainKey)
	statsInterval      = time.Duration(config.GetInt(config.StatsIntervalKey)) * time.Second
	nodePeers          = config.GetStringSlice(config.NodePeersKey)
	utxoExpiryDuration = time.Duration(config.GetInt(config.UtxoExpiryDurationKey))
	rootPath           = config.GetRootPath()
)

func main() {
	log.SetLevel(log.Level(logLevel))

	if profilerEnabled := !noProfiler; profilerEnabled {
		profilerSvc, err := profiler.NewService(profiler.ServiceOpts{
			Port:          profilerPort,
			StatsInterval: statsInterval,
			Datadir:       profilerDir,
		})
		if err != nil {
			log.WithError(err).Fatal("profiler: error while starting")
		}

		profilerSvc.Start()
		defer func() {
			profilerSvc.Stop()
		}()
	}

	net := network.Name
	if net == "regtest" {
		net = "nigiri"
	}

	bcScannerConfig := neutrino_scanner.NodeServiceArgs{
		Network:             net,
		FiltersDatadir:      filtersDir,
		BlockHeadersDatadir: blockHeadersDir,
		Peers:               nodePeers,
	}
	serviceCfg := grpc_interface.ServiceConfig{
		Port:         port,
		NoTLS:        noTLS,
		TLSLocation:  tlsDir,
		ExtraIPs:     tlsExtraIPs,
		ExtraDomains: tlsExtraDomains,
	}
	appCfg := &appconfig.AppConfig{
		Version:                 version,
		Commit:                  commit,
		Date:                    date,
		RootPath:                rootPath,
		Network:                 network,
		UtxoExpiryDuration:      utxoExpiryDuration * time.Second,
		RepoManagerType:         dbType,
		BlockchainScannerType:   bcScannerType,
		RepoManagerConfig:       dbDir,
		BlockchainScannerConfig: bcScannerConfig,
	}

	serviceManager, err := interfaces.NewGrpcServiceManager(serviceCfg, appCfg)
	if err != nil {
		log.WithError(err).Fatal("service: error while initializing")
	}
	defer func() {
		serviceManager.Service.Stop()
	}()

	serviceManager.Service.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan
}
