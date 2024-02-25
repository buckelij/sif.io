package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/xsrftoken"
)

func TestValidXsrf(t *testing.T) {
	testBlobClient := &TestBlobClient{}
	wm := Webmail{XsrfSecret: "123", BlobClient: testBlobClient}

	formToken := xsrftoken.Generate(wm.XsrfSecret, "", "")
	data := url.Values{}
	data.Set("xsrftoken", formToken)
	r := httptest.NewRequest("POST", "/login", strings.NewReader(data.Encode()))
	r.AddCookie(&http.Cookie{Name: "xsrftoken", Value: formToken})
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.ParseForm()
	if !wm.validXsrf(r) {
		t.Fatal()
	}

}

func TestSetSecurityHeaders(t *testing.T) {
	testBlobClient := &TestBlobClient{}
	wm := Webmail{XsrfSecret: "123", BlobClient: testBlobClient}

	rr := httptest.NewRecorder()
	styleNonce := wm.setSecurityHeaders(rr)
	if rr.Result().Header["Content-Security-Policy"][0] != "default-src 'none'; style-src 'nonce-"+styleNonce+"'; form-action 'self'" {
		t.Fatal("incorrect CSP" + rr.Result().Header["Content-Security-Policy"][0])
	}
}

func TestValidSession(t *testing.T) {
	testBlobClient := &TestBlobClient{}
	wm := Webmail{XsrfSecret: "123", BlobClient: testBlobClient}

	formToken := xsrftoken.Generate(wm.XsrfSecret, "buckelij", "session")
	data := url.Values{}
	data.Set("xsrftoken", formToken)
	r := httptest.NewRequest("POST", "/login", strings.NewReader(data.Encode()))
	r.AddCookie(&http.Cookie{Name: "session", Value: formToken})
	r.AddCookie(&http.Cookie{Name: "user", Value: "buckelij"})
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.ParseForm()
	if !wm.validSession(r) {
		t.Fatal()
	}
}

func TestValidCredentials(t *testing.T) {
	gets := make([][]byte, 0)
	hsh, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	gets = append(gets, hsh)
	gets = append(gets, hsh)
	testBlobClient := &TestBlobClient{gets: gets}
	wm := Webmail{XsrfSecret: "123", BlobClient: testBlobClient}
	if !wm.validCredentials("testuser", "testpass") {
		t.Fatal()
	}

	if wm.validCredentials("testuser", "badpass") {
		t.Fatal()
	}
}
