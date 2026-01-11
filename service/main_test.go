package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMainHandleXSS(t *testing.T) {
	// Setup static dir for tests
	staticDir = "./static"
	// Initialize templates
	templateHTML = parseTemplate("html.template")

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Inject malicious X-Forwarded-For
	maliciousIP := "<script>alert(1)</script>"
	req.Header.Set("X-Forwarded-For", maliciousIP)
	// Set User-Agent to something generic to ensure we get HTML
	req.Header.Set("User-Agent", "Mozilla/5.0")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mainHandle)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()

	// Check for unescaped script tag
	if strings.Contains(body, maliciousIP) {
		t.Errorf("handler returned unescaped XSS payload: %v", maliciousIP)
	}

	// Check for escaped script tag
	expected := "&lt;script&gt;alert(1)&lt;/script&gt;"
	if !strings.Contains(body, expected) {
		t.Errorf("handler did not return properly escaped payload. Expected content containing: %v", expected)
	}
}

func TestCleanHandleXSS(t *testing.T) {
	templateClean = parseTemplate("clean.template")

	req, err := http.NewRequest("GET", "/clean", nil)
	if err != nil {
		t.Fatal(err)
	}

	maliciousIP := "<script>alert(1)</script>"
	req.Header.Set("X-Forwarded-For", maliciousIP)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(cleanHandle)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()

	if strings.Contains(body, maliciousIP) {
		t.Errorf("handler returned unescaped XSS payload: %v", maliciousIP)
	}

	expected := "&lt;script&gt;alert(1)&lt;/script&gt;"
	if !strings.Contains(body, expected) {
		t.Errorf("handler did not return properly escaped payload. Expected content containing: %v", expected)
	}
}

func TestXMLHandleXSS(t *testing.T) {
	templateXML = parseTemplate("xml.template")

	req, err := http.NewRequest("GET", "/xml", nil)
	if err != nil {
		t.Fatal(err)
	}

	maliciousIP := "<script>alert(1)</script>"
	req.Header.Set("X-Forwarded-For", maliciousIP)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(xmlHandle)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()

	// For XML, html/template also escapes
	if strings.Contains(body, maliciousIP) {
		t.Errorf("handler returned unescaped XSS payload in XML: %v", maliciousIP)
	}

	expected := "&lt;script&gt;alert(1)&lt;/script&gt;"
	if !strings.Contains(body, expected) {
		t.Errorf("handler did not return properly escaped payload in XML. Expected content containing: %v", expected)
	}
}
