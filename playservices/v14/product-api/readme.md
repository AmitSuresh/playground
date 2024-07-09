cd ./../currency

go run server.go

cd ./../product-api

go run client.go

curl localhost:9090/products/2

result:
{"id":2,"name":"Espresso","description":"Short and strong coffee without milk","price":0.4975,"sku":"fjd34"}

make sure serverAddress is the same on server.go and client.go. Issue: Port was 9092 in 1 file and 8080 in the other.