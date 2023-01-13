package pgtest

import (
	"context"
	"github.com/stretchr/testify/suite"
	"github.com/vulpemventures/go-elements/network"
	appconfig "github.com/vulpemventures/ocean/internal/app-config"
	electrum_scanner "github.com/vulpemventures/ocean/internal/infrastructure/blockchain-scanner/electrum"
	postgresdb "github.com/vulpemventures/ocean/internal/infrastructure/storage/db/postgres"
	"github.com/vulpemventures/ocean/internal/interfaces"
	grpc_interface "github.com/vulpemventures/ocean/internal/interfaces/grpc"
	dbtest "github.com/vulpemventures/ocean/test/testutil"
	"time"
)

var (
	mnemonic = "tell tell behave duty duty file file bacon design element attract era enter turtle cycle foot bone sand elevator useless release affair giggle engage"
	password = "password"

	ctx     = context.Background()
	grpcSvc *interfaces.ServiceManager
)

type GrpcDbTestSuite struct {
	suite.Suite
}

func (g *GrpcDbTestSuite) SetupSuite() {
	bcScannerConfig := electrum_scanner.ServiceArgs{
		Addr: "tcp://localhost:50001",
	}
	serviceCfg := grpc_interface.ServiceConfig{
		Port:  18000,
		NoTLS: true,
	}
	appCfg := &appconfig.AppConfig{
		RootPath:              "m/84'/1'",
		Network:               &network.Regtest,
		UtxoExpiryDuration:    360 * time.Second,
		RepoManagerType:       "postgres",
		BlockchainScannerType: "electrum",
		RepoManagerConfig: postgresdb.DbConfig{
			DbUser:     "root",
			DbPassword: "secret",
			DbHost:     "127.0.0.1",
			DbPort:     5432,
			DbName:     "oceand-db-test",
			MigrationSourceURL: "file://../.." +
				"/internal/infrastructure/storage/db/postgres/migration",
		},
		BlockchainScannerConfig: bcScannerConfig,
	}

	svc, err := interfaces.NewGrpcServiceManager(serviceCfg, appCfg)
	if err != nil {
		g.FailNow(err.Error())
	}

	if err := svc.Service.Start(); err != nil {
		g.FailNow(err.Error())
	}

	grpcSvc = svc

	if err := dbtest.SetupDB(); err != nil {
		g.FailNow(err.Error())
	}
}

func (g *GrpcDbTestSuite) TearDownSuite() {
	if err := dbtest.TruncateDB(); err != nil {
		g.FailNow(err.Error())
	}
}

func (g *GrpcDbTestSuite) BeforeTest(suiteName, testName string) {
	if err := dbtest.TruncateDB(); err != nil {
		g.FailNow(err.Error())
	}
}

func (g *GrpcDbTestSuite) AfterTest(suiteName, testName string) {
}
