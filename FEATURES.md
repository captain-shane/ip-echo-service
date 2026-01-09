# Rate Limiting & CORS - Feature Documentation

## ✅ Successfully Implemented

### Rate Limiting
**Configuration**: 10 requests per 10 seconds per IP address

**How it works**:
- Each IP address gets a "token bucket" with 10 tokens
- Tokens refill at 1 per second
- When the bucket is empty, requests get HTTP 429 (Too Many Requests)
- Old visitor records are cleaned up every 5 minutes to prevent memory leaks

**Testing Results**:
```
Requests 1-10: Status 200 ✅
Requests 11-12: Status 429 ✅ (Rate limited as expected)
```

### CORS (Cross-Origin Resource Sharing)
**What is CORS?**
CORS is a browser security feature that controls whether JavaScript code from one website can request data from another domain.

**Example**: If a developer builds a web app at `example.com` and wants to call your API at `yourdomain.com/json` from their JavaScript code, the browser will block it by default **unless** your server sends CORS headers saying "this is allowed."

**Our Configuration**:
- Allows requests from ANY origin (`Access-Control-Allow-Origin: *`)
- Allows GET and OPTIONS methods
- Allows Content-Type header
- Handles browser "preflight" OPTIONS requests

This means developers can now use your IP service directly from browser-based apps!

## Usage Examples

### From curl/terminal:
```bash
curl http://yourdomain.com/json
```

### From JavaScript in a browser:
```javascript
fetch('http://yourdomain.com/json')
  .then(response => response.json())
  .then(data => console.log('My IP:', data.ip_address));
```

This will now work without CORS errors!
