package ledger

import (
	"errors"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/substrafoundation/substra-orchestrator/lib/persistence"
)

// GetLedgerFromContext will return the ledger DB from invocation context
func GetLedgerFromContext(ctx contractapi.TransactionContextInterface) (persistence.Database, error) {
	stub := ctx.GetStub()

	return &DB{ccStub: stub}, nil
}

// DB is the distributed ledger persistence layer implementing persistence.Database
type DB struct {
	ccStub shim.ChaincodeStubInterface
}

// PutState stores data in the ledger
func (l *DB) PutState(resource string, key string, data []byte) error {
	return l.ccStub.PutState(key, data)
}
func (l *DB) GetState(resource string, key string) ([]byte, error) {
	return l.ccStub.GetState(key)
}
func (l *DB) GetAll(resource string) ([][]byte, error) {
	return [][]byte{}, errors.New("unimplemented")
}
