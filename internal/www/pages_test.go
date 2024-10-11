package www

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

var sampleHtml = `<!DOCTYPE html>
    <html>
    <head>
    	<meta charset="UTF-8">
    	<meta name="viewport" content="width=device-width, initial-scale=1">
    	<title>hello world</title>
    </head>
    <body>hello world</body>
    </html>`

func TestIndex404(t *testing.T) {
	pageHandler := Index(sampleHtml)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	pageHandler(w, req)

	resp := w.Result()
	// body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 404 {
		t.Fail()
	}
}
func TestIndex(t *testing.T) {
	pageHandler := Index(sampleHtml)

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	w := httptest.NewRecorder()
	pageHandler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Error("non 200 response")
	}

	if !strings.Contains(string(body), "hello world") {
		t.Error("unexpected response body")
	}
}

func TestPage(t *testing.T) {
	pageHandler := Page(sampleHtml)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	pageHandler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Error("non 200 response")
	}

	if !strings.Contains(string(body), "hello world") {
		t.Error("unexpected response body")
	}
}
