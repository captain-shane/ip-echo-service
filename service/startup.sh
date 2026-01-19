#!/bin/sh
set -e

# GeoIP database download script for Cloud Run
# Downloads MaxMind GeoLite2 databases at container startup

GEOIP_DIR="${GEOIP_DIR:-./geoip}"
CITY_DB="ipcity.mmdb"
ASN_DB="iporg.mmdb"

# Create geoip directory if needed
mkdir -p "$GEOIP_DIR"

# Download databases if credentials are provided and DBs don't exist
if [ -n "$MAXMIND_ACCOUNT_ID" ] && [ -n "$MAXMIND_LICENSE_KEY" ]; then
    echo "[startup] MaxMind credentials found, checking for GeoIP databases..."
    
    # Download GeoLite2-City database
    if [ ! -f "$GEOIP_DIR/$CITY_DB" ]; then
        echo "[startup] Downloading GeoLite2-City database..."
        if curl -L -f -o "$GEOIP_DIR/city.tar.gz" \
            -u "${MAXMIND_ACCOUNT_ID}:${MAXMIND_LICENSE_KEY}" \
            'https://download.maxmind.com/geoip/databases/GeoLite2-City/download?suffix=tar.gz'; then
            tar -xzf "$GEOIP_DIR/city.tar.gz" -C "$GEOIP_DIR"
            find "$GEOIP_DIR" -name "*.mmdb" -path "*City*" -exec mv {} "$GEOIP_DIR/$CITY_DB" \;
            rm -rf "$GEOIP_DIR/city.tar.gz" "$GEOIP_DIR"/GeoLite2-City_*
            echo "[startup] Successfully installed $CITY_DB"
        else
            echo "[startup] WARNING: Failed to download GeoLite2-City. GeoIP features limited."
        fi
    else
        echo "[startup] $CITY_DB already exists, skipping download."
    fi
    
    # Download GeoLite2-ASN database (for ISP/Org info)
    if [ ! -f "$GEOIP_DIR/$ASN_DB" ]; then
        echo "[startup] Downloading GeoLite2-ASN database..."
        if curl -L -f -o "$GEOIP_DIR/asn.tar.gz" \
            -u "${MAXMIND_ACCOUNT_ID}:${MAXMIND_LICENSE_KEY}" \
            'https://download.maxmind.com/geoip/databases/GeoLite2-ASN/download?suffix=tar.gz'; then
            tar -xzf "$GEOIP_DIR/asn.tar.gz" -C "$GEOIP_DIR"
            find "$GEOIP_DIR" -name "*.mmdb" -path "*ASN*" -exec mv {} "$GEOIP_DIR/$ASN_DB" \;
            rm -rf "$GEOIP_DIR/asn.tar.gz" "$GEOIP_DIR"/GeoLite2-ASN_*
            echo "[startup] Successfully installed $ASN_DB"
        else
            echo "[startup] WARNING: Failed to download GeoLite2-ASN. ISP info unavailable."
        fi
    else
        echo "[startup] $ASN_DB already exists, skipping download."
    fi
else
    echo "[startup] No MaxMind credentials provided. GeoIP features will be limited."
    echo "[startup] Set MAXMIND_ACCOUNT_ID and MAXMIND_LICENSE_KEY to enable GeoIP."
fi

# Start the application
# Cloud Run sets PORT environment variable
PORT="${PORT:-8080}"
echo "[startup] Starting ip-service on port $PORT..."
exec ./ip-service -addr ":$PORT" -static ./static -geoip "$GEOIP_DIR"
