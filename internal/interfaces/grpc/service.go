package grpc_interface

import (
	"fmt"
	"math/big"

	log "github.com/sirupsen/logrus"
	pb "github.com/vulpemventures/ocean/api-spec/protobuf/gen/go/ocean/v1alpha"
	appconfig "github.com/vulpemventures/ocean/internal/app-config"
	"github.com/vulpemventures/ocean/internal/core/domain"
	grpc_handler "github.com/vulpemventures/ocean/internal/interfaces/grpc/handler"
	grpc_interceptor "github.com/vulpemventures/ocean/internal/interfaces/grpc/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	tlsKeyFile        = "key.pem"
	tlsCertFile       = "cert.pem"
	serialNumberLimit = new(big.Int).Lsh(big.NewInt(1), 128)
)

type service struct {
	config                   ServiceConfig
	appConfig                *appconfig.AppConfig
	grpcServer               *grpc.Server
	chCloseStreamConnections chan (struct{})

	log func(format string, a ...interface{})
}

func NewService(config ServiceConfig, appConfig *appconfig.AppConfig) (*service, error) {
	logFn := func(format string, a ...interface{}) {
		format = fmt.Sprintf("service: %s", format)
		log.Infof(format, a...)
	}
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %s", err)
	}
	if err := appConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid app config: %s", err)
	}

	if !config.insecure() {
		if err := generateTLSKeyPair(
			config.TLSLocation, config.ExtraIPs, config.ExtraDomains,
		); err != nil {
			return nil, fmt.Errorf("error while creating TLS keypair: %s", err)
		}
		logFn("created TLS keypair in path %s", config.TLSLocation)
	}
	chCloseStreamConnections := make(chan struct{})
	return &service{config, appConfig, nil, chCloseStreamConnections, logFn}, nil
}

func (s *service) Start() error {
	if err := s.registerHandlerForWalletEvents(); err != nil {
		return err
	}

	withOnlyWalletService := true
	srv, err := s.start(withOnlyWalletService)
	if err != nil {
		return err
	}

	s.log("start listening on %s", s.config.address())

	s.grpcServer = srv
	return nil
}

func (s *service) Stop() {
	onlyGrpcServer := true
	allServices := !onlyGrpcServer
	s.stop(allServices)
	s.log("shutdown")
}

func (s *service) registerHandlerForWalletEvents() error {
	s.appConfig.WalletService().RegisterHandlerForWalletEvent(
		domain.WalletUnlocked, func(_ domain.WalletEvent) {
			onlyGrpcServer := true
			s.stop(onlyGrpcServer)
			withOnlyWalletService := true
			withAllServices := !withOnlyWalletService
			srv, _ := s.start(withAllServices)
			s.grpcServer = srv
		},
	)
	return nil
}

func (s *service) start(withOnlyWalletService bool) (*grpc.Server, error) {
	s.appConfig.BlockchainScanner().Start()

	grpcConfig := []grpc.ServerOption{
		grpc_interceptor.UnaryInterceptor(), grpc_interceptor.StreamInterceptor(),
	}
	if !s.config.insecure() {
		creds, err := credentials.NewServerTLSFromFile(
			s.config.tlsCertPath(), s.config.tlsKeyPath(),
		)
		if err != nil {
			return nil, err
		}
		grpcConfig = append(grpcConfig, grpc.Creds(creds))
	}

	grpcServer := grpc.NewServer(grpcConfig...)

	walletHandler := grpc_handler.NewWalletHandler(s.appConfig.WalletService())
	pb.RegisterWalletServiceServer(grpcServer, walletHandler)

	s.log("registered wallet handler on public interface")

	if !withOnlyWalletService {
		accountHandler := grpc_handler.NewAccountHandler(s.appConfig.AccountService())
		txHandler := grpc_handler.NewTransactionHandler(s.appConfig.TransactionService())
		notifyHandler := grpc_handler.NewNotificationHandler(
			s.appConfig.NotificationService(), s.chCloseStreamConnections,
		)

		pb.RegisterAccountServiceServer(grpcServer, accountHandler)
		pb.RegisterTransactionServiceServer(grpcServer, txHandler)
		pb.RegisterNotificationServiceServer(grpcServer, notifyHandler)
		s.log("registered account handler on public interface")
		s.log("registered transaction handler on public interface")
		s.log("registered notification handler on public interface")
	}

	go grpcServer.Serve(s.config.listener())

	return grpcServer, nil
}

func (s *service) stop(onlyGrpcServer bool) {
	select {
	case s.chCloseStreamConnections <- struct{}{}:
		s.log("closed stream connections")
	default:
	}

	s.grpcServer.GracefulStop()
	s.log("stopped grpc server")
	if onlyGrpcServer {
		return
	}

	s.appConfig.BlockchainScanner().Stop()
	s.log("stopped blockchain scanner")
	s.appConfig.RepoManager().Close()
	s.log("closed connection with db")
}
