package interceptors

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/server/common"
	"github.com/substra/orchestrator/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const headerMSPID = "mspid"

type MSPIDInterceptor struct {
	checkMSPID   bool
	orgCaCertIDs common.OrgCACertList
}

// NewMSPIDInterceptor instanciate a new interceptor
func NewMSPIDInterceptor() (*MSPIDInterceptor, error) {
	var orgCACerts common.OrgCACertList
	verifyClientMSPID := false

	if common.MustGetEnvFlag("VERIFY_CLIENT_MSP_ID") {
		verifyClientMSPID = true
		var err error
		orgCACerts, err = common.GetOrgCACerts()
		if err != nil {
			return nil, err
		}

		log.Debug().Interface("orgCACerts", orgCACerts).Msg("MSP ID will be checked")
	}

	return &MSPIDInterceptor{
		checkMSPID:   verifyClientMSPID,
		orgCaCertIDs: orgCACerts,
	}, nil
}

// UnaryServerInterceptor enforces MSPID presence in context
func (i *MSPIDInterceptor) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Passthrough for ignored methods
	for _, m := range IgnoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return handler(ctx, req)
		}
	}

	newCtx, err := i.extractFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return handler(newCtx, req)
}

// StreamServerInterceptor enforces MSPID presence in context
func (i *MSPIDInterceptor) StreamServerInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	newCtx, err := i.extractFromContext(stream.Context())
	if err != nil {
		return err
	}
	streamWithContext := common.BindStreamToContext(newCtx, stream)

	return handler(srv, streamWithContext)
}

func (i *MSPIDInterceptor) extractFromContext(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.NewBadRequest("could not extract metadata")
	}

	if len(md.Get(headerMSPID)) != 1 {
		return nil, errors.NewBadRequest(fmt.Sprintf("missing or invalid header '%s'", headerMSPID))
	}

	MSPID := md.Get(headerMSPID)[0]

	if i.checkMSPID {
		err := i.verifyClientMSPID(ctx, MSPID)
		if err != nil {
			return nil, err
		}
	}

	return context.WithValue(ctx, CtxMSPIDKey, MSPID), nil
}

// VerifyClientMSPID returns an error if the provided MSPID string doesn't match
// one of the Subject Organizations of the provided context's client TLS certificate
// or if the issuer is not valid for the given organization.
func (i *MSPIDInterceptor) verifyClientMSPID(ctx context.Context, mspid string) error {
	logger := log.Ctx(ctx).With().Str("MSPID", mspid).Logger()
	peer, ok := peer.FromContext(ctx)

	if !ok || peer == nil || peer.AuthInfo == nil {
		return errors.NewInternal("invalid MSPID: failed to extract MSP ID from TLS context")
	}

	tlsInfo, ok := peer.AuthInfo.(credentials.TLSInfo)

	orgMatchCert := false

	if ok &&
		len(tlsInfo.State.PeerCertificates) != 0 &&
		len(tlsInfo.State.PeerCertificates[0].Subject.Organization) != 0 {

		orgs := tlsInfo.State.PeerCertificates[0].Subject.Organization
		logger.Debug().Strs("orgs", orgs).Msg("checking MSPID against cert organizations")
		for _, org := range orgs {
			if org == mspid {
				orgMatchCert = true
				break
			}
		}
	}

	if !orgMatchCert {
		return errors.NewPermissionDenied(fmt.Sprintf("invalid MSPID: cannot find MSPID %q in client TLS certificate Subject Organizations", mspid))
	}

	certIDs, ok := i.orgCaCertIDs[mspid]
	if !ok {
		return errors.NewPermissionDenied(fmt.Sprintf("invalid MSPID: cannot find MSPID %q in allowed organizations", mspid))
	}

	validOrgCA := false
	for _, cert := range tlsInfo.State.PeerCertificates {
		authKeyID := hex.EncodeToString(cert.AuthorityKeyId)
		logger.Debug().
			Strs("orgCertIDs", certIDs).
			Str("clientAuthKeyID", authKeyID).
			Msg("checking that client cert is signed by legitimate CA for organization")
		if utils.SliceContains(certIDs, authKeyID) {
			validOrgCA = true
			break
		}
	}

	if validOrgCA {
		return nil
	}

	return errors.NewPermissionDenied(fmt.Sprintf("invalid MSPID: invalid issuer for MSPID %q", mspid))
}

type ctxMSPIDMarker struct{}

var (
	// CtxMSPIDKey is the identifier of the MSPID in context.
	// Prefer the convenient ExtractMSPID method to retrieve the MSPID from context.
	CtxMSPIDKey = &ctxMSPIDMarker{}
)

// ExtractMSPID retrieves MSPID from request context
// MSPID is expected to be set by InterceptMSPID
func ExtractMSPID(ctx context.Context) (string, error) {
	mspid, ok := ctx.Value(CtxMSPIDKey).(string)
	if !ok {
		return "", errors.NewInternal("MSPID not found in context")
	}
	return mspid, nil
}
