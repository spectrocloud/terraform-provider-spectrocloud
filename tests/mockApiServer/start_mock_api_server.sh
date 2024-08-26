#!/bin/bash

export MOCK_SERVER_PATH="$TF_SRC/tests/mockApiServer"

# Navigate to the mock API server directory
cd $MOCK_SERVER_PATH || exit

# Generate the private key
openssl genpkey -algorithm RSA -out mock_server.key -pkeyopt rsa_keygen_bits:2048

# Generate the self-signed certificate with default input
openssl req -new -x509 -key mock_server.key -out mock_server.crt -days 365 -subj "/C=US/ST=CA/L=City/O=Organization/OU=Department/CN=localhost"

# Build the Go project
go build -o MockAPIServer main.go

# Run the server in the background and redirect output to server.log
nohup ./MockAPIServer > mock_api_server.log 2>&1 &