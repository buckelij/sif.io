package smtp

import (
	"testing"

	"github.com/microcosm-cc/bluemonday"
)

var simpleEmail = []byte(`Subject: test html mail
From: sender@example.com
To: recipient@example.com
Content-Type: multipart/alternative; boundary=bcaec520ea5d6918e204a8cea3b4

--bcaec520ea5d6918e204a8cea3b4
Content-Type: text/plain; charset=ISO-8859-1

*hi!*

--bcaec520ea5d6918e204a8cea3b4
Content-Type: text/html; charset=ISO-8859-1
Content-Transfer-Encoding: quoted-printable

<p><b>hi!</b></p>

--bcaec520ea5d6918e204a8cea3b4--`)

func TestParseMimeMessage(t *testing.T) {

	mm, err := ParseMimeMessage(simpleEmail, bluemonday.UGCPolicy())
	if err != nil {
		t.Error(err)
	}
	if mm.From != "sender@example.com" {
		t.Errorf("invalid from %s", mm.From)
	}
	if mm.To != "recipient@example.com" {
		t.Errorf("invalid from %s", mm.From)
	}
	if string(mm.TextContent) != "*hi!*\n" {
		t.Errorf("Unexpected TextContent: wanted '*hi!*' got '%v'", string(mm.TextContent))
	}
	if string(mm.HtmlContent) != "<p><b>hi!</b></p>\n" {
		t.Errorf("Unexpected TextContent: wanted '<p><b>hi!</b></p>' got '%v'", string(mm.HtmlContent))
	}
}
