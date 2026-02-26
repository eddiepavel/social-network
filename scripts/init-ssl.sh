#!/bin/bash
# Usage: ./init-ssl.sh yourdomain.com your@email.com
# Run this ONCE on your VPS to get the initial SSL certificate.

set -e

DOMAIN=$1
EMAIL=$2

if [ -z "$DOMAIN" ] || [ -z "$EMAIL" ]; then
  echo "Usage: $0 <domain> <email>"
  exit 1
fi

echo "==> Obtaining SSL certificate for $DOMAIN..."

# Step 1: Start nginx with a temporary HTTP-only config for the ACME challenge
# Create a minimal nginx config that only serves the challenge
mkdir -p /tmp/nginx-init
cat > /tmp/nginx-init/default.conf <<EOF
server {
    listen 80;
    server_name $DOMAIN;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 200 'Setting up SSL...';
        add_header Content-Type text/plain;
    }
}
EOF

# Step 2: Start a temporary nginx container for the challenge
docker compose -f docker-compose.prod.yml up -d nginx || true
docker compose -f docker-compose.prod.yml stop nginx

docker run -d --name nginx-init \
  -p 80:80 \
  -v /tmp/nginx-init/default.conf:/etc/nginx/conf.d/default.conf:ro \
  -v social-network_certbot-webroot:/var/www/certbot \
  nginx:alpine

# Step 3: Run certbot to get the certificate
docker run --rm \
  -v social-network_certbot-webroot:/var/www/certbot \
  -v social-network_certbot-certs:/etc/letsencrypt \
  certbot/certbot certonly \
    --webroot \
    --webroot-path=/var/www/certbot \
    --email "$EMAIL" \
    --agree-tos \
    --no-eff-email \
    -d "$DOMAIN"

# Step 4: Clean up temporary container
docker stop nginx-init && docker rm nginx-init
rm -rf /tmp/nginx-init

echo ""
echo "==> SSL certificate obtained successfully!"
echo "==> Now start the full stack:"
echo "    export DOMAIN=$DOMAIN"
echo "    export DOCKER_USERNAME=<your-dockerhub-user>"
echo "    docker compose -f docker-compose.prod.yml up -d"
