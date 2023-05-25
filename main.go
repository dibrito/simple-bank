package main

import (
	"context"
	"database/sql"
	"os"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	// "log"
	"net"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/dibrito/simple-bank/api"
	db "github.com/dibrito/simple-bank/db/sqlc"
	_ "github.com/dibrito/simple-bank/docs/statik"
	"github.com/dibrito/simple-bank/gapi"
	"github.com/dibrito/simple-bank/mail"
	"github.com/dibrito/simple-bank/pb"
	"github.com/dibrito/simple-bank/util"
	"github.com/dibrito/simple-bank/worker"

	// we need: v4/database/file and v4/source/file imports to run migration within the code
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
		log.Fatal().Err(err).Msg("cannot load config:%v")
	}

	if config.Env == "dev" {
		// To log a human-friendly, colorized output
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to open db connection:%v")
	}

	// run db migration
	runDBMigration(config.DBMigrationPath, config.DBSource)

	store := db.NewStore(conn)
	// run redis
	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributer := worker.NewRedisDistributor(redisOpt)
	go runTaskProcessor(config, redisOpt, store)
	go runGatawayServer(config, store, taskDistributer)
	runGRPCServer(config, store, taskDistributer)
}

func runDBMigration(migrationUrl, dbSource string) {
	migration, err := migrate.New(migrationUrl, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create a new migrate instance:%v")
	}

	err = migration.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("failed to run migration up:%v")
	}
	log.Info().Msg("db migrated successfully!")
}

func runGRPCServer(config util.Config, store db.Store, td worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, td)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server:%v")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)

	pb.RegisterSimpleBankServer(grpcServer, server)
	// optinonal but  allows the gRPC client to easily explore
	// what RPCs are available on the server, and how to call them.
	reflection.Register(grpcServer)

	// start the server to listen to grpc requests in a port
	listener, err := net.Listen("tcp", config.GRPCAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener:%v")
	}
	log.Info().Msgf("start grpc server at:%v", listener.Addr().String())

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start grpc server:%v")
	}
}

func runGatawayServer(config util.Config, store db.Store, td worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, td)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server:%v")
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

	// we’re calling the RegisterSimpleBankHandlerServer() function,
	// Which performs in-process translation between HTTP and gRPC.
	// Or in other words,
	// it will call the handler function of the gRPC server directly,
	// without going through any gRPC interceptor.
	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot register handler server")
	}
	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	// fs := http.FileServer(http.Dir("./docs/swagger"))
	fsStatik, err := fs.New()
	if err != nil {
		log.Fatal().Err(err).Err(err).Msg("cannot create statik file system:%v")
	}
	// slash at the end is necessary otherwise css and other files will break
	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(fsStatik))
	mux.Handle("/swagger/", swaggerHandler)
	// start the server to listen to grpc requests in a port
	listener, err := net.Listen("tcp", config.HttpAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener:%v")
	}
	log.Info().Msgf("start HTTP gateway server at:%v", listener.Addr().String())

	logger := gapi.HttpLogger(mux)
	err = http.Serve(listener, logger)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start HTTP gateway server:%v")
	}
}

func runHttpServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server:%v")
	}

	err = server.Start(config.HttpAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server:%v")
	}
	log.Info().Msg("=============http server up=============")
}

// We will have to call runTaskProcessor&nbsp; in a separate go routine
// because when the processor starts,
// the Asynq server will block and keep polling Redis for new tasks.
// its design is pretty similar to that of an HTTP webserver.
// So it blocks, just like the HTTP server block while waiting for requests from the client.
func runTaskProcessor(c util.Config, redisOpt asynq.RedisClientOpt, store db.Store) {
	mailer := mail.NewGmailSender(c.EmailSenderName, c.EmailSenderAddress, c.EmailSenderPassword)
	tp := worker.NewRedisTaskProcessor(redisOpt, store, mailer)
	log.Info().Msg("start task processor")
	err := tp.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("start task processor")
	}
}
