package interceptors

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

type fakeTLSAuthInfo struct{}

func (t fakeTLSAuthInfo) AuthType() string {
	return "fake"
}
func TestExtractMSPID(t *testing.T) {
	ctx := context.TODO()

	ctxWithMSPID := context.WithValue(ctx, CtxMSPIDKey, "OrgMSP")

	extracted, err := ExtractMSPID(ctxWithMSPID)
	assert.NoError(t, err, "extraction should not fail")
	assert.Equal(t, "OrgMSP", extracted, "MSPID should be extracted from context")

	_, err = ExtractMSPID(ctx)
	assert.Error(t, err, "Extraction should fail on empty context")
}

func TestVerifyClientMSPID(t *testing.T) {

	MSPID := "my-msp-id"
	org1AuthKeyID := []byte{12, 34, 03}
	org2AuthKeyID := []byte{32, 124, 77}

	interceptor := MSPIDInterceptor{
		orgCaCertIDs: map[string][]string{
			"my-msp-id": {hex.EncodeToString(org1AuthKeyID)},
			"otherorg":  {hex.EncodeToString(org2AuthKeyID)},
		},
	}

	var verify = func(isValid bool, p *peer.Peer) func(*testing.T) {
		return func(t *testing.T) {
			ctx := context.TODO()
			ctx = peer.NewContext(ctx, p)
			err := interceptor.verifyClientMSPID(ctx, MSPID)
			msg := "Should return a validation error"
			if isValid {
				msg = fmt.Sprintf("Should not return a validation error: %v", err)
			}
			assert.Equal(t, isValid, err == nil, msg)
		}
	}

	t.Run("Certificate with a missing peer", verify(false, nil))

	t.Run("Certificate with an invalid peer", verify(false, &peer.Peer{}))

	t.Run("Certificate with an incorrect auth info type", verify(false, &peer.Peer{AuthInfo: fakeTLSAuthInfo{}}))

	t.Run("Certificate with an empty TLS info", verify(false, &peer.Peer{AuthInfo: credentials.TLSInfo{}}))

	t.Run("Certificate with an empty connection state", verify(false,
		&peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{}}}))

	t.Run("Certificate with an empty certificate", verify(false,
		&peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{{}}},
		}}))

	t.Run("Certificate with an empty subject", verify(false,
		&peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{{Subject: pkix.Name{}}}},
		}}))

	t.Run("Certificate with an empty list of organizations", verify(false,
		&peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{{Subject: pkix.Name{Organization: []string{}}}}},
		}}))

	t.Run("Certificate with a valid MSPID and CA", verify(true,
		&peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{{Subject: pkix.Name{Organization: []string{MSPID}}, AuthorityKeyId: org1AuthKeyID}}},
		}}))

	t.Run("Certificate with both a valid MSPID+CA and and invalid one", verify(true,
		&peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{{Subject: pkix.Name{Organization: []string{"other mspid", MSPID}}, AuthorityKeyId: org1AuthKeyID}}},
		}}))

	t.Run("Certificate with a valid MSPID but invalid CA for org", verify(false,
		&peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{{Subject: pkix.Name{Organization: []string{MSPID}}, AuthorityKeyId: org2AuthKeyID}}},
		}}))

	t.Run("Certificate with both a valid MSPID and invalid one but invalid CA for org", verify(false,
		&peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{{Subject: pkix.Name{Organization: []string{"other mspid", MSPID}}, AuthorityKeyId: org2AuthKeyID}}},
		}}))
}
