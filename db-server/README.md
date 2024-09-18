# Docker-Based Client-Server Application

This project demonstrates how to set up a client-server application using Docker and Docker Compose, perform HTTP requests with `curl`, and access the Swagger UI for API documentation.

## Prerequisites

Before running the application, make sure you have the following installed:

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)
- A terminal (Linux, macOS, or WSL on Windows)

## Step-by-Step Setup Guide

### 1. Create a New Docker Network

First, create a new network bridge using Docker to allow the services to communicate with each other.

```bash
docker network create web
```

### 2. Start the Services
Use Docker Compose to bring up the services:

```bash
docker-compose up -d
```

### 3. Create a New Order (POST)
Send a POST request to create a new order:

```bash
curl -X POST http://client-server.localhost:80/orders \
-H "Content-Type: application/json" \
-H "x-correlationid: bec3c24e-f068-4b44-b990-35da972d6796" \
-d '{
    "cargoId": 123,
    "lineItems": [
        {
            "productId": 22,
            "sellerId": 234
        }
    ],
    "shipmentNumber": 234
}'
```

### 4. Retrieve an Order by ID (GET)
Fetch an order by its ID (e.g., 1):
```bash
curl -X GET http://client-server.localhost:80/orders/1 \
-H "x-correlationid: bec3c24e-f068-4b44-b990-35da972d6796"
```

### 5. Access Swagger UI
Visit the Swagger UI to explore the API documentation:
http://client-server.localhost/docs

### Teardown
To stop the services and remove the containers, you can run:
```bash
docker-compose down
```

### Key Points:
- **Clear setup steps**: For creating the network, starting services, and making API calls.
- **`curl` examples**: Provided for both POST and GET requests with headers and payloads.
- **Swagger UI access**: Included for easier API exploration.