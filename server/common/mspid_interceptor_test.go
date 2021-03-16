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

package common

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

type FakeTLSAuthInfo struct{}

func (t FakeTLSAuthInfo) AuthType() string {
	return "fake"
}
func TestExtractMSPID(t *testing.T) {
	ctx := context.TODO()

	ctxWithMSPID := context.WithValue(ctx, ctxMSPIDKey, "OrgMSP")

	extracted, err := ExtractMSPID(ctxWithMSPID)
	assert.NoError(t, err, "extraction should not fail")
	assert.Equal(t, "OrgMSP", extracted, "MSPID should be extracted from context")

	_, err = ExtractMSPID(ctx)
	assert.Error(t, err, "Extraction should fail on empty context")
}

func TestVerifyClientMSPID(t *testing.T) {

	MSPID := "my-msp-id"
	var p *peer.Peer

	//
	// Error cases
	//
	p = nil
	testVerifyClientMSPID(t, MSPID, p, false, "Certificate with a missing peer should not be validated with success")

	p = &peer.Peer{}
	testVerifyClientMSPID(t, MSPID, p, false, "Certificate with an invalid peer should not be validated with success")

	p = &peer.Peer{AuthInfo: FakeTLSAuthInfo{}}
	testVerifyClientMSPID(t, MSPID, p, false, "Certificate with an incorrect auth info type should not be validated with success")

	p = &peer.Peer{AuthInfo: credentials.TLSInfo{}}
	testVerifyClientMSPID(t, MSPID, p, false, "Certificate with an empty TLS info should not be validated with success")

	p = &peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{}}}
	testVerifyClientMSPID(t, MSPID, p, false, "Certificate with an empty connection state should not be validated with success")

	p = &peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{VerifiedChains: [][]*x509.Certificate{}}}}
	testVerifyClientMSPID(t, MSPID, p, false, "Certificate with an empty certificate should not be validated with success")

	p = &peer.Peer{
		AuthInfo: credentials.TLSInfo{
			State: tls.ConnectionState{
				VerifiedChains: [][]*x509.Certificate{
					{},
				},
			},
		},
	}
	testVerifyClientMSPID(t, MSPID, p, false, "Certificate with an empty certificate should not be validated with success")

	p = &peer.Peer{
		AuthInfo: credentials.TLSInfo{
			State: tls.ConnectionState{
				VerifiedChains: [][]*x509.Certificate{
					{&x509.Certificate{}},
				},
			},
		},
	}
	testVerifyClientMSPID(t, MSPID, p, false, "Certificate with an empty certificate should not be validated with success")

	p = &peer.Peer{
		AuthInfo: credentials.TLSInfo{
			State: tls.ConnectionState{
				VerifiedChains: [][]*x509.Certificate{
					{&x509.Certificate{Subject: pkix.Name{}}},
				},
			},
		},
	}
	testVerifyClientMSPID(t, MSPID, p, false, "Certificate with an empty subject should not be validated with success")

	p = &peer.Peer{
		AuthInfo: credentials.TLSInfo{
			State: tls.ConnectionState{
				VerifiedChains: [][]*x509.Certificate{
					{&x509.Certificate{Subject: pkix.Name{Organization: []string{}}}},
				},
			},
		},
	}
	testVerifyClientMSPID(t, MSPID, p, false, "Certificate with an empty list of organizations should not be validated with success")

	//
	// Success cases
	//
	p = &peer.Peer{
		AuthInfo: credentials.TLSInfo{
			State: tls.ConnectionState{
				VerifiedChains: [][]*x509.Certificate{
					{&x509.Certificate{Subject: pkix.Name{Organization: []string{MSPID}}}},
				},
			},
		},
	}
	testVerifyClientMSPID(t, MSPID, p, true, "Certificate with a valid MSPID should be verified without error")

	p = &peer.Peer{
		AuthInfo: credentials.TLSInfo{
			State: tls.ConnectionState{
				VerifiedChains: [][]*x509.Certificate{
					{&x509.Certificate{Subject: pkix.Name{Organization: []string{"some other mspid", MSPID}}}},
				},
			},
		},
	}
	testVerifyClientMSPID(t, MSPID, p, true, "Certificate with both a valid MSPID and and invalid one should be verified without error")
}

func testVerifyClientMSPID(t *testing.T, MSPID string, p *peer.Peer, shouldSucceed bool, msg string) {
	ctx := context.TODO()
	ctx = peer.NewContext(ctx, p)

	err := VerifyClientMSPID(ctx, MSPID)

	if shouldSucceed {
		assert.NoError(t, err, msg)
	} else {
		assert.Error(t, err, msg)
	}
}
