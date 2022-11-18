package main

import (
	postgresdb "github.com/vulpemventures/ocean/internal/infrastructure/storage/db/postgres"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	_ "net/http/pprof" // #nosec

	log "github.com/sirupsen/logrus"
	appconfig "github.com/vulpemventures/ocean/internal/app-config"
	"github.com/vulpemventures/ocean/internal/config"
	elements_scanner "github.com/vulpemventures/ocean/internal/infrastructure/blockchain-scanner/elements"
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
	tlsDir             = filepath.Join(datadir, config.TLSLocation)
	scannerDir         = filepath.Join(datadir, config.ScannerLocation)
	profilerDir        = filepath.Join(datadir, config.ProfilerLocation)
	filtersDir         = filepath.Join(scannerDir, "filters")
	blockHeadersDir    = filepath.Join(scannerDir, "headers")
	esploraUrl         = config.GetString(config.EsploraUrlKey)
	tlsExtraIPs        = config.GetStringSlice(config.TLSExtraIPKey)
	tlsExtraDomains    = config.GetStringSlice(config.TLSExtraDomainKey)
	statsInterval      = time.Duration(config.GetInt(config.StatsIntervalKey)) * time.Second
	nodeRpcAddr        = config.GetString(config.ElementsNodeRpcAddrKey)
	utxoExpiryDuration = time.Duration(config.GetInt(config.UtxoExpiryDurationKey))
	rootPath           = config.GetRootPath()
	dbUser             = config.GetString(config.DbUserKey)
	dbPassword         = config.GetString(config.DbPassKey)
	dbHost             = config.GetString(config.DbHostKey)
	dbPort             = config.GetInt(config.DbPortKey)
	dbName             = config.GetString(config.DbNameKey)
	migrationSourceURL = config.GetString(config.DbMigrationPath)
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
		defer profilerSvc.Stop()
	}

	bcScannerConfig := elements_scanner.ServiceArgs{
		RpcAddr:             nodeRpcAddr,
		Network:             network.Name,
		FiltersDatadir:      filtersDir,
		BlockHeadersDatadir: blockHeadersDir,
		EsploraUrl:          esploraUrl,
	}
	serviceCfg := grpc_interface.ServiceConfig{
		Port:         port,
		NoTLS:        noTLS,
		TLSLocation:  tlsDir,
		ExtraIPs:     tlsExtraIPs,
		ExtraDomains: tlsExtraDomains,
	}
	appCfg := &appconfig.AppConfig{
		Version:               version,
		Commit:                commit,
		Date:                  date,
		RootPath:              rootPath,
		Network:               network,
		UtxoExpiryDuration:    utxoExpiryDuration * time.Second,
		RepoManagerType:       dbType,
		BlockchainScannerType: bcScannerType,
		RepoManagerConfig: postgresdb.DbConfig{
			DbUser:             dbUser,
			DbPassword:         dbPassword,
			DbHost:             dbHost,
			DbPort:             dbPort,
			DbName:             dbName,
			MigrationSourceURL: migrationSourceURL,
		},
		BlockchainScannerConfig: bcScannerConfig,
	}

	serviceManager, err := interfaces.NewGrpcServiceManager(serviceCfg, appCfg)
	if err != nil {
		log.WithError(err).Fatal("service: error while initializing")
	}

	if err := serviceManager.Service.Start(); err != nil {
		log.WithError(err).Fatal("service: error while starting")
	}
	defer serviceManager.Service.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan
}
