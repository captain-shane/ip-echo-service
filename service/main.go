package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/oschwald/geoip2-golang"
	"golang.org/x/time/rate"
	"gopkg.in/yaml.v2"
)

var (
	cityReader *geoip2.Reader
	orgReader  *geoip2.Reader

	templateHTML      *template.Template
	templateCleanJSON *template.Template
	templateXML       *template.Template
	templateClean     *template.Template

	staticDir  string
	geoIPDir   string
	listenAddr string
	tlsEnabled bool
	certFile   string
	keyFile    string

	// Rate limiter per IP
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

// visitor tracks rate limiter for each IP
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type GeoLocation struct {
	Org         string
	Details     string
	CountryCode string
	City        string
	Country     string
	State       string
}

type IPResponse struct {
	IPv6        bool
	Address     string
	Hostname    string
	Geo         string
	ISP         string
	CountryCode string
	City        string
	Country     string
}

func main() {
	flag.StringVar(&staticDir, "static", "./static", "Path to static files")
	flag.StringVar(&geoIPDir, "geoip", "./geoip", "Path to GeoIP databases")
	flag.StringVar(&listenAddr, "addr", ":8080", "Address to listen on")
	flag.BoolVar(&tlsEnabled, "tls", false, "Enable TLS")
	flag.StringVar(&certFile, "cert", "", "TLS Certificate file")
	flag.StringVar(&keyFile, "key", "", "TLS Key file")
	flag.Parse()

	// Optional GeoIP
	cityPath := filepath.Join(geoIPDir, "ipcity.mmdb")
	if _, err := os.Stat(cityPath); err == nil {
		cityReader, err = geoip2.Open(cityPath)
		if err != nil {
			log.Printf("Failed to open city GeoIP: %v", err)
		} else {
			defer cityReader.Close()
		}
	} else {
		log.Printf("City GeoIP database not found at %s. GeoIP features limited.", cityPath)
	}

	orgPath := filepath.Join(geoIPDir, "iporg.mmdb")
	if _, err := os.Stat(orgPath); err == nil {
		orgReader, err = geoip2.Open(orgPath)
		if err != nil {
			log.Printf("Failed to open org GeoIP: %v", err)
		} else {
			defer orgReader.Close()
		}
	} else {
		log.Printf("Org GeoIP database not found at %s. GeoIP features limited.", orgPath)
	}

	// Templates
	templateHTML = parseTemplate("html.template")
	templateCleanJSON = parseTemplate("cleanjson.template")
	templateClean = parseTemplate("clean.template")
	templateXML = parseTemplate("xml.template")

	// Clean up old visitors every 5 minutes
	go cleanupVisitors()

	r := mux.NewRouter()
	// Apply rate limiting and CORS to all routes
	r.Use(rateLimitMiddleware)
	r.Use(corsMiddleware)

	// Generic handlers
	r.HandleFunc("/", mainHandle).Methods("GET")
	r.HandleFunc("/json", jsonHandle)
	r.HandleFunc("/yaml", yamlHandle)
	r.HandleFunc("/xml", xmlHandle)
	r.HandleFunc("/text", textHandle)
	r.HandleFunc("/clean", cleanHandle)
	r.HandleFunc("/headers", headersHandle)

	// Static files - with path traversal protection
	r.PathPrefix("/static/").HandlerFunc(safeStaticHandler)
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		serveStaticFile(w, r, "favicon.ico")
	})
	r.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		serveStaticFile(w, r, "robots.txt")
	})

	log.Printf("Listening on %s", listenAddr)
	if tlsEnabled {
		if certFile == "" || keyFile == "" {
			log.Fatal("TLS enabled but cert or key file not specified")
		}
		log.Fatal(http.ListenAndServeTLS(listenAddr, certFile, keyFile, r))
	} else {
		log.Fatal(http.ListenAndServe(listenAddr, r))
	}
}

func parseTemplate(filename string) *template.Template {
	path := filepath.Join(staticDir, filename)
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		log.Printf("Warning: could not parse template %s: %v", path, err)
		// Return empty template to avoid nil panic, though handlers should handle it
		return template.New(filename)
	}
	return tmpl
}

func getAddress(r *http.Request) string {
	// Trust X-Forwarded-For if behind a proxy
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func getGeo(ipStr string) GeoLocation {
	var loc GeoLocation
	loc.Details = "Unknown"
	loc.CountryCode = "XX"

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return loc
	}

	if cityReader != nil {
		record, err := cityReader.City(ip)
		if err == nil {
			loc.City = record.City.Names["en"]
			loc.Country = record.Country.Names["en"]
			loc.CountryCode = record.Country.IsoCode
			if len(record.Subdivisions) > 0 {
				loc.State = record.Subdivisions[0].IsoCode
			}
			parts := []string{}
			if loc.City != "" {
				parts = append(parts, loc.City)
			}
			if loc.State != "" {
				parts = append(parts, loc.State)
			}
			if loc.Country != "" {
				parts = append(parts, loc.Country)
			}
			if len(parts) > 0 {
				loc.Details = strings.Join(parts, ", ")
			}
		}
	}

	if orgReader != nil {
		isp, err := orgReader.Enterprise(ip)
		if err == nil {
			loc.Org = isp.Traits.ISP
		}
	}

	return loc
}

func reverseDNS(ip string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	names, err := net.DefaultResolver.LookupAddr(ctx, ip)
	if err != nil || len(names) == 0 {
		return ip
	}
	return strings.TrimSuffix(names[0], ".")
}

func getResponse(r *http.Request) IPResponse {
	ip := getAddress(r)
	geo := getGeo(ip)
	hostname := reverseDNS(ip)
	isIPv6 := strings.Contains(ip, ":")

	return IPResponse{
		IPv6:        isIPv6,
		Address:     ip,
		Hostname:    hostname,
		Geo:         geo.Details,
		ISP:         geo.Org,
		CountryCode: geo.CountryCode,
		City:        geo.City,
		Country:     geo.Country,
	}
}

func mainHandle(w http.ResponseWriter, r *http.Request) {
	// Security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'unsafe-inline'")
	
	if strings.Contains(r.UserAgent(), "curl") || strings.Contains(r.UserAgent(), "HTTPie") {
		textHandle(w, r)
		return
	}
	resp := getResponse(r)
	if templateHTML != nil {
		if err := templateHTML.Execute(w, resp); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	} else {
		fmt.Fprintf(w, "IP: %s", resp.Address)
	}
}

func cleanHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	resp := getResponse(r)
	if templateClean != nil {
		if err := templateClean.Execute(w, resp); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func jsonHandle(w http.ResponseWriter, r *http.Request) {
	resp := getResponse(r)
	
	output := struct {
		IP          string `json:"ip_address"`
		Location    string `json:"location"`
		Hostname    string `json:"hostname"`
		ISP         string `json:"isp"`
		City        string `json:"city"`
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
	}{
		IP:          resp.Address,
		Location:    resp.Geo,
		Hostname:    resp.Hostname,
		ISP:         resp.ISP,
		City:        resp.City,
		Country:     resp.Country,
		CountryCode: resp.CountryCode,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func yamlHandle(w http.ResponseWriter, r *http.Request) {
	resp := getResponse(r)

	output := struct {
		IP          string `yaml:"ip_address"`
		Location    string `yaml:"location"`
		Hostname    string `yaml:"hostname"`
		ISP         string `yaml:"isp"`
		City        string `yaml:"city"`
		Country     string `yaml:"country"`
		CountryCode string `yaml:"country_code"`
	}{
		IP:          resp.Address,
		Location:    resp.Geo,
		Hostname:    resp.Hostname,
		ISP:         resp.ISP,
		City:        resp.City,
		Country:     resp.Country,
		CountryCode: resp.CountryCode,
	}

	w.Header().Set("Content-Type", "text/yaml")
	yaml.NewEncoder(w).Encode(output)
}

func xmlHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	resp := getResponse(r)
	if templateXML != nil {
		if err := templateXML.Execute(w, resp); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func textHandle(w http.ResponseWriter, r *http.Request) {
	ip := getAddress(r)
	fmt.Fprintf(w, "%s\n", ip)
}

func headersHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// Filter out sensitive internal headers
	sensitiveHeaders := map[string]bool{
		"X-Forwarded-For": true,
		"X-Real-Ip":       true,
		"Cookie":          true,
		"Authorization":   true,
	}
	for name, headers := range r.Header {
		if !sensitiveHeaders[name] {
			for _, h := range headers {
				fmt.Fprintf(w, "%v: %v\n", name, h)
			}
		}
	}
}

// safeStaticHandler prevents path traversal attacks
func safeStaticHandler(w http.ResponseWriter, r *http.Request) {
	// Clean the path and ensure it stays within staticDir
	requestPath := r.URL.Path
	if !strings.HasPrefix(requestPath, "/static/") {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	
	relPath := strings.TrimPrefix(requestPath, "/static/")
	relPath = path.Clean("/" + relPath)
	if strings.Contains(relPath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	
	fullPath := filepath.Join(staticDir, relPath)
	http.ServeFile(w, r, fullPath)
}

func serveStaticFile(w http.ResponseWriter, r *http.Request, filename string) {
	// Ensure only the exact file is served
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, filepath.Join(staticDir, filename))
}


// Rate limiting: 10 requests per 10 seconds per IP
func getVisitor(ip string) *rate.Limiter {
mu.Lock()
defer mu.Unlock()

v, exists := visitors[ip]
if !exists {
// Allow 10 requests with burst of 10 (1 request per second average)
limiter := rate.NewLimiter(1, 10)
visitors[ip] = &visitor{limiter, time.Now()}
return limiter
}

v.lastSeen = time.Now()
return v.limiter
}

// Cleanup old visitors to prevent memory leak
func cleanupVisitors() {
for {
time.Sleep(5 * time.Minute)
mu.Lock()
for ip, v := range visitors {
if time.Since(v.lastSeen) > 10*time.Minute {
delete(visitors, ip)
}
}
mu.Unlock()
}
}

// Rate limiting middleware
func rateLimitMiddleware(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
ip := getAddress(r)
limiter := getVisitor(ip)

if !limiter.Allow() {
http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
return
}

next.ServeHTTP(w, r)
})
}

// CORS middleware - Allow public API access from browsers
func corsMiddleware(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Access-Control-Allow-Origin", "*")
w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

// Handle preflight requests
if r.Method == "OPTIONS" {
w.WriteHeader(http.StatusOK)
return
}

next.ServeHTTP(w, r)
})
}
