# IP Echo Service - Complete Documentation

## 📋 Table of Contents
- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Deployment Scenarios](#deployment-scenarios)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Security](#security)
- [Troubleshooting](#troubleshooting)

---

## Overview

A lightweight, high-performance IP Echo service written in Go. Provides IP address information, geolocation, hostname resolution, and ISP details with multiple output formats.

**Technology Stack:**
- **Language**: Go 1.25
- **Router**: Gorilla Mux
- **GeoIP**: MaxMind GeoLite2 (City + ASN)
- **Container**: Docker (Alpine-based)
- **Reverse Proxy**: Nginx (optional)

---

## Features

### Core Functionality
✅ IP address detection (IPv4/IPv6)  
✅ Reverse DNS hostname lookup  
✅ GeoIP location (city, state, country)  
✅ ISP/Organization detection  
✅ Multiple output formats (JSON, XML, YAML, Plain Text, HTML)  
✅ HTTP request headers inspection  

### Security & Performance
✅ **Rate Limiting**: 10 requests per 10 seconds per IP  
✅ **CORS Enabled**: Public API access from browsers  
✅ **Security Headers**: X-Frame-Options, CSP, X-Content-Type-Options  
✅ **Path Traversal Protection**: Sanitized file serving  
✅ **DNS Timeout**: 2-second timeout to prevent DoS  
✅ **Non-root Container**: Runs as `appuser`  
✅ **Resource Limits**: CPU/Memory caps in Docker  

### API Endpoints
- `GET /` - HTML interface (or plain text for curl)
- `GET /json` - JSON response
- `GET /yaml` - YAML response
- `GET /xml` - XML response
- `GET /text` - Plain text IP only
- `GET /clean` - Clean HTML (no extras)
- `GET /headers` - View HTTP headers

---

## Architecture

```
┌─────────────────┐
│   Client        │
│  (Browser/API)  │
└────────┬────────┘
         │
    ┌────▼─────────────────┐
    │  Nginx (Optional)    │  ← HTTPS/SSL Termination
    │  Reverse Proxy       │     Rate limiting (additional)
    └────┬─────────────────┘     Load balancing
         │
    ┌────▼─────────────────┐
    │   Docker Container   │
    │   ┌──────────────┐   │
    │   │  IP Service  │   │  ← Rate limiting (built-in)
    │   │   Port 8080  │   │     CORS handling
    │   └──────────────┘   │     Security headers
    │                      │
    │   ┌──────────────┐   │
    │   │ GeoIP DBs    │   │  ← Optional geolocation
    │   │ (MaxMind)    │   │
    │   └──────────────┘   │
    └──────────────────────┘
```

---

## Installation

### Prerequisites
- **Docker** (20.10+)
- **Docker Compose** (2.0+)
- **MaxMind License Key** (free, for GeoIP features)

### Quick Start

1. **Clone/Download** the project
   ```bash
   cd /path/to/ip_service
   ```

2. **Run Setup Script** (downloads GeoIP databases)
   ```bash
   ./setup.sh
   ```
   - Enter your MaxMind License Key when prompted
   - Get a free key: https://www.maxmind.com/en/geolite2/signup

3. **Start the Service**
   ```bash
   # Development (localhost:8090)
   docker compose up -d --build
   
   # Or production (port 80)
   docker compose -f docker-compose.prod.yaml up -d --build
   ```

4. **Test**
   ```bash
   curl http://localhost:8090/json
   ```

---

## Deployment Scenarios

### Scenario 1: Direct HTTP (Port 80)
**Use case**: Simple deployment on a VM with no other services

```bash
# Use production docker-compose
docker compose -f docker-compose.prod.yaml up -d --build
```

**Configuration**:
- Service listens on port **80**
- No SSL/TLS
- Good for: Internal networks, testing, HTTP-only services

---

### Scenario 2: Direct HTTPS (Port 443)
**Use case**: Public-facing service with built-in TLS

**Steps**:
1. Obtain SSL certificate and key (e.g., Let's Encrypt)
2. Place files in a secure directory (e.g., `/etc/ssl/certs/`)
3. Run with TLS flags:

```bash
docker run -d \
  -p 443:8443 \
  -v /etc/ssl/certs:/certs:ro \
  ip-service \
  ./ip-service -tls -cert /certs/fullchain.pem -key /certs/privkey.pem -addr :8443
```

**Or modify docker-compose**:
```yaml
services:
  ip-service:
    # ... other config ...
    ports:
      - "443:8443"
    volumes:
      - /etc/ssl/certs:/certs:ro
    command: ["./ip-service", "-tls", "-cert", "/certs/fullchain.pem", "-key", "/certs/privkey.pem", "-addr", ":8443"]
```

---

### Scenario 3: Behind Nginx Reverse Proxy (RECOMMENDED)
**Use case**: Production deployment with SSL termination, multiple services

**Architecture**:
- Nginx handles: SSL/TLS, compression, caching, additional rate limiting
- IP Service handles: Application logic

**Setup**:

1. **Start IP Service** (internal port)
   ```bash
   docker compose up -d --build
   # Service runs on 127.0.0.1:8090
   ```

2. **Configure Nginx** (`/etc/nginx/conf.d/ip-service.conf`):

   ```nginx
   server {
       listen 80;
       listen [::]:80;
       server_name ip.yourdomain.com;
       
       # Redirect HTTP to HTTPS
       return 301 https://$host$request_uri;
   }

   server {
       listen 443 ssl http2;
       listen [::]:443 ssl http2;
       server_name ip.yourdomain.com;
       
       # SSL Configuration
       ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
       ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
       ssl_protocols TLSv1.2 TLSv1.3;
       ssl_ciphers HIGH:!aNULL:!MD5;
       
       # Proxy to Docker container
       location / {
           proxy_pass http://127.0.0.1:8090;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
           proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
           proxy_set_header X-Forwarded-Proto $scheme;
           
           # Timeouts
           proxy_connect_timeout 5s;
           proxy_send_timeout 10s;
           proxy_read_timeout 10s;
       }
   }
   ```

3. **Reload Nginx**:
   ```bash
   sudo nginx -t
   sudo systemctl reload nginx
   ```

---

### Scenario 4: Cloud Deployment (Azure/GCP/AWS)

#### Google Cloud Run (Serverless):

**Best for**: Auto-scaling, zero maintenance, automatic HTTPS.

1. **Build and push image**:
   ```bash
   cd service
   gcloud builds submit --tag us-east1-docker.pkg.dev/YOUR_PROJECT/cloud-run-source-deploy/ip-echo-service
   ```

2. **Deploy with MaxMind Credentials**:
   ```bash
   gcloud run deploy ip-echo-service \
     --image us-east1-docker.pkg.dev/YOUR_PROJECT/cloud-run-source-deploy/ip-echo-service \
     --region us-east1 \
     --allow-unauthenticated \
     --set-env-vars "MAXMIND_ACCOUNT_ID=YOUR_ID,MAXMIND_LICENSE_KEY=YOUR_KEY"
   ```

#### Azure VM:
```bash
# 1. Create VM with public IP
# 2. Open ports 80/443 in Network Security Group
# 3. SSH into VM
# 4. Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 5. Deploy
git clone <your-repo> ip-service
cd ip-service
./setup.sh
docker compose -f docker-compose.prod.yaml up -d --build
```

#### GCP Compute Engine:
```bash
# Similar to Azure, use gcloud CLI
gcloud compute instances create ip-service \
  --image-family=ubuntu-2204-lts \
  --image-project=ubuntu-os-cloud \
  --tags=http-server,https-server
```

#### AWS EC2:
```bash
# Launch Ubuntu instance with security group allowing 80/443
# Install Docker and deploy as above
```

---

## Configuration

### Command-Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--static` | `./static` | Path to static files (templates, favicon) |
| `--geoip` | `./geoip` | Path to MaxMind GeoIP databases |
| `--addr` | `:8080` | Listen address (e.g., `:80`, `0.0.0.0:8080`) |
| `--tls` | `false` | Enable TLS/HTTPS |
| `--cert` | `""` | Path to TLS certificate file |
| `--key` | `""` | Path to TLS key file |

### Environment Variables

**Docker Compose** (`docker-compose.yaml`):
```yaml
environment:
  - CUSTOM_VAR=value  # Add custom env vars here
```

### Resource Limits

**Development** (`docker-compose.yaml`):
- CPU: 0.5 cores
- Memory: 256MB

**Production** (`docker-compose.prod.yaml`):
- CPU: 1.0 cores
- Memory: 512MB

**Adjust in docker-compose**:
```yaml
deploy:
  resources:
    limits:
      cpus: '2.0'      # Increase for high traffic
      memory: 1024M
```

---

## API Reference

### Response Formats

#### JSON (`/json`)
```json
{
  "ip_address": "203.0.113.42",
  "hostname": "example.com",
  "isp": "Example ISP Ltd",
  "city": "San Francisco",
  "country": "United States",
  "country_code": "US",
  "location": "San Francisco, CA, United States"
}
```

#### XML (`/xml`)
```xml
<ip-response>
  <ip-address>203.0.113.42</ip-address>
  <location>San Francisco, CA, United States</location>
  <hostname>example.com</hostname>
  <isp>Example ISP Ltd</isp>
  <city>San Francisco</city>
  <country>United States</country>
  <country-code>US</country-code>
</ip-response>
```

#### YAML (`/yaml`)
```yaml
ip_address: 203.0.113.42
location: San Francisco, CA, United States
hostname: example.com
isp: Example ISP Ltd
city: San Francisco
country: United States
country_code: US
```

#### Plain Text (`/text`)
```
203.0.113.42
```

### CORS Support

All endpoints support CORS for browser-based JavaScript applications:

```javascript
// Works from any website
fetch('https://ip.yourdomain.com/json')
  .then(response => response.json())
  .then(data => console.log('My IP:', data.ip_address));
```

### Rate Limiting

- **Limit**: 10 requests per 10 seconds per IP
- **Response**: HTTP 429 when exceeded
- **Headers**: No rate limit headers (consider adding if needed)

---

## Security

### Built-in Protection

1. **Path Traversal**: All file paths sanitized
2. **DNS DoS**: 2-second timeout on DNS lookups
3. **Rate Limiting**: Per-IP token bucket
4. **Security Headers**: CSP, X-Frame-Options, X-Content-Type-Options
5. **CORS**: Controlled cross-origin access
6. **Non-root User**: Container runs as `appuser` (UID varies)
7. **Resource Limits**: Prevents resource exhaustion

### Additional Recommendations

**For Production**:
- Use HTTPS (Nginx reverse proxy recommended)
- Enable firewall (UFW, iptables, cloud security groups)
- Monitor logs: `docker logs ip-service`
- Update regularly: `docker compose pull && docker compose up -d`
- Consider adding Fail2ban for additional rate limiting

---

## Troubleshooting

### Service won't start
```bash
# Check logs
docker logs ip-service

# Common issues:
# 1. Port already in use
sudo netstat -tulpn | grep :8080

# 2. Permission denied
sudo chmod +x setup.sh

# 3. Docker daemon not running
sudo systemctl start docker
```

### GeoIP not working
```bash
# Check databases exist
ls -lh service/geoip/

# Should see:
# ipcity.mmdb
# iporg.mmdb

# Re-run setup
./setup.sh
```

### 429 Rate Limit Errors
- Wait 10 seconds between request bursts
- For testing, temporarily remove rate limit middleware in `main.go`
- For production, adjust limits in code or add Nginx rate limiting

### Health Check Failing
```bash
# Manual health check
docker exec ip-service wget -q -O - http://localhost:8080

# Should return your server's IP
```

---

## Maintenance

### Updating GeoIP Databases

MaxMind updates databases regularly. Update monthly:

```bash
./setup.sh  # Will re-download if you provide license key
docker compose restart
```

### Viewing Metrics

```bash
# Container stats
docker stats ip-service

# Logs (real-time)
docker logs -f ip-service

# Last 100 lines
docker logs --tail 100 ip-service
```

### Backing Up

Important files to backup:
- `service/geoip/*.mmdb` (or just re-download)
- SSL certificates (if using TLS mode)
- `docker-compose.yaml` (if customized)

---

## Performance Tuning

### For High Traffic

1. **Increase Resources**:
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '4.0'
         memory: 2048M
   ```

2. **Disable DNS Lookups** (faster):
   Comment out reverse DNS in `main.go` line ~208

3. **Use Nginx Caching**:
   ```nginx
   proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=ip_cache:10m;
   
   location / {
       proxy_cache ip_cache;
       proxy_cache_valid 200 1m;
       # ... rest of config
   }
   ```

4. **Run Multiple Instances** (load balancing):
   ```yaml
   services:
     ip-service:
       deploy:
         replicas: 3
   ```

---

## License

MIT License - See `LICENSE` file

---

## Support & Contributing

- **Issues**: Check logs first, then open GitHub issue
- **Security**: Report vulnerabilities privately
- **Pull Requests**: Welcome! Follow Go conventions

---

**Version**: 2.1  
**Last Updated**: 2026-01-19
