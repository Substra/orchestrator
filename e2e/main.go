// Package main implements end to end testing client.
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"fmt"

	"github.com/go-playground/log/v7"
	"github.com/go-playground/log/v7/handlers/console"
	"github.com/owkin/orchestrator/e2e/client"
	"github.com/owkin/orchestrator/e2e/scenarios"
	"github.com/owkin/orchestrator/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type tagList struct {
	list []string
}

func (l *tagList) String() string {
	return strings.Join(l.list, "-")
}

func (l *tagList) Set(value string) error {
	l.list = append(l.list, value)
	return nil
}

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
	list         = flag.Bool("list", false, "List available tests and their tags")
	nameFilter   = flag.String("name", "", "Filter test by name")
	tagFilter    = tagList{list: []string{}}
)

func main() {
	flag.Var(&tagFilter, "tag", "Filter test by tags")
	flag.Parse()

	if *list {
		listTests()
		return
	}

	cLog := console.New(true)
	levels := make([]log.Level, 0)
	for _, lvl := range log.AllLevels {
		if !*debugEnabled && lvl == log.DebugLevel {
			continue
		}

		levels = append(levels, lvl)
	}

	log.AddHandler(cLog, levels...)

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
	conn, err := grpc.DialContext(dialCtx, *serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	log.Debug("Starting testing")

	if len(tagFilter.list) == 0 && *nameFilter == "" {
		tagFilter.list = append(tagFilter.list, "short")
	}

	testClientFactory := client.NewTestClientFactory(conn, *mspid, *channel, *chaincode)

scenario:
	for name, sc := range scenarios.GatherTestScenarios() {
		if *nameFilter != "" && *nameFilter != name {
			// skip non matching test
			continue
		}
		for _, tag := range tagFilter.list {
			if utils.StringInSlice(sc.Tags, tag) {
				break
			}
			// No match
			continue scenario
		}

		logger := log.WithField("name", name)
		logger.Debug("starting scenario")
		func() {
			defer logger.WithTrace().Info("test succeeded")
			sc.Exec(testClientFactory)
		}()

	}
}

// listTests will output the list of available scenario and their associated tags.
func listTests() {
	w := tabwriter.NewWriter(os.Stdout, 0, 1, 8, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "name\ttags")
	fmt.Fprintln(w, "----\t----")

	for name, sc := range scenarios.GatherTestScenarios() {
		fmt.Fprintf(w, "%s\t%s\n", name, sc.Tags)
	}
	w.Flush()
}
