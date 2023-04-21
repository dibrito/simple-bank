package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/dibrito/simple-bank/api"
	db "github.com/dibrito/simple-bank/db/sqlc"
	_ "github.com/dibrito/simple-bank/docs/statik"
	"github.com/dibrito/simple-bank/gapi"
	"github.com/dibrito/simple-bank/pb"
	"github.com/dibrito/simple-bank/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
)

// const (
// 	dbDriver = "postgres"
// 	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
// 	address  = "0.0.0.0:8080"
// )

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatalf("cannot load config:%v", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalf("unable to open db connection:%v", err)
	}

	store := db.NewStore(conn)
	go runGatawayServer(config, store)
	runGRPCServer(config, store)
}

func runGRPCServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatalf("cannot create server:%v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)
	// optinonal but  allows the gRPC client to easily explore
	// what RPCs are available on the server, and how to call them.
	reflection.Register(grpcServer)

	// start the server to listen to grpc requests in a port
	listener, err := net.Listen("tcp", config.GRPCAddress)
	if err != nil {
		log.Fatalf("cannot create listener:%v", err)
	}
	log.Printf("start grpc server at:%v", listener.Addr().String())

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("cannot start grpc server:%v", err)
	}
}

func runGatawayServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatalf("cannot create server:%v", err)
	}

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})
	grpcMux := runtime.NewServeMux(jsonOption)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("cannot register handler server")
	}
	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	// fs := http.FileServer(http.Dir("./docs/swagger"))
	fsStatik, err := fs.New()
	if err != nil {
		log.Fatalf("cannot create statik file system:%v", err)
	}
	// slash at the end is necessary otherwise css and other files will break
	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(fsStatik))
	mux.Handle("/swagger/", swaggerHandler)
	// start the server to listen to grpc requests in a port
	listener, err := net.Listen("tcp", config.HttpAddress)
	if err != nil {
		log.Fatalf("cannot create listener:%v", err)
	}
	log.Printf("start HTTP gateway server at:%v", listener.Addr().String())

	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatalf("cannot start HTTP gateway server:%v", err)
	}
}

func runHttpServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatalf("cannot create server:%v", err)
	}

	err = server.Start(config.HttpAddress)
	if err != nil {
		log.Fatalf("cannot start server:%v", err)
	}
	log.Println("=============http server up=============")
}
