package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/owkin/orchestrator/lib/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const headerMSPID = "mspid"

var ignoredMethods = [...]string{
	"grpc.health",
}

// InterceptMSPID is a gRPC interceptor and will make the MSPID from headers available to request context
func InterceptMSPID(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Passthrough for ignored methods
	for _, m := range ignoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return handler(ctx, req)
		}
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.NewBadRequest("could not extract metadata")
	}

	if len(md.Get(headerMSPID)) != 1 {
		return nil, errors.NewBadRequest(fmt.Sprintf("missing or invalid header '%s'", headerMSPID))
	}

	MSPID := md.Get(headerMSPID)[0]

	if MustGetEnvFlag("VERIFY_CLIENT_MSP_ID") {
		err := VerifyClientMSPID(ctx, MSPID)
		if err != nil {
			return nil, err
		}
	}

	newCtx := context.WithValue(ctx, CtxMSPIDKey, MSPID)
	return handler(newCtx, req)
}

// VerifyClientMSPID returns an error if the provided MSPID string doesn't match
// one of the Subject Organizations of the provided context's client TLS certificate.
func VerifyClientMSPID(ctx context.Context, MSPID string) error {
	peer, ok := peer.FromContext(ctx)

	if !ok || peer == nil || peer.AuthInfo == nil {
		return errors.NewInternal("error validating client MSPID: failed to extract MSP ID from TLS context")
	}

	tlsInfo, ok := peer.AuthInfo.(credentials.TLSInfo)

	if ok &&
		len(tlsInfo.State.PeerCertificates) != 0 &&
		len(tlsInfo.State.PeerCertificates[0].Subject.Organization) != 0 {

		orgs := tlsInfo.State.PeerCertificates[0].Subject.Organization
		for _, org := range orgs {
			if org == MSPID {
				return nil // OK
			}
		}
	}

	return errors.NewPermissionDenied(fmt.Sprintf("invalid client MSPID: cannot find MSPID %v in client TLS certificate Subject Organizations", MSPID))
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
	invocator, ok := ctx.Value(CtxMSPIDKey).(string)
	if !ok {
		return "", errors.NewInternal("MSPID not found in context")
	}
	return invocator, nil
}
