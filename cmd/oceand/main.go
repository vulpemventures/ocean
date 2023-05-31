package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	_ "net/http/pprof" // #nosec

	log "github.com/sirupsen/logrus"
	appconfig "github.com/vulpemventures/ocean/internal/app-config"
	"github.com/vulpemventures/ocean/internal/config"
	electrum_scanner "github.com/vulpemventures/ocean/internal/infrastructure/blockchain-scanner/electrum"
	postgresdb "github.com/vulpemventures/ocean/internal/infrastructure/storage/db/postgres"
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
	dbType             = config.GetString(config.DbTypeKey)
	bcScannerType      = config.GetString(config.BlockchainScannerTypeKey)
	logLevel           = config.GetInt(config.LogLevelKey)
	datadir            = config.GetDatadir()
	port               = config.GetInt(config.PortKey)
	profilerPort       = config.GetInt(config.ProfilerPortKey)
	network            = config.GetNetwork()
	noTLS              = config.GetBool(config.NoTLSKey)
	noProfiler         = config.GetBool(config.NoProfilerKey)
	tlsDir             = filepath.Join(datadir, config.TLSLocation)
	profilerDir        = filepath.Join(datadir, config.ProfilerLocation)
	electrumUrl        = config.GetString(config.ElectrumUrlKey)
	tlsExtraIPs        = config.GetStringSlice(config.TLSExtraIPKey)
	tlsExtraDomains    = config.GetStringSlice(config.TLSExtraDomainKey)
	statsInterval      = time.Duration(config.GetInt(config.StatsIntervalKey)) * time.Second
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

	bcScannerConfig := electrum_scanner.ServiceArgs{
		Addr:    electrumUrl,
		Network: network,
	}
	serviceCfg := grpc_interface.ServiceConfig{
		Port:         port,
		NoTLS:        noTLS,
		TLSLocation:  tlsDir,
		ExtraIPs:     tlsExtraIPs,
		ExtraDomains: tlsExtraDomains,
	}
	repoManagerConfig := dbConfigFromType()
	appCfg := &appconfig.AppConfig{
		Version:                 version,
		Commit:                  commit,
		Date:                    date,
		RootPath:                rootPath,
		Network:                 network,
		UtxoExpiryDuration:      utxoExpiryDuration * time.Second,
		RepoManagerType:         dbType,
		BlockchainScannerType:   bcScannerType,
		RepoManagerConfig:       repoManagerConfig,
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

func dbConfigFromType() interface{} {
	switch dbType {
	case "postgres":
		return postgresdb.DbConfig{
			DbUser:             dbUser,
			DbPassword:         dbPassword,
			DbHost:             dbHost,
			DbPort:             dbPort,
			DbName:             dbName,
			MigrationSourceURL: migrationSourceURL,
		}
	case "badger":
		return filepath.Join(datadir, "db")
	case "inmemory":
		fallthrough
	default:
		return nil
	}
}
