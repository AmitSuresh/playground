docker build -t client-server .

docker run -d --network web -p 9090:9090 --env-file /c/"Program Files"/Go/src/goworkspace/github.com/AmitSuresh/playground/playservices/v14/product-api/.env --name client-server client-server

docker build -t grpc-server .

docker run -d --network web -p 9092:9092 --env-file /c/"Program Files"/Go/src/goworkspace/github.com/AmitSuresh/playground/playservices/v14/currency/.env --name grpc-server grpc-server

docker stop grpc-server

docker build -t db-server .

docker run -d --network web -p 9090:9090 --env-file /c/"Program Files"/Go/src/goworkspace/github.com/AmitSuresh/playground/db-server/.env --name db-server db-server
docker stop db-server

docker run -d --network web --name db -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres


docker stop $(docker ps -aq)
docker rm $(docker ps -aq)
docker rmi $(docker images -q)
docker network rm $(docker network ls -q)
docker volume rm $(docker volume ls -q)
docker system prune -a --volumes

docker-compose up -d
docker-compose down
docker images
docker rmi <image-id-or-name>
docker rm -f db-server

docker ps
docker logs <client-app-container-id>
docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' <client-app-container-id>
netstat -ano | findstr 9093

docker network create web
docker network inspect web

grpcurl -v -plaintext localhost:9092 list
grpcurl -v -plaintext grpc-server:9092 list

check_install:
    which swagger || go install github.com/go-swagger/go-swagger/cmd/swagger

swagger: check_install
    swagger generate spec -o ./docs/swagger.yaml --scan-models

    