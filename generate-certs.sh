#!/bin/bash

echo "Generating self-signed SSL certificate for development..."

openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes \
  -subj "/C=US/ST=Development/L=Development/O=Backend/OU=Development/CN=localhost"

echo "Certificate generated successfully!"
