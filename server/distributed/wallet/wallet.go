// Copyright 2021 Owkin Inc.
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

// Package wallet holds the logic around chaincode identity management.
package wallet

import (
	"io/ioutil"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

// Wallet is a wrapper around gateway.Wallet.
// It provides convenience methods to ensure a specific identity exists.
type Wallet struct {
	*gateway.Wallet
	certPath string
	keyPath  string
}

// New returns a ready to use instance of in-memory wallet
func New(certPath string, keyPath string) *Wallet {
	return &Wallet{gateway.NewInMemoryWallet(), certPath, keyPath}
}

// EnsureIdentity make sure the given identity is present in the wallet.
func (w *Wallet) EnsureIdentity(label string, mspid string) error {
	if !w.Exists(label) {
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
