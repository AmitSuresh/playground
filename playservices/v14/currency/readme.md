//protoc -I=currency/protos currency/protos/currency.proto --go_out=currency/protos --go-grpc_out=require_unimplemented_servers=false:currency/protos
//protoc -I=protos/ currency.proto --go_out=protos/ --go-grpc_out=require_unimplemented_servers=false:protos/
//protoc -I=protos/ currency.proto --go_out=protos/ --go-grpc_out=protos/
//protoc -I=currency/protos currency/protos/currency.proto --go_out=currency/protos --go-grpc_out=currency/protos
//protoc -I=. currency.proto --go_out=. --go-grpc_out=.