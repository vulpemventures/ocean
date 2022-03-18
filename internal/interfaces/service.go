package interfaces

import (
	"fmt"

	appconfig "github.com/vulpemventures/ocean/internal/app-config"
	"github.com/vulpemventures/ocean/internal/core/domain"
	cypher "github.com/vulpemventures/ocean/internal/infrastructure/mnemonic-cypher/aes128"
	store "github.com/vulpemventures/ocean/internal/infrastructure/mnemonic-store/in-memory"
	grpc_interface "github.com/vulpemventures/ocean/internal/interfaces/grpc"
)

// Service interface defines the methods that every kind of interface, whether
// gRPC, REST, or whatever must be compliant with.
type Service interface {
	Start() error
	Stop()
}

type ServiceManager struct {
	Service
}

func NewGrpcServiceManager(
	config grpc_interface.ServiceConfig, appConfig *appconfig.AppConfig,
) (*ServiceManager, error) {
	svc, err := grpc_interface.NewService(config, appConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initalize grpc service: %s", err)
	}

	domain.MnemonicCypher = cypher.NewAES128Cypher()
	domain.MnemonicStore = store.NewInMemoryMnemonicStore()
	return &ServiceManager{svc}, nil
}
