# Quick Reference Card

## Common Commands

### Start/Stop Service
```bash
# Development
docker compose up -d              # Start
docker compose down               # Stop
docker compose restart            # Restart
docker compose logs -f            # View logs

# Production
docker compose -f docker-compose.prod.yaml up -d
docker compose -f docker-compose.prod.yaml down
```

### Testing Endpoints
```bash
# Plain text (curl detection)
curl http://localhost:8090

# JSON
curl http://localhost:8090/json | jq .

# YAML
curl http://localhost:8090/yaml

# XML
curl http://localhost:8090/xml

# Headers
curl http://localhost:8090/headers

# With external IP (via nginx/domain)
curl http://yourdomain.com/json
```

### Rate Limit Testing
```bash
# Should see 200 for first 10, then 429
for i in {1..12}; do 
  curl -s -o /dev/null -w "Request $i: %{http_code}\n" http://localhost:8090
done
```

### Container Management
```bash
# View running containers
docker ps

# Container stats
docker stats ip-service

# Shell into container
docker exec -it ip-service sh

# View health status
docker inspect ip-service | jq '.[0].State.Health'
```

### Nginx Commands
```bash
# Test config
sudo nginx -t

# Reload (no downtime)
sudo systemctl reload nginx

# Restart
sudo systemctl restart nginx

# View Nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

## Port Reference

| Service | Port | Access | Notes |
|---------|------|--------|-------|
| IP Service (dev) | 8090 | `127.0.0.1:8090` | Local only |
| IP Service (prod) | 80 | `0.0.0.0:80` | Public HTTP |
| IP Service (TLS) | 8443 | Configurable | Direct HTTPS |
| Nginx | 80/443 | Public | Reverse proxy |

## File Locations

### In Container
- Binary: `/app/ip-service`
- Static files: `/app/static/`
- GeoIP DBs: `/app/geoip/` (optional)

### On Host
- Project: `/path/to/ip_service/`
- Service code: `service/main.go`
- Docker compose: `docker-compose.yaml`
- Nginx config: `/etc/nginx/conf.d/` (if using reverse proxy)
- SSL certs (if used): `/etc/letsencrypt/` or custom path

## Flag Reference

```bash
./ip-service \
  -static ./static \        # Template directory
  -geoip ./geoip \          # GeoIP database directory
  -addr :8080 \             # Listen address
  -tls \                    # Enable TLS (requires -cert and -key)
  -cert /path/cert.pem \    # TLS certificate
  -key /path/key.pem        # TLS private key
```

## Response Structure

### JSON Fields
```json
{
  "ip_address": "string",      // Client IP
  "hostname": "string",        // Reverse DNS
  "isp": "string",            // ISP/Org from GeoIP
  "city": "string",           // City from GeoIP
  "country": "string",        // Country name
  "country_code": "string",   // ISO country code
  "location": "string"        // Combined location
}
```

## Security Defaults

| Feature | Default | Customizable |
|---------|---------|--------------|
| Rate Limit | 10 req/10s per IP | Yes (code) |
| CORS | Enabled (`*`) | Yes (code) |
| Max Request Size | Go default (~10MB) | Yes (code) |
| DNS Timeout | 2 seconds | Yes (code) |
| Container User | `appuser` (non-root) | No |
| CPU Limit | 0.5-1.0 cores | Yes (compose) |
| RAM Limit | 256-512MB | Yes (compose) |

## Troubleshooting Quick Checks

```bash
# Is service running?
docker ps | grep ip-service

# Can I reach it locally?
curl http://localhost:8090

# Check logs for errors
docker logs ip-service --tail 50

# Is port available?
sudo ss -tulpn | grep :8090

# Is nginx working?
sudo nginx -t && sudo systemctl status nginx

# GeoIP loaded?
docker logs ip-service | grep -i geoip
```

## Common Issues & Solutions

### Port 8090 already in use
```bash
# Find what's using it
sudo lsof -i :8090
# Kill or change port in docker-compose.yaml
```

### GeoIP databases missing
```bash
# Re-run setup
./setup.sh
# Or manually place in service/geoip/
```

### Rate limited locally
```bash
# Wait 10 seconds or restart container to reset
docker compose restart
```

### Nginx can't reach service
```bash
# Check docker port binding
docker port ip-service
# Should show: 8080/tcp -> 127.0.0.1:8090
```

## URLs

- **MaxMind Signup**: https://www.maxmind.com/en/geolite2/signup
- **Docker Docs**: https://docs.docker.com
- **Nginx Docs**: https://nginx.org/en/docs/
- **Go Rate Package**: https://pkg.go.dev/golang.org/x/time/rate
