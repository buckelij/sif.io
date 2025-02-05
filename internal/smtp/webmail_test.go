package smtp

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/xsrftoken"
)

type TestBlobClient struct {
	wg       sync.WaitGroup // create a wait group, this will allow you to block later
	uploaded []string
	gets     [][]byte // stub values to be returned
}

func (c *TestBlobClient) Put(oid string, data []byte) error {
	c.uploaded = append(c.uploaded, oid)
	c.wg.Done()
	return nil
}

func (c *TestBlobClient) Get(oid string) ([]byte, error) {
	v := c.gets[0]
	c.gets = c.gets[1:]
	return v, nil
}

func (c *TestBlobClient) ListMail() ([]string, error) {
	return []string{}, nil
}

func TestValidXsrf(t *testing.T) {
	testBlobClient := &TestBlobClient{}
	wm := NewWebMailer("123", testBlobClient)

	formToken := xsrftoken.Generate(wm.xsrfSecret, "", "")
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
	wm := NewWebMailer("123", testBlobClient)

	rr := httptest.NewRecorder()
	styleNonce := wm.setSecurityHeaders(rr)
	if rr.Result().Header["Content-Security-Policy"][0] != "default-src 'none'; style-src 'nonce-"+styleNonce+"'; form-action 'self'" {
		t.Fatal("incorrect CSP" + rr.Result().Header["Content-Security-Policy"][0])
	}
}

func TestValidSession(t *testing.T) {
	testBlobClient := &TestBlobClient{}
	wm := NewWebMailer("123", testBlobClient)

	formToken := xsrftoken.Generate(wm.xsrfSecret, "buckelij", "session")
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
	wm := NewWebMailer("123", testBlobClient)
	if !wm.validCredentials("testuser", "testpass") {
		t.Fatal()
	}

	if wm.validCredentials("testuser", "badpass") {
		t.Fatal()
	}
}
