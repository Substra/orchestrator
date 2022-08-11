package chaincode

import (
	"io/ioutil"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/rs/zerolog/log"
)

// Wallet holds the logic around chaincode identity management.
// It is a wrapper around gateway.Wallet and provides convenience methods
// to ensure a specific identity exists.
type Wallet struct {
	*gateway.Wallet
	certPath string
	keyPath  string
	m        sync.RWMutex
}

// NewWallet returns a ready to use instance of in-memory wallet
func NewWallet(certPath string, keyPath string) *Wallet {
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
		log.Info().Str("label", label).Str("mspid", mspid).Msg("Identity added to wallet")
	}

	return nil
}
