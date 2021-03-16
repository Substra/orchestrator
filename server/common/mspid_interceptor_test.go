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
	"fmt"
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

	params := []struct {
		peer          *peer.Peer
		shouldSucceed bool
		msg           string
	}{
		{
			nil,
			false,
			"Certificate with a missing peer",
		},
		{
			&peer.Peer{},
			false,
			"Certificate with an invalid peer",
		},
		{
			&peer.Peer{AuthInfo: FakeTLSAuthInfo{}},
			false,
			"Certificate with an incorrect auth info type",
		},
		{
			&peer.Peer{AuthInfo: credentials.TLSInfo{}},
			false,
			"Certificate with an empty TLS info",
		},
		{
			&peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{}}},
			false,
			"Certificate with an empty connection state",
		},
		{
			&peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{VerifiedChains: [][]*x509.Certificate{}}}},
			false,
			"Certificate with an empty certificate",
		},
		{
			&peer.Peer{
				AuthInfo: credentials.TLSInfo{
					State: tls.ConnectionState{
						VerifiedChains: [][]*x509.Certificate{
							{},
						},
					},
				},
			},
			false,
			"Certificate with an empty certificate",
		},
		{
			&peer.Peer{
				AuthInfo: credentials.TLSInfo{
					State: tls.ConnectionState{
						VerifiedChains: [][]*x509.Certificate{
							{&x509.Certificate{}},
						},
					},
				},
			},
			false,
			"Certificate with an empty certificate",
		},
		{
			&peer.Peer{
				AuthInfo: credentials.TLSInfo{
					State: tls.ConnectionState{
						VerifiedChains: [][]*x509.Certificate{
							{&x509.Certificate{Subject: pkix.Name{}}},
						},
					},
				},
			},
			false,
			"Certificate with an empty subject",
		},
		{
			&peer.Peer{
				AuthInfo: credentials.TLSInfo{
					State: tls.ConnectionState{
						VerifiedChains: [][]*x509.Certificate{
							{&x509.Certificate{Subject: pkix.Name{Organization: []string{}}}},
						},
					},
				},
			},
			false,
			"Certificate with an empty list of organizations",
		},
		{
			&peer.Peer{
				AuthInfo: credentials.TLSInfo{
					State: tls.ConnectionState{
						VerifiedChains: [][]*x509.Certificate{
							{&x509.Certificate{Subject: pkix.Name{Organization: []string{MSPID}}}},
						},
					},
				},
			},
			true,
			"Certificate with a valid MSPID",
		},
		{
			&peer.Peer{
				AuthInfo: credentials.TLSInfo{
					State: tls.ConnectionState{
						VerifiedChains: [][]*x509.Certificate{
							{&x509.Certificate{Subject: pkix.Name{Organization: []string{"some other mspid", MSPID}}}},
						},
					},
				},
			},
			true,
			"Certificate with both a valid MSPID and and invalid one",
		},
	}

	for i := range params {
		actual := testVerifyClientMSPID(t, MSPID, params[i].peer, params[i].shouldSucceed, params[i].msg)
		shouldSucceed := params[i].shouldSucceed
		prefix := "Should return a validation error"
		if shouldSucceed {
			prefix = "Should not return a validation error"
		}
		assert.Equal(t, shouldSucceed, actual == nil, fmt.Sprintf("%v: %v", prefix, params[i].msg))
	}
}

func testVerifyClientMSPID(t *testing.T, MSPID string, p *peer.Peer, shouldSucceed bool, msg string) error {
	ctx := context.TODO()
	ctx = peer.NewContext(ctx, p)
	return VerifyClientMSPID(ctx, MSPID)
}
