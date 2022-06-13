package common

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/lib/errors"
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
				log.WithField("cacert", f).Fatal("Failed to load TLS client CA certificate")
			}
			if !certPool.AppendCertsFromPEM(pemClientCA) {
				log.WithField("cacert", f).Fatal("Failed to add TLS client CA certificate to pool")
			}
			log.WithField("cacert", f).Info("Loaded TLS client CA certificate")
		}

		config.ClientAuth = tls.RequireAndVerifyClientCert
		config.ClientCAs = certPool
	}

	tlsCredentials := credentials.NewTLS(config)
	return grpc.Creds(tlsCredentials)
}

type clientCACertCallback = func(org, filepath string) error

// walkClientCACerts will walk the TLS client directory, stopping at the first error.
// Expected structure is to have a directory per org, and one or more certificates in it.
// The callback will receive the parent organization name along with the file path.
func walkClientCACerts(root string, callback clientCACertCallback) error {
	fileInfo, err := ioutil.ReadDir(root)
	if err != nil {
		return err
	}

	// for each org directory
	for _, orgDir := range fileInfo {
		if !orgDir.IsDir() {
			continue
		}
		orgFullPath := path.Join(root, orgDir.Name())
		fileInfo, err = ioutil.ReadDir(orgFullPath)
		if err != nil {
			return err
		}

		// for each inode inside the org directory
		for _, file := range fileInfo {
			filePath := path.Join(orgFullPath, file.Name())

			// resolve symlinks
			for file.Mode()&os.ModeSymlink != 0 {
				resolved, err := os.Readlink(filePath)
				if err != nil {
					return err
				}
				filePath = path.Join(path.Dir(filePath), resolved)
				file, err = os.Stat(filePath)
				if err != nil {
					return err
				}
				filePath = path.Join(orgFullPath, file.Name())
			}

			// we found a CA cert file
			if !file.IsDir() {
				err = callback(orgDir.Name(), filePath)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// findCACerts walks a client CA cert root directory and returns
// a list of all the client CA certificate it finds
func findCACerts(root string) ([]string, error) {
	var files []string

	callback := func(_, filePath string) error {
		files = append(files, filePath)
		return nil
	}

	err := walkClientCACerts(root, callback)
	if err != nil {
		return nil, err
	}

	return files, nil
}

type OrgCACertList = map[string][]string

// GetOrgCACerts returns the valid CA keys per organization (mspid).
func GetOrgCACerts() (OrgCACertList, error) {
	orgCACerts := make(OrgCACertList)

	clientCAPath, ok := GetEnv("TLS_CLIENT_CA_CERT_DIR")
	if !ok {
		return nil, errors.NewInternal("ORCHESTRATOR_TLS_CLIENT_CA_CERT_DIR env var is not set")
	}

	callback := func(org, filePath string) error {
		key, err := getCAKeyID(filePath)
		if err != nil {
			return err
		}
		orgCACerts[org] = append(orgCACerts[org], key)

		return nil
	}

	err := walkClientCACerts(clientCAPath, callback)
	if err != nil {
		return nil, err
	}

	return orgCACerts, nil
}

// getCAKeyID returns the identifier of the CA certificate
func getCAKeyID(file string) (string, error) {
	rawPem, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	certPemBlock, _ := pem.Decode(rawPem)
	if err != nil {
		return "", err
	}

	cert, err := x509.ParseCertificate(certPemBlock.Bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(cert.SubjectKeyId), nil
}
