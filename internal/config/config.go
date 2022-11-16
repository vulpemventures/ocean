package config

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/spf13/viper"
	"github.com/vulpemventures/go-elements/network"
)

const (
	// DatadirKey is the key to customize the ocean datadir.
	DatadirKey = "DATADIR"
	// DatabaseTypeKey is the key to customize the type of database to use.
	DatabaseTypeKey = "DATABASE_TYPE"
	// BlockchainScannerTypeKey is the key to customize the type of blockchain
	// scanner to use.
	BlockchainScannerTypeKey = "BLOCKCHAIN_SCANNER_TYPE"
	// PortKey is the key to customize the port where the wallet will be listening to.
	PortKey = "PORT"
	// ProfilerPortKey is the key to customize the port where the profiler will
	// be listening to.
	ProfilerPortKey = "PROFILER_PORT"
	// NetworkKey is the key to customize the Liquid network.
	NetworkKey = "NETWORK"
	// NativeAssetKey is the key to customize the native LBTC asset of the Liquid
	// network. Should be used only for testing purposes.
	NativeAssetKey = "NATIVE_ASSET"
	// LogLevelKey is the key to customize the log level to catch more specific
	// or more high level logs.
	LogLevelKey = "LOG_LEVEL"
	// TLSExtraIPKey is the key to bind one or more public IPs to the TLS key pair.
	// Should be used only when enabling TLS.
	TLSExtraIPKey = "TLS_EXTRA_IP"
	// TLSExtraDomainKey is the key to bind one or more public dns domains to the
	// TLS key pair. Should be used only when enabling TLS.
	TLSExtraDomainKey = "TLS_EXTRA_DOMAIN"
	// NoTLSKey is the key to disable TLS encryption.
	NoTLSKey = "NO_TLS"
	// NoProfilerKey is the key to disable Prometheus profiling.
	NoProfilerKey = "NO_PROFILER"
	// StatsIntervalKey is the key to customize the interval for the profiled to
	// gather profiling stats.
	StatsIntervalKey = "STATS_INTERVAL"
	// NodePeersKey is the key to customize the list of peers the embedded SPV
	// node will connect to when started
	NodePeersKey = "NODE_PEERS"
	// ElementsNodeRpcAddrKey is the key to set the rpc address of the node to connect
	// to when using elements-node-based blockchain scanner.
	ElementsNodeRpcAddrKey = "NODE_RPC_ADDR"
	// UtxoExpiryDurationKey is the key to customize the waiting time for one or
	// more previously locked utxos to be unlocked if not yet spent.
	UtxoExpiryDurationKey = "UTXO_EXPIRY_DURATION_IN_SECONDS"
	// RootPathKey is the key to use a custom root path for the wallet,
	// instead of the default m/84'/[1776|1]' (depending on network).
	RootPathKey = "ROOT_PATH"
	// EsploraUrlKey is the key for the esplora block esplorer consumed by the
	// neutrino blockchain scanner
	EsploraUrlKey = "ESPLORA_URL"

	// DbLocation is the folder inside the datadir containing db files.
	DbLocation = "db"
	// TLSLocation is the folder inside the datadir containing TLS key and
	// certificate.
	TLSLocation = "tls"
	// ScannerLocation is the folder inside the datadir containing blockchain
	// scanner files.
	ScannerLocation = "blockchain"
	// ProfilerLocation is the folder inside the datadir containing profiler
	// stats files.
	ProfilerLocation = "stats"
	// DbUserKey is user used to connect to db
	DbUserKey = "DB_USER"
	// DbPassKey is password used to connect to db
	DbPassKey = "DB_PASS"
	// DbHostKey is host where db is installed
	DbHostKey = "DB_HOST"
	// DbPortKey is port on which db is listening
	DbPortKey = "DB_PORT"
	// DbNameKey is name of database
	DbNameKey = "DB_NAME"
	// DbMigrationPath is the path to migration files
	DbMigrationPath = "DB_MIGRATION_PATH"
)

var (
	vip *viper.Viper

	defaultDatadir            = btcutil.AppDataDir("oceand", false)
	defaultDbType             = "badger"
	defaultBcScannerType      = "elements"
	defaultPort               = 18000
	defaultLogLevel           = 4
	defaultNetwork            = network.Liquid.Name
	defaultProfilerPort       = 18001
	defaultStatsInterval      = 600 // 10 minutes
	defaultUtxoExpiryDuration = 360 // 6 minutes (3 blocks)
	defaultEsploraUrl         = "https://blockstream.info/liquid/api"

	supportedNetworks = map[string]*network.Network{
		network.Liquid.Name:  &network.Liquid,
		network.Testnet.Name: &network.Testnet,
		network.Regtest.Name: &network.Regtest,
	}
	coinTypeByNetwork = map[string]int{
		network.Liquid.Name:  1776,
		network.Testnet.Name: 1,
		network.Regtest.Name: 1,
	}
	SupportedDbs = supportedType{
		"badger":   {},
		"inmemory": {},
		"postgres": {},
	}
	SupportedBcScanners = supportedType{
		"neutrino": {},
		"elements": {},
	}
)

func init() {
	vip = viper.New()
	vip.SetEnvPrefix("OCEAN")
	vip.AutomaticEnv()

	vip.SetDefault(DatadirKey, defaultDatadir)
	vip.SetDefault(DatabaseTypeKey, defaultDbType)
	vip.SetDefault(BlockchainScannerTypeKey, defaultBcScannerType)
	vip.SetDefault(PortKey, defaultPort)
	vip.SetDefault(NetworkKey, defaultNetwork)
	vip.SetDefault(LogLevelKey, defaultLogLevel)
	vip.SetDefault(NoTLSKey, false)
	vip.SetDefault(NoProfilerKey, false)
	vip.SetDefault(ProfilerPortKey, defaultProfilerPort)
	vip.SetDefault(StatsIntervalKey, defaultStatsInterval)
	vip.SetDefault(UtxoExpiryDurationKey, defaultUtxoExpiryDuration)
	vip.SetDefault(EsploraUrlKey, defaultEsploraUrl)
	vip.SetDefault(DbUserKey, "root")
	vip.SetDefault(DbPassKey, "secret")
	vip.SetDefault(DbHostKey, "127.0.0.1")
	vip.SetDefault(DbPortKey, 5432)
	vip.SetDefault(DbNameKey, "alertsd-db-pg")
	vip.SetDefault(DbMigrationPath, "file://internal/infrastructure/storage/db/postgres/migration")

	if err := validate(); err != nil {
		log.Fatalf("invalid config: %s", err)
	}

	if err := initDatadir(); err != nil {
		log.Fatalf("config: error while creating datadir: %s", err)
	}
}

func validate() error {
	datadir := GetString(DatadirKey)
	if len(datadir) <= 0 {
		return fmt.Errorf("datadir must not be null")
	}

	net := GetString(NetworkKey)
	if len(net) == 0 {
		return fmt.Errorf("network must not be null")
	}
	if _, ok := supportedNetworks[net]; !ok {
		nets := make([]string, 0, len(supportedNetworks))
		for net := range supportedNetworks {
			nets = append(nets, net)
		}
		return fmt.Errorf("unknown network, must be one of: %v", nets)
	}

	if net == network.Regtest.Name {
		if asset := GetString(NativeAssetKey); len(asset) > 0 {
			buf, err := hex.DecodeString(asset)
			if err != nil {
				return fmt.Errorf("invalid native asset string format, must be hex")
			}
			if len(buf) != 32 {
				return fmt.Errorf(
					"invalid native asset length, must be exactly 32 bytes in hex " +
						"string format",
				)
			}
		}
	}

	dbType := GetString(DatabaseTypeKey)
	if _, ok := SupportedDbs[dbType]; !ok {
		return fmt.Errorf("unsupported database type, must be one of %s", SupportedDbs)
	}

	bcScannerType := GetString(BlockchainScannerTypeKey)
	if _, ok := SupportedBcScanners[bcScannerType]; !ok {
		return fmt.Errorf(
			"unsupported blockchain scanner type, must be one of %s", SupportedBcScanners,
		)
	}

	if bcScannerType == "neutrino" {
		nodePeers := GetStringSlice(NodePeersKey)
		if len(nodePeers) == 0 {
			return fmt.Errorf("node peers list must not be empty")
		}
	}

	port := GetInt(PortKey)
	noProfiler := GetBool(NoProfilerKey)
	if !noProfiler {
		profilerPort := GetInt(ProfilerPortKey)
		if port == profilerPort {
			return fmt.Errorf("port and profiler port must not be equal")
		}
	}

	return nil
}

func GetDatadir() string {
	return filepath.Join(GetString(DatadirKey), GetString(NetworkKey))
}

func GetNetwork() *network.Network {
	net := supportedNetworks[GetString(NetworkKey)]

	if net.Name == network.Regtest.Name {
		if nativeAsset := GetString(NativeAssetKey); nativeAsset != "" {
			net.AssetID = nativeAsset
		}
	}

	return net
}

func GetRootPath() string {
	rootPath := GetString(RootPathKey)
	if rootPath != "" {
		return rootPath
	}

	coinType := coinTypeByNetwork[GetString(NetworkKey)]
	return fmt.Sprintf("m/84'/%d'", coinType)
}

func GetString(key string) string {
	return vip.GetString(key)
}

func GetInt(key string) int {
	return vip.GetInt(key)
}

func GetBool(key string) bool {
	return vip.GetBool(key)
}

func GetStringSlice(key string) []string {
	return vip.GetStringSlice(key)
}

func Set(key string, val interface{}) {
	vip.Set(key, val)
}

func Unset(key string) {
	vip.Set(key, nil)
}

func IsSet(key string) bool {
	return vip.IsSet(key)
}

func initDatadir() error {
	datadir := GetDatadir()
	if err := makeDirectoryIfNotExists(filepath.Join(datadir, DbLocation)); err != nil {
		return err
	}

	noProfiler := GetBool(NoProfilerKey)
	if !noProfiler {
		if err := makeDirectoryIfNotExists(filepath.Join(datadir, ProfilerLocation)); err != nil {
			return err
		}
	}

	noTls := GetBool(NoTLSKey)
	if noTls {
		return nil
	}
	if err := makeDirectoryIfNotExists(filepath.Join(datadir, TLSLocation)); err != nil {
		return err
	}
	return nil
}

func makeDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModeDir|0755)
	}
	return nil
}

type supportedType map[string]struct{}

func (t supportedType) String() string {
	types := make([]string, 0, len(t))
	for tt := range t {
		types = append(types, tt)
	}
	return strings.Join(types, " | ")
}
