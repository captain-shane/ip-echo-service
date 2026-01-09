# Security Features

This document outlines the security features and best practices implemented in the IP Echo Service.

## Built-in Security Features

### Rate Limiting
**Protection against abuse and DoS attacks**

- **Algorithm**: Token bucket (1 request per second, burst of 10)
- **Scope**: Per source IP address
- **Response**: HTTP 429 (Too Many Requests) when exceeded
- **Cleanup**: Automatic visitor cleanup every 5 minutes to prevent memory leaks

**How it works**: Each IP address gets 10 "tokens". Each request consumes 1 token. Tokens refill at 1 per second. When empty, requests are denied until tokens refill.

### CORS (Cross-Origin Resource Sharing)
**Enables secure browser-based API access**

- **Origin**: `Access-Control-Allow-Origin: *` (allows all domains)
- **Methods**: GET, OPTIONS
- **Headers**: Content-Type allowed
- **Preflight**: Handles browser OPTIONS requests properly

**Use case**: Web applications can call this API directly from JavaScript without CORS errors.

### Security Headers
**Protection against common web attacks**

All HTML responses include:
- `X-Content-Type-Options: nosniff` - Prevents MIME sniffing attacks
- `X-Frame-Options: DENY` - Prevents clickjacking/iframe embedding
- `Content-Security-Policy: default-src 'self'; style-src 'unsafe-inline'` - Restricts resource loading

### Input Validation & Sanitization

**Path Traversal Protection**:
- Static file paths sanitized with `path.Clean()`
- Detection and blocking of `..` sequences
- Filename validation (no path components allowed)

**Header Filtering**:
- Sensitive headers filtered from `/headers` endpoint
- Blocks: `X-Forwarded-For`, `Cookie`, `Authorization`

**IP Parsing**:
- All IPs validated through `net.ParseIP()` before processing
- Prevents injection attacks

### DNS Protection

**Timeout on Reverse DNS Lookups**:
- 2-second context timeout on all DNS operations
- Prevents service hanging on slow/malicious DNS servers
- Graceful fallback to IP address if lookup fails

### Container Security

**Non-root Execution**:
- Runs as unprivileged user `appuser`
- No root access within container
- Minimal privileges

**Minimal Base Image**:
- Alpine Linux (5MB base)
- Reduces attack surface
- Only essential packages installed

**Multi-stage Build**:
- Source code not included in final image
- Only compiled binary and assets
- Smaller image size

**Resource Limits**:
- Development: 0.5 CPU cores, 256MB RAM
- Production: 1.0 CPU cores, 512MB RAM
- Prevents resource exhaustion attacks

### Network Security

**Port Binding**:
- Development: `127.0.0.1:8090` (localhost only)
- Production: Configurable (recommended behind firewall/proxy)

**TLS Support**:
- Optional built-in TLS via command-line flags
- Supports custom certificates
- Recommended: Use reverse proxy for SSL termination

## Security Best Practices

### Recommended Deployment Architecture

```
Internet → Firewall → Nginx (SSL) → Docker Container
```

**Why?**
- Nginx handles SSL/TLS termination
- Additional layer of rate limiting
- Load balancing capability
- Better logging and monitoring

### Production Hardening Checklist

**Before deploying to production**:

- [ ] Deploy behind reverse proxy (Nginx/Traefik)
- [ ] Enable HTTPS with valid SSL certificate
- [ ] Configure firewall (UFW, iptables, or cloud security groups)
- [ ] Set up log monitoring/aggregation
- [ ] Enable auto-updates for Docker images
- [ ] Schedule regular GeoIP database updates
- [ ] Configure backup for SSL certificates
- [ ] Review resource limits based on traffic
- [ ] Set up health monitoring/alerting
- [ ] Document incident response procedures

### Secure Configuration Examples

**Nginx Reverse Proxy (Recommended)**:
```nginx
server {
    listen 443 ssl http2;
    server_name yourdomain.com;
    
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    
    # Additional rate limiting (beyond built-in)
    limit_req_zone $binary_remote_addr zone=api:10m rate=20r/s;
    limit_req zone=api burst=30 nodelay;
    
    location / {
        proxy_pass http://127.0.0.1:8090;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

**Firewall Rules (UFW)**:
```bash
# Allow only necessary ports
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

**Docker Security**:
```yaml
# In docker-compose.yaml
services:
  ip-service:
    # ... other config ...
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp
```

## Vulnerability Assessment

### Attack Surface Analysis

| Vector | Risk Level | Mitigation |
|--------|-----------|------------|
| DoS via slow DNS | **LOW** | 2-second timeout implemented |
| Rate limit bypass | **LOW** | Per-IP token bucket algorithm |
| Path traversal | **NONE** | Path sanitization active |
| XSS attacks | **LOW** | Go templates auto-escape HTML |
| Header injection | **LOW** | Header filtering enabled |
| Memory exhaustion | **LOW** | Resource limits + automatic cleanup |
| Container escape | **VERY LOW** | Non-root user, minimal privileges |

### OWASP Top 10 Compliance

| Category | Status | Notes |
|----------|--------|-------|
| A01: Broken Access Control | ✅ | N/A (public service, no access control needed) |
| A02: Cryptographic Failures | ✅ | Optional TLS support, no sensitive data stored |
| A03: Injection | ✅ | No SQL/command injection vectors |
| A04: Insecure Design | ✅ | Secure by design (rate limiting, timeouts) |
| A05: Security Misconfiguration | ✅ | Security headers, non-root user, minimal image |
| A06: Vulnerable Components | ✅ | Minimal dependencies, regular updates recommended |
| A07: Authentication Failures | ✅ | N/A (no authentication) |
| A08: Software/Data Integrity | ✅ | Multi-stage Docker build, reproducible |
| A09: Logging Failures | ⚠️ | Basic logging (consider structured logging for production) |
| A10: SSRF | ✅ | No outbound requests based on user input |

## Monitoring & Incident Response

### What to Monitor

1. **Rate Limit Hits**: High 429 responses may indicate abuse
2. **Error Rates**: Spike in 500 errors needs investigation
3. **DNS Timeouts**: Many timeouts could indicate DNS issues
4. **Resource Usage**: CPU/RAM approaching limits
5. **Container Health**: Failed health checks

### Log Locations

```bash
# Application logs
docker logs ip-service

# Nginx logs (if using reverse proxy)
/var/log/nginx/access.log
/var/log/nginx/error.log
```

### Security Incident Response

**If you suspect a security issue**:

1. Check logs for suspicious patterns
2. Review rate limit hits and sources
3. Verify container hasn't been compromised: `docker exec ip-service ps aux`
4. Update immediately if vulnerability found
5. Document and report

## Updates & Maintenance

### Regular Security Tasks

**Monthly**:
- Update GeoIP databases (run `./setup.sh`)
- Review access logs for anomalies
- Check for Docker image updates

**Quarterly**:
- Review and update SSL certificates (if not auto-renewed)
- Update base Docker images: `docker compose pull && docker compose up -d`
- Review rate limiting effectiveness

**Annually**:
- Security audit of configuration
- Review and update dependencies: Check for Go security advisories
- Update documentation

### Updating Dependencies

```bash
# Update Go dependencies
cd service/
go get -u ./...
go mod tidy

# Rebuild and deploy
cd ..
docker compose build --no-cache
docker compose up -d
```

## Responsible Disclosure

If you discover a security vulnerability:

1. **Do not** open a public GitHub issue
2. Email security concerns privately (provide contact in your GitHub repo)
3. Include: Description, steps to reproduce, impact assessment
4. Allow reasonable time for fix before public disclosure

## Security Audit Summary

**Overall Security Rating**: ✅ **Production Ready**

**Strengths**:
- Multiple layers of protection (defense in depth)
- Minimal attack surface
- Industry best practices followed
- Resource limits prevent DoS
- Security headers prevent common attacks

**Recommendations**:
- Deploy behind HTTPS reverse proxy in production
- Enable structured logging for better audit trail
- Consider adding request authentication for private deployments
- Implement log aggregation for larger deployments

---

**Last Updated**: 2026-01-09  
**Recommended Review Frequency**: Every 6 months or after major changes
