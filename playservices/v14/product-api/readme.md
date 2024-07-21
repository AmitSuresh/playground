cd ./../currency

go run server.go

cd ./../product-api

go run client.go

curl localhost:9090/products?id=669d104d954dd4e12e6f82a7&currency=USD

make sure serverAddress is the same on server.go and client.go. Issue: Port was 9092 in 1 file and 9090 in the client.