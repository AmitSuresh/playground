check_install:
	which swagger || GO111MODULE=off go get github.com/go-swagger/go-swagger/cmd/swagger

swagger: check_install
	GO111MODULE=off go run github.com/go-swagger/go-swagger/cmd/swagger generate spec -o ./swagger.yaml --scan-models

generate_client:
	cd sdk && go run github.com/go-swagger/go-swagger/cmd/swagger generate client -f ../swagger.yaml -A playservices