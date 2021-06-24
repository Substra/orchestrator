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

	ctxWithMSPID := context.WithValue(ctx, CtxMSPIDKey, "OrgMSP")

	extracted, err := ExtractMSPID(ctxWithMSPID)
	assert.NoError(t, err, "extraction should not fail")
	assert.Equal(t, "OrgMSP", extracted, "MSPID should be extracted from context")

	_, err = ExtractMSPID(ctx)
	assert.Error(t, err, "Extraction should fail on empty context")
}

func TestVerifyClientMSPID(t *testing.T) {

	MSPID := "my-msp-id"

	var verify = func(isValid bool, p *peer.Peer) func(*testing.T) {
		return func(t *testing.T) {
			ctx := context.TODO()
			ctx = peer.NewContext(ctx, p)
			err := VerifyClientMSPID(ctx, MSPID)
			msg := "Should return a validation error"
			if isValid {
				msg = "Should not return a validation error"
			}
			assert.Equal(t, isValid, err == nil, msg)
		}
	}

	t.Run("Certificate with a missing peer", verify(false, nil))

	t.Run("Certificate with an invalid peer", verify(false, &peer.Peer{}))

	t.Run("Certificate with an incorrect auth info type", verify(false, &peer.Peer{AuthInfo: FakeTLSAuthInfo{}}))

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

	t.Run("Certificate with a valid MSPID", verify(true,
		&peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{{Subject: pkix.Name{Organization: []string{MSPID}}}}},
		}}))

	t.Run("Certificate with both a valid MSPID and and invalid one", verify(true,
		&peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{{Subject: pkix.Name{Organization: []string{"other mspid", MSPID}}}}},
		}}))
}
