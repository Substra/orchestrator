module github.com/owkin/orchestrator

go 1.16

require (
	github.com/Masterminds/squirrel v1.5.1
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/go-playground/log/v7 v7.0.2
	github.com/golang-migrate/migrate/v4 v4.15.1
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20200728190242-9b3ae92d8664
	github.com/hyperledger/fabric-contract-api-go v1.1.2-0.20210104111150-d852efd3f6af
	github.com/hyperledger/fabric-protos-go v0.0.0-20200917184523-71c4060efc42
	github.com/hyperledger/fabric-sdk-go v1.0.1-0.20210729165856-3be4ed253dcf
	github.com/jackc/pgconn v1.10.0
	github.com/jackc/pgerrcode v0.0.0-20201024163028-a0d42d470451
	github.com/jackc/pgx/v4 v4.13.0
	github.com/looplab/fsm v0.3.0
	github.com/onsi/gomega v1.10.4 // indirect
	github.com/pashagolub/pgxmock v1.4.0
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/afero v1.3.4 // indirect
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/testify v1.7.0
	github.com/yoheimuta/go-protoparser/v4 v4.4.0
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
)

// Temporary fork pending https://github.com/hyperledger/fabric-sdk-go/pull/195
replace github.com/hyperledger/fabric-sdk-go => github.com/mblottiere/fabric-sdk-go v1.0.1-0.20211206140729-01275ccd8f71
