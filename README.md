# IP Echo Service

A professional, lightweight IP address lookup service with geolocation, built in Go.

## 🚀 Quick Start

```bash
# 1. Get a free MaxMind license key (for GeoIP)
# Visit: https://www.maxmind.com/en/geolite2/signup

# 2. Run setup script
./setup.sh

# 3. Start the service
docker compose up -d --build

# 4. Test it
curl http://localhost:8090/json
```

## 📚 Full Documentation

**See [DOCUMENTATION.md](./DOCUMENTATION.md)** for complete information on:
- All deployment scenarios (direct HTTP, HTTPS, reverse proxy, cloud)
- Full API reference
- Security features
- Configuration options
- Troubleshooting

## ✨ Features

- **Multiple Formats**: JSON, XML, YAML, Plain Text, HTML
- **GeoIP Lookup**: City, country, ISP (requires MaxMind databases)
- **Rate Limited**: 10 requests/10 seconds per IP
- **CORS Enabled**: Use from browser JavaScript
- **Secure**: Non-root container, security headers, path traversal protection
- **Production Ready**: Resource limits, health checks, structured logging

## 🎯 Use Cases

### As a Development Tool
```bash
curl https://ip.yourdomain.com
```

### From JavaScript
```javascript
fetch('https://ip.yourdomain.com/json')
  .then(r => r.json())
  .then(data => console.log(data.ip_address));
```

### With Reverse Proxy (Nginx)
See [Deployment Scenarios](./DOCUMENTATION.md#deployment-scenarios) in full docs.

## 📦 What's Included

```
ip_service/
├── service/
│   ├── main.go              # Application code
│   ├── Dockerfile           # Container definition
│   ├── static/              # HTML templates, assets
│   └── geoip/               # Place GeoIP databases here
├── docker-compose.yaml      # Development (port 8090)
├── docker-compose.prod.yaml # Production (port 80)
├── setup.sh                 # Setup script (downloads GeoIP)
├── DOCUMENTATION.md         # Complete documentation ⭐
├── SECURITY.md              # Security features & guide
└── README.md                # This file
```

## 🔧 Configuration

**Development** (localhost only):
```bash
docker compose up -d
# Runs on http://127.0.0.1:8090
```

**Production** (public port 80):
```bash
docker compose -f docker-compose.prod.yaml up -d
# Runs on http://0.0.0.0:80
```

**With TLS/HTTPS**:
```bash
docker run -d -p 443:8443 \
  -v /path/to/certs:/certs:ro \
  ip-service \
  ./ip-service -tls -cert /certs/fullchain.pem -key /certs/privkey.pem -addr :8443
```

**Behind Nginx** (recommended for production):
```nginx
location / {
    proxy_pass http://127.0.0.1:8090;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    # See DOCUMENTATION.md for complete config
}
```

## 🛡️ Security

- ✅ Rate limiting (10 req/10sec per IP)
- ✅ CORS support for browser APIs
- ✅ Security headers (CSP, X-Frame-Options, etc.)
- ✅ Path traversal protection
- ✅ Non-root container user
- ✅ Resource limits (CPU/RAM)
- ✅ DNS timeout (prevents DoS)

See [SECURITY.md](./SECURITY.md) for complete security documentation.

## 🌍 API Endpoints

| Endpoint | Description |
|----------|-------------|
| `/` | HTML interface (or plain text for curl) |
| `/json` | JSON response |
| `/yaml` | YAML response |
| `/xml` | XML response |
| `/text` | Plain text IP only |
| `/headers` | View HTTP headers |

Example response:
```json
{
  "ip_address": "203.0.113.42",
  "hostname": "example.com",
  "isp": "Example ISP",
  "city": "San Francisco",
  "country": "United States",
  "country_code": "US"
}
```

## 🐛 Troubleshooting

**Service won't start?**
```bash
docker logs ip-service
```

**GeoIP not working?**
```bash
ls service/geoip/  # Should see ipcity.mmdb, iporg.mmdb
./setup.sh         # Re-run setup
```

**Port conflict?**
```bash
sudo netstat -tulpn | grep :8090
```

See [Troubleshooting](./DOCUMENTATION.md#troubleshooting) in full docs for more.

## 📝 License

MIT License - See [LICENSE](./LICENSE)

## Acknowledgments

Built upon the original [wtfismyip](https://codeberg.org/wtfismyip/wtfismyip) project.

---

**For complete documentation including all deployment scenarios, see [DOCUMENTATION.md](./DOCUMENTATION.md)**
