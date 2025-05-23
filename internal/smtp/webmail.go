package smtp

import (
	"encoding/base64"
	"html/template"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/xsrftoken"
)

type Webmail struct {
	xsrfSecret string
	blobClient BlobClient
	sanitizer  *bluemonday.Policy
	noTls      bool
}

func NewWebMailer(xsrfSecret string, blobClient BlobClient, noTls bool) *Webmail {
	return &Webmail{
		xsrfSecret: xsrfSecret,
		blobClient: blobClient,
		sanitizer:  bluemonday.UGCPolicy(),
		noTls:      noTls,
	}
}

func (wm *Webmail) ListenAndServeWebmail() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		mails := []string{}
		if wm.validSession(req) {
			var err error
			mails, err = wm.blobClient.ListMail()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			slices.Reverse(mails)
		}
		wm.page(wm.indexTmpl(), struct{ Mails []string }{Mails: mails})(w, req)
	})
	http.HandleFunc("/login", wm.loginFormHandler)
	http.HandleFunc("/mail/", wm.showMailHandler)

	log.Println("Starting webmail server at", "0.0.0.0:8443")
	if wm.noTls {
		log.Fatal(http.ListenAndServe("0.0.0.0:8443", nil))
	} else {
		s := &http.Server{
			Addr:      "0.0.0.0:8443",
			TLSConfig: NewSSLmanager(wm.blobClient).TLSConfig(),
		}
		log.Fatal(s.ListenAndServeTLS("", ""))
	}
}

// checks session, sets cors xsrf and other headers, renders page
func (wm *Webmail) page(content string, data any) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		styleNonce, formToken := wm.setPageHeadersAndCookies(w)
		pageTmpl := template.Must(template.New("rendered").Parse(wm.header() + content + wm.footer()))
		err := pageTmpl.Execute(w, struct {
			XsrfToken  string
			StyleNonce string
			LoggedIn   bool
			Data       any
		}{formToken, styleNonce, wm.validSession(req), data})
		if err != nil {
			log.Printf("failed render: %v", err)
		}

		log.Printf("path=%q ip=%q", req.URL.Path, req.RemoteAddr)
	}
}

func (wm *Webmail) setPageHeadersAndCookies(w http.ResponseWriter) (styleNonce, formToken string) {
	newStyleNonce := wm.setSecurityHeaders(w)
	newFormToken := xsrftoken.Generate(wm.xsrfSecret, "", "")
	http.SetCookie(w, &http.Cookie{
		Name:     "xsrftoken",
		Value:    newFormToken,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	return newStyleNonce, newFormToken
}

// checks credentials in login POST and sets session
func (wm *Webmail) loginFormHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.NotFound(w, req)
		return
	}
	if !wm.validXsrf(req) {
		http.Redirect(w, req, "/", http.StatusForbidden)
		return
	}
	if wm.validCredentials(req.FormValue("user"), req.FormValue("password")) {
		sessionCookie := xsrftoken.Generate(wm.xsrfSecret, req.FormValue("user"), "session")
		http.SetCookie(w, &http.Cookie{
			Name:     "user",
			Value:    req.FormValue("user"),
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sessionCookie,
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		http.Redirect(w, req, "/", http.StatusFound)
		return
	} else {
		http.Redirect(w, req, "/", http.StatusForbidden)
		return
	}
}

// Shows a mail
func (wm *Webmail) showMailHandler(w http.ResponseWriter, req *http.Request) {
	defer log.Printf("path=%q ip=%q", req.URL.Path, req.RemoteAddr)
	if !wm.validSession(req) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	b, err := wm.blobClient.Get(strings.TrimPrefix(req.URL.EscapedPath(), "/mail/"))
	if err != nil {
		log.Printf("showMailHandler %v: %v", req.URL.EscapedPath(), err)
		return
	}

	parsedMimeMessage, err := ParseMimeMessage(b, wm.sanitizer)
	if err != nil {
		wm.page(`<div>{{.Data.Body}}</div>`, struct{ Body string }{Body: "Error:" + err.Error() + "\n" + string(b)})(w, req)
		return
	} else {
		wm.page(wm.showMailTmpl(), parsedMimeMessage)(w, req)
	}
}

func (wm *Webmail) setSecurityHeaders(w http.ResponseWriter) (styleNonce string) {
	styleNonce = base64.StdEncoding.EncodeToString([]byte(xsrftoken.Generate(wm.xsrfSecret, "", "style")))
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'nonce-"+styleNonce+"'; form-action 'self'")
	return styleNonce
}

func (wm *Webmail) validSession(req *http.Request) bool {
	if session, err := req.Cookie("session"); err == nil {
		if user, err := req.Cookie("user"); err == nil {
			return xsrftoken.ValidFor(session.Value, wm.xsrfSecret, user.Value, "session", xsrftoken.Timeout)
		}
	}
	return false
}

func (wm *Webmail) validXsrf(req *http.Request) bool {
	xsrftokenCookie, err := req.Cookie("xsrftoken")
	if err != nil || req.FormValue("xsrftoken") == "" {
		return false
	}
	if req.FormValue("xsrftoken") != xsrftokenCookie.Value {
		return false
	}

	return xsrftoken.ValidFor(req.FormValue("xsrftoken"), wm.xsrfSecret, "", "", xsrftoken.Timeout)
}

// auth handler, comparing to a bcrypt in blob storage
func (wm *Webmail) validCredentials(user string, password string) bool {
	hsh, err := wm.blobClient.Get("bcrypt/" + user)
	if err != nil {
		return false
	}
	return (bcrypt.CompareHashAndPassword(hsh, []byte(password)) == nil)
}

// HTML
func (wm *Webmail) indexTmpl() string {
	return `<div>
		<div class="flex-container">
		<header><h2>Webmail</h2></header>
			{{if .LoggedIn}}
				<ul>
				{{ range .Data.Mails}}
					<li><a href="/mail/{{.}}">{{.}}</a></li>
				{{ end }}
				</ul>
			{{else}}
				<form method="POST" action="/login">
					<label>User:</label><br />
					<input type="text" name="user"><br />
					<label>Password:</label><br />
					<input type="password" name="password"><br />
					<input type="hidden" name="xsrftoken" value="{{ .XsrfToken }}">
					<input type="submit">
				</form>
			{{end}}
		</div>
	</div>`
}

func (wm *Webmail) showMailTmpl() string {
	return `<div>
	    <h3>Message</h3>
		<ul>
		  <li><strong>From</strong>: {{ .Data.From }}</li>
		  <li><strong>To</strong>: {{ .Data.To }}</li>
		  <li><strong>Date</strong>: {{ .Data.Date }}</li>
		  <li><strong>Subject</strong>: {{ .Data.Subject }}</li>
		</ul>
		{{if eq .Data.SanitizedHtmlContent ""}}
			{{.RawContent}}
		{{else}}
			{{.Data.SanitizedHtmlContent}}
		{{end}}
		<div>
		<h3>Attached mime parts</h3>
		<ul>
		{{ range $key, $value := .Data.AttachedMimeParts }}
			<li><strong>{{ $key }}</strong>: {{ len $value }}</li>
		{{ end }}
		 </ul>
		</div>
	</div>`
}

func (wm *Webmail) header() string {
	return `<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>sifio webmail</title>
		<style nonce="{{ .StyleNonce }}">
			.flex-container {
				display: -webkit-flex;
				display: flex;
				-webkit-flex-flow: row wrap;
				flex-flow: row wrap;
				text-align: left;
			}
			.flex-container > * {
				padding: 15px;
				-webkit-flex: 1 100%;
				flex: 1 100%;
			}
			.article {
				text-align: left;
				background: #3A6EA5;
				color: white;
			}
			header {background: #dd9c37; color: white;}
			.nav {background: azure;}
			.nav ul {
				list-style-type: none;
				padding: 0;
			}
			.nav ul a {
			}
			@media all and (min-width: 768px) {
				.nav {text-align:left;-webkit-flex: 1 auto;flex:1 auto;-webkit-order:1;order:1;}
				.article {-webkit-flex:5 0px;flex:5 0px;-webkit-order:2;order:2;}
				footer {-webkit-order:3;order:3;}
			}
		</style>
	</head>
	<body>`
}

func (wm *Webmail) footer() string {
	return `</body></html>`
}
