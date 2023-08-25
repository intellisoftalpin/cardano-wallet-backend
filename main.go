package main

import (
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	walletPB "github.com/intellisoftalpin/proto/proto-gen/wallet"

	"github.com/intellisoftalpin/cardano-wallet-backend/config"
	"github.com/intellisoftalpin/cardano-wallet-backend/wallet"
)

func main() {
	// grpc.EnableTracing = true

	loadedConfig, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", ":"+loadedConfig.ServerPort)
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}

	// ----------------------------------------------------------------------
	var opts []grpc.ServerOption

	// creds, err := cert.SetupTLS(loadedConfig.TLS)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// opts = []grpc.ServerOption{grpc.Creds(creds)}

	grpcServer := grpc.NewServer(opts...)

	// ----------------------------------------------------------------------

	walletPB.RegisterWalletServer(grpcServer, wallet.NewServer(loadedConfig))
	grpcServer.Serve(listener)
}
