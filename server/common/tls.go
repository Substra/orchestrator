package common

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
	"path"

	"github.com/go-playground/log/v7"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GetTLSOptions will return server option with optional TLS and mTLS setup.
// This may panic on missing or invalid configuration env var.
func GetTLSOptions() grpc.ServerOption {
	if !MustGetEnvFlag("TLS_ENABLED") {
		return nil
	}

	tlsCertFile := MustGetEnv("TLS_CERT_PATH")
	tlsKeyFile := MustGetEnv("TLS_KEY_PATH")

	log.WithField("cert", tlsCertFile).WithField("key", tlsKeyFile).Info("Loading TLS certificates")

	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(tlsCertFile, tlsKeyFile)
	if err != nil {
		log.WithError(err).Fatal("Failed to load server TLS certificate")
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		MinVersion:   tls.VersionTLS12,
	}

	if MustGetEnvFlag("MTLS_ENABLED") {
		certPool := x509.NewCertPool()

		// 1. Add our own CA to trusted client CAs
		// This allows:
		// - Clients to connect using certificates issued by our CA
		// - Connecting to ourselves using our own cert/key (e.g. health probe)
		serverCA := MustGetEnv("TLS_SERVER_CA_CERT")
		pemServerCA, err := ioutil.ReadFile(serverCA)
		if err != nil {
			log.WithField("CA Cert", serverCA).Fatal("Failed to load TLS server CA")
		}
		if !certPool.AppendCertsFromPEM(pemServerCA) {
			log.WithField("CA Cert", serverCA).Fatal("Failed to add TLS client CA certificate to pool")
		}

		// 2. Add provided client CAs
		clientCAPath := MustGetEnv("TLS_CLIENT_CA_CERT_DIR")
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
