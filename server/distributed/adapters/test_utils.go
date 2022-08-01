package adapters

import "github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"

var FabricTimeout = status.New(status.ClientStatus, status.Timeout.ToInt32(), "request timed out or been cancelled", nil)
