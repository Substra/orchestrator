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

	"github.com/go-playground/log/v7"
	"github.com/go-playground/log/v7/handlers/console"
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
	chaincode    = flag.String("chaincode", "mycc", "Chaincode to use (only relevant in distributed mode)")

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
	initTestClientFactory(conn, *mspid, *channel, *chaincode)
}

func tearDown() {
	if conn == nil {
		return
	}

	err := conn.Close()
	if err != nil {
		log.Fatalf("fail to close gRPC connection: %v", err)
	}
}

func setUpLogging() {
	cLog := console.New(true)
	levels := make([]log.Level, 0)
	for _, lvl := range log.AllLevels {
		if !*debugEnabled && lvl == log.DebugLevel {
			continue
		}

		levels = append(levels, lvl)
	}

	log.AddHandler(cLog, levels...)

}

func initGrpcConn() {
	var opts []grpc.DialOption
	if *tlsEnabled {
		b, err := ioutil.ReadFile(*caFile)
		if err != nil {
			log.WithError(err).Fatal("failed to read cacert")
		}
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(b) {
			log.Fatal("failed to append cert")
		}
		cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
		if err != nil {
			log.Fatal("failed to load client keypair")
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

	log.WithField("server_addr", *serverAddr).Info("connecting to server")

	dialCtx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	var err error
	conn, err = grpc.DialContext(dialCtx, *serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
}
