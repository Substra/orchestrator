// Package testing provides helpers to ease testing chaincode related code
package testing

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/mock"
)

// MockedStub implements ChaincodeStubInterface
type MockedStub struct {
	mock.Mock
}

// GetArgs is a mock
func (m *MockedStub) GetArgs() [][]byte {
	args := m.Called()
	return args.Get(0).([][]byte)
}

// GetStringArgs is a mock
func (m *MockedStub) GetStringArgs() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

// GetFunctionAndParameters is a mock
func (m *MockedStub) GetFunctionAndParameters() (string, []string) {
	args := m.Called()
	return args.String(0), args.Get(1).([]string)
}

// GetArgsSlice is a mock
func (m *MockedStub) GetArgsSlice() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

// GetTxID is a mock
func (m *MockedStub) GetTxID() string {
	args := m.Called()
	return args.String(0)
}

// GetChannelID is a mock
func (m *MockedStub) GetChannelID() string {
	args := m.Called()
	return args.String(0)
}

// InvokeChaincode is a mock
func (m *MockedStub) InvokeChaincode(chaincodeName string, args [][]byte, channel string) pb.Response {
	a := m.Called(chaincodeName, args, channel)
	return a.Get(0).(pb.Response)
}

// GetState is a mock
func (m *MockedStub) GetState(key string) ([]byte, error) {
	args := m.Called(key)
	return args.Get(0).([]byte), args.Error(1)
}

// PutState is a mock
func (m *MockedStub) PutState(key string, value []byte) error {
	args := m.Called(key, value)
	return args.Error(0)
}

// DelState is a mock
func (m *MockedStub) DelState(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

// SetStateValidationParameter is a mock
func (m *MockedStub) SetStateValidationParameter(key string, ep []byte) error {
	args := m.Called(key, ep)
	return args.Error(0)
}

// GetStateValidationParameter is a mock
func (m *MockedStub) GetStateValidationParameter(key string) ([]byte, error) {
	args := m.Called(key)
	return args.Get(0).([]byte), args.Error(1)
}

// GetStateByRange is a mock
func (m *MockedStub) GetStateByRange(startKey, endKey string) (shim.StateQueryIteratorInterface, error) {
	args := m.Called(startKey, endKey)
	return args.Get(0).(shim.StateQueryIteratorInterface), args.Error(1)
}

// GetStateByRangeWithPagination is a mock
func (m *MockedStub) GetStateByRangeWithPagination(startKey, endKey string, pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {
	args := m.Called(startKey, endKey, pageSize, bookmark)
	return args.Get(0).(shim.StateQueryIteratorInterface), args.Get(1).(*pb.QueryResponseMetadata), args.Error(2)
}

// GetStateByPartialCompositeKey is a mock
func (m *MockedStub) GetStateByPartialCompositeKey(objectType string, keys []string) (shim.StateQueryIteratorInterface, error) {
	args := m.Called(objectType, keys)
	return args.Get(0).(shim.StateQueryIteratorInterface), args.Error(1)
}

// GetStateByPartialCompositeKeyWithPagination is a mock
func (m *MockedStub) GetStateByPartialCompositeKeyWithPagination(objectType string, keys []string, pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {
	args := m.Called(objectType, keys, pageSize, bookmark)
	return args.Get(0).(shim.StateQueryIteratorInterface), args.Get(1).(*pb.QueryResponseMetadata), args.Error(2)
}

// CreateCompositeKey is a mock
func (m *MockedStub) CreateCompositeKey(objectType string, attributes []string) (string, error) {
	args := m.Called(objectType, attributes)
	return args.String(0), args.Error(1)
}

// SplitCompositeKey is a mock
func (m *MockedStub) SplitCompositeKey(compositeKey string) (string, []string, error) {
	args := m.Called(compositeKey)
	return args.String(0), args.Get(1).([]string), args.Error(2)
}

// GetQueryResult is a mock
func (m *MockedStub) GetQueryResult(query string) (shim.StateQueryIteratorInterface, error) {
	args := m.Called(query)
	return args.Get(0).(shim.StateQueryIteratorInterface), args.Error(1)
}

// GetQueryResultWithPagination is a mock
func (m *MockedStub) GetQueryResultWithPagination(query string, pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {
	args := m.Called(query, pageSize, bookmark)
	return args.Get(0).(shim.StateQueryIteratorInterface), args.Get(1).(*pb.QueryResponseMetadata), args.Error(2)
}

// GetHistoryForKey is a mock
func (m *MockedStub) GetHistoryForKey(key string) (shim.HistoryQueryIteratorInterface, error) {
	args := m.Called(key)
	return args.Get(0).(shim.HistoryQueryIteratorInterface), args.Error(1)
}

// GetPrivateData is a mock
func (m *MockedStub) GetPrivateData(collection, key string) ([]byte, error) {
	args := m.Called(collection, key)
	return args.Get(0).([]byte), args.Error(1)
}

// GetPrivateDataHash is a mock
func (m *MockedStub) GetPrivateDataHash(collection, key string) ([]byte, error) {
	args := m.Called(collection, key)
	return args.Get(0).([]byte), args.Error(1)
}

// PutPrivateData is a mock
func (m *MockedStub) PutPrivateData(collection string, key string, value []byte) error {
	args := m.Called(collection, key, value)
	return args.Error(0)
}

// DelPrivateData is a mock
func (m *MockedStub) DelPrivateData(collection, key string) error {
	args := m.Called(collection, key)
	return args.Error(0)
}

// PurgePrivateData is a mock
func (m *MockedStub) PurgePrivateData(collection string, key string) error {
	args := m.Called(collection, key)
	return args.Error(0)
}

// SetPrivateDataValidationParameter is a mock
func (m *MockedStub) SetPrivateDataValidationParameter(collection, key string, ep []byte) error {
	args := m.Called(collection, key)
	return args.Error(0)
}

// GetPrivateDataValidationParameter is a mock
func (m *MockedStub) GetPrivateDataValidationParameter(collection, key string) ([]byte, error) {
	args := m.Called(collection, key)
	return args.Get(0).([]byte), args.Error(1)
}

// GetPrivateDataByRange is a mock
func (m *MockedStub) GetPrivateDataByRange(collection, startKey, endKey string) (shim.StateQueryIteratorInterface, error) {
	args := m.Called(collection, startKey, endKey)
	return args.Get(0).(shim.StateQueryIteratorInterface), args.Error(1)
}

// GetPrivateDataByPartialCompositeKey is a mock
func (m *MockedStub) GetPrivateDataByPartialCompositeKey(collection, objectType string, keys []string) (shim.StateQueryIteratorInterface, error) {
	args := m.Called(collection, objectType, keys)
	return args.Get(0).(shim.StateQueryIteratorInterface), args.Error(1)
}

// GetPrivateDataQueryResult is a mock
func (m *MockedStub) GetPrivateDataQueryResult(collection, query string) (shim.StateQueryIteratorInterface, error) {
	args := m.Called(collection, query)
	return args.Get(0).(shim.StateQueryIteratorInterface), args.Error(1)
}

// GetCreator is a mock
func (m *MockedStub) GetCreator() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

// GetTransient is a mock
func (m *MockedStub) GetTransient() (map[string][]byte, error) {
	args := m.Called()
	return args.Get(0).(map[string][]byte), args.Error(1)
}

// GetBinding is a mock
func (m *MockedStub) GetBinding() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

// GetDecorations is a mock
func (m *MockedStub) GetDecorations() map[string][]byte {
	args := m.Called()
	return args.Get(0).(map[string][]byte)
}

// GetSignedProposal is a mock
func (m *MockedStub) GetSignedProposal() (*pb.SignedProposal, error) {
	args := m.Called()
	return args.Get(0).(*pb.SignedProposal), args.Error(1)
}

// GetTxTimestamp is a mock
func (m *MockedStub) GetTxTimestamp() (*timestamp.Timestamp, error) {
	args := m.Called()
	return args.Get(0).(*timestamp.Timestamp), args.Error(1)
}

// SetEvent is a mock
func (m *MockedStub) SetEvent(name string, payload []byte) error {
	args := m.Called(name, payload)
	return args.Error(0)
}

// MockedStateQueryIterator implements StateQueryIterator
type MockedStateQueryIterator struct {
	mock.Mock
}

// HasNext is a mock
func (m *MockedStateQueryIterator) HasNext() bool {
	args := m.Called()
	return args.Bool(0)
}

// Close is a mock
func (m *MockedStateQueryIterator) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Next is a mock
func (m *MockedStateQueryIterator) Next() (*queryresult.KV, error) {
	args := m.Called()
	return args.Get(0).(*queryresult.KV), args.Error(1)
}
