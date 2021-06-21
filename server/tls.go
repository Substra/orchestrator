// Copyright 2020 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
	"path"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/server/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func getTLSOptions() grpc.ServerOption {
	if !common.MustGetEnvFlag("TLS_ENABLED") {
		return nil
	}

	tlsCertFile := common.MustGetEnv("TLS_CERT_PATH")
	tlsKeyFile := common.MustGetEnv("TLS_KEY_PATH")

	log.WithField("cert", tlsCertFile).WithField("key", tlsKeyFile).Info("Loading TLS certificates")

	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(tlsCertFile, tlsKeyFile)
	if err != nil {
		log.WithError(err).Fatal("Failed to load server TLS certificate")
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}

	if common.MustGetEnvFlag("MTLS_ENABLED") {
		certPool := x509.NewCertPool()

		// 1. Add our own CA to trusted client CAs
		// This allows:
		// - Clients to connect using certificates issued by our CA
		// - Connecting to ourselves using our own cert/key (e.g. health probe)
		serverCA := common.MustGetEnv("TLS_SERVER_CA_CERT")
		pemServerCA, err := ioutil.ReadFile(serverCA)
		if err != nil {
			log.WithField("CA Cert", serverCA).Fatal("Failed to load TLS server CA")
		}
		if !certPool.AppendCertsFromPEM(pemServerCA) {
			log.WithField("CA Cert", serverCA).Fatal("Failed to add TLS client CA certificate to pool")
		}

		// 2. Add provided client CAs
		clientCAPath := common.MustGetEnv("TLS_CLIENT_CA_CERT_DIR")
		clientCAFiles, err := findCACerts(clientCAPath)
		if err != nil {
			log.WithError(err).Fatal("Failed to load TLS client CA certificates")
		}
		for _, f := range clientCAFiles {
			pemClientCA, err := ioutil.ReadFile(f)
			if err != nil {
				log.WithField("CA Cert", f).Fatal("Failed to load TLS client CA certificate")
			}
			if !certPool.AppendCertsFromPEM(pemClientCA) {
				log.WithField("CA Cert", f).Fatal("Failed to add TLS client CA certificate to pool")
			}
			log.WithField("cacert", f).Info("Loaded TLS client CA certificate")
		}

		config.ClientAuth = tls.RequireAndVerifyClientCert
		config.ClientCAs = certPool
	}

	tlsCredentials := credentials.NewTLS(config)
	return grpc.Creds(tlsCredentials)
}

// findCACerts walks a client CA cert root directory and returns
// a list of all the client CA certificate it finds
func findCACerts(root string) ([]string, error) {
	var files []string

	fileInfo, err := ioutil.ReadDir(root)
	if err != nil {
		return files, err
	}

	// for each org directory
	for _, orgDir := range fileInfo {
		if !orgDir.IsDir() {
			continue
		}
		orgFullPath := path.Join(root, orgDir.Name())
		fileInfo, err = ioutil.ReadDir(orgFullPath)
		if err != nil {
			return files, err
		}

		// for each inode inside the org directory
		for _, file := range fileInfo {
			filePath := path.Join(orgFullPath, file.Name())

			// resolve symlinks
			for file.Mode()&os.ModeSymlink != 0 {
				resolved, err := os.Readlink(filePath)
				if err != nil {
					return files, err
				}
				filePath = path.Join(path.Dir(filePath), resolved)
				file, err = os.Stat(filePath)
				if err != nil {
					return files, err
				}
				filePath = path.Join(orgFullPath, file.Name())
			}

			// we found a CA cert file
			if !file.IsDir() {
				files = append(files, filePath)
			}
		}
	}

	return files, nil
}
