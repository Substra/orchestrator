// Package wallet holds the logic around chaincode identity management.
package wallet

import (
	"io/ioutil"
	"sync"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

// Wallet is a wrapper around gateway.Wallet.
// It provides convenience methods to ensure a specific identity exists.
type Wallet struct {
	*gateway.Wallet
	certPath string
	keyPath  string
	m        sync.RWMutex
}

// New returns a ready to use instance of in-memory wallet
func New(certPath string, keyPath string) *Wallet {
	return &Wallet{gateway.NewInMemoryWallet(), certPath, keyPath, sync.RWMutex{}}
}

// EnsureIdentity make sure the given identity is present in the wallet.
func (w *Wallet) EnsureIdentity(label string, mspid string) error {
	w.m.RLock()
	knownIdentity := w.Exists(label)
	w.m.RUnlock()

	if !knownIdentity {
		w.m.Lock()
		defer w.m.Unlock()
		cert, err := ioutil.ReadFile(w.certPath)
		if err != nil {
			return err
		}

		key, err := ioutil.ReadFile(w.keyPath)
		if err != nil {
			return err
		}

		identity := gateway.NewX509Identity(mspid, string(cert), string(key))

		err = w.Put(label, identity)
		if err != nil {
			return err
		}
		log.WithField("label", label).WithField("mspid", mspid).Info("Identity added to wallet")
	}

	return nil
}
