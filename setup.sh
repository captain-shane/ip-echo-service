#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

GEOIP_DIR="./service/geoip"
CITY_DB="ipcity.mmdb"
ORG_DB="iporg.mmdb"

echo -e "${GREEN}=== IP Service Setup ===${NC}"

# Check for GeoIP directory
if [ ! -d "$GEOIP_DIR" ]; then
    echo -e "Creating GeoIP directory..."
    mkdir -p "$GEOIP_DIR"
fi

# Function to download DB
download_db() {
    local edition_id=$1
    local output_file=$2
    local license_key=$3

    echo -e "Downloading $edition_id..."
    url="https://download.maxmind.com/app/geoip_download?edition_id=${edition_id}&license_key=${license_key}&suffix=tar.gz"
    
    # Download to temp file
    if curl -L -f -o "temp.tar.gz" "$url"; then
        echo -e "Extracting..."
        tar -xzf temp.tar.gz
        # Find the mmdb file in the extracted folder (ignoring the folder name which changes with date)
        find . -name "*.mmdb" -exec mv {} "$GEOIP_DIR/$output_file" \;
        rm -rf temp.tar.gz GeoLite2*
        echo -e "${GREEN}Successfully installed $output_file${NC}"
        return 0
    else
        echo -e "${RED}Failed to download $edition_id. Check your License Key.${NC}"
        rm -f temp.tar.gz
        return 1
    fi
}

# Check if DBs exist
if [ -f "$GEOIP_DIR/$CITY_DB" ] && [ -f "$GEOIP_DIR/$ORG_DB" ]; then
    echo -e "${GREEN}GeoIP databases found.${NC}"
else
    echo -e "${YELLOW}GeoIP databases are missing.${NC}"
    echo -e "You need a free license key from MaxMind to download them automatically."
    echo -e "Get one here: https://www.maxmind.com/en/geolite2/signup"
    echo -n "Enter your MaxMind License Key (or press Enter to skip): "
    read -r LICENSE_KEY

    if [ -n "$LICENSE_KEY" ]; then
        download_db "GeoLite2-City" "$CITY_DB" "$LICENSE_KEY"
        download_db "GeoLite2-ASN" "$ORG_DB" "$LICENSE_KEY"
    else
        echo -e "${YELLOW}Skipping download.${NC}"
        echo -e "Please manually place '$CITY_DB' and '$ORG_DB' in '$GEOIP_DIR' to enable GeoIP features."
    fi
fi

echo -e "\n${GREEN}=== Ready to Deploy ===${NC}"
echo -e "1. Run locally: docker compose up -d --build"
echo -e "2. Run production: docker compose -f docker-compose.prod.yaml up -d --build"

echo -n "Do you want to start the Local server now? (y/n): "
read -r START_CONFIRM

if [[ "$START_CONFIRM" =~ ^[Yy]$ ]]; then
    docker compose up -d --build
    echo -e "${GREEN}Service started!${NC}"
    echo -e "Test with: curl localhost:8090"
fi
