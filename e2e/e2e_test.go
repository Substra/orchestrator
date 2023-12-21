//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	debugEnabled = flag.Bool("debug", false, "Debug mode (very verbose)")
	tlsEnabled   = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile       = flag.String("cafile", "", "The file containing the CA root cert file")
	certFile     = flag.String("certfile", "", "The file containing the client cert file")
	keyFile      = flag.String("keyfile", "", "The file containing the client cert key")
	serverAddr   = flag.String("server_addr", "localhost:9000", "The server address in the format of host:port")
	mspid        = flag.String("mspid", "MyOrg1MSP", "MSP ID")
	channel      = flag.String("channel", "mychannel", "Channel to use")

	conn *grpc.ClientConn
)

func TestMain(m *testing.M) {
	flag.Parse()

	if flag.Lookup("test.list").Value.String() == "" {
		// Set up test environment when not listing test cases
		setUp()
	}
	exitCode := m.Run()
	tearDown()
	os.Exit(exitCode)
}

func setUp() {
	setUpLogging()
	initGrpcConn()
	initTestClientFactory(conn, *mspid, *channel)
}

func tearDown() {
	if conn == nil {
		return
	}

	err := conn.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("fail to close gRPC connection")
	}
}

func setUpLogging() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if *debugEnabled {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func initGrpcConn() {
	var opts []grpc.DialOption
	if *tlsEnabled {
		b, err := ioutil.ReadFile(*caFile)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to read cacert")
		}
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(b) {
			log.Fatal().Msg("failed to append cert")
		}
		cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
		if err != nil {
			log.Fatal().Msg("failed to load client keypair")
		}
		config := &tls.Config{
			InsecureSkipVerify: false,
			Certificates:       []tls.Certificate{cert},
			RootCAs:            cp,
			MinVersion:         tls.VersionTLS12,
		}
		creds := grpc.WithTransportCredentials(credentials.NewTLS(config))

		opts = append(opts, creds)
	} else {
		creds := grpc.WithTransportCredentials(insecure.NewCredentials())
		opts = append(opts, creds)
	}

	opts = append(opts, grpc.WithBlock())

	log.Info().Str("server_addr", *serverAddr).Msg("connecting to server")

	dialCtx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	var err error
	conn, err = grpc.DialContext(dialCtx, *serverAddr, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to dial")
	}
}
