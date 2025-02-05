package smtp

// adapted from https://github.com/kirabou/parseMIMEemail.go

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
)

type MimeMail struct {
	From              string
	To                string
	Date              string
	Subject           string
	RawContent        []byte
	TextContent       []byte
	HtmlContent       []byte
	AttachedMimeParts map[string][]byte
	sanitizer         *bluemonday.Policy
}

func ParseMimeMessage(message []byte, sanitizer *bluemonday.Policy) (*MimeMail, error) {
	messageReader := bytes.NewReader(message)
	//  Parse the message to separate the Header and the Body with mail.ReadMessage()
	m, err := mail.ReadMessage(messageReader)
	if err != nil {
		return nil, err
	}
	mm := MimeMail{RawContent: message, AttachedMimeParts: make(map[string][]byte), sanitizer: sanitizer}

	// Record only the main headers of the message. The "From","To" and "Subject" headers
	// have to be decoded if they were encoded using RFC 2047 to allow non ASCII characters.
	// We use a mime.WordDecode for that.
	dec := new(mime.WordDecoder)
	mm.From, _ = dec.DecodeHeader(m.Header.Get("From"))
	mm.To, _ = dec.DecodeHeader(m.Header.Get("To"))
	mm.Subject, _ = dec.DecodeHeader(m.Header.Get("Subject"))

	mediaType, params, err := mime.ParseMediaType(m.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		return nil, errors.New("not multipart mime")
	}

	// Recursivey parse the MIME parts of the Body, starting with the first
	// level where the MIME parts are separated with params["boundary"].
	err = mm.parsePart(m.Body, params["boundary"], 1)
	return &mm, err
}

func (mm *MimeMail) SanitizedHtmlContent() template.HTML {
	if mm.HtmlContent == nil {
		return ""
	}
	return template.HTML(mm.sanitizer.Sanitize(string(mm.HtmlContent)))
}

// parsePart parses the MIME part from mime_data, each part being separated by
// boundary. If one of the part read is itself a multipart MIME part, the
// function calls itself to recursively parse all the parts. The parts read
// are decoded and written to separate files, named uppon their Content-Description
// (or boundary if no Content-Description available) with the appropriate
// file extension. Index is incremented at each recursive level and is used in
// building the filename under which the part is stored.
func (mm *MimeMail) parsePart(mime_data io.Reader, boundary string, index int) error {
	reader := multipart.NewReader(mime_data, boundary)
	if reader == nil {
		return nil
	}

	// Go through each of the MIME part of the message Body with NextPart(),
	// and read the content of the MIME part with ioutil.ReadAll()
	for {
		new_part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return fmt.Errorf("error going through the MIME parts: %v", err)
			}
		}
		mediaType, params, err := mime.ParseMediaType(new_part.Header.Get("Content-Type"))
		if err == nil && strings.HasPrefix(mediaType, "multipart/") {
			err = mm.parsePart(new_part, params["boundary"], index+1)
			if err != nil {
				return err
			}
		} else {
			filename, mediaType := mm.buildFileName(new_part, boundary, 1)
			data, err := mm.decodePart(new_part)
			if err != nil {
				return err
			}
			mm.AttachedMimeParts[filename] = data
			if strings.HasPrefix(mediaType, "text/plain") && mm.TextContent == nil {
				mm.TextContent = data
			}
			if strings.HasPrefix(mediaType, "text/html") && mm.HtmlContent == nil {
				mm.HtmlContent = data
			}
		}
	}

	return nil
}

// buildFileName builds a file name for a MIME part, using information extracted from
// the part itself, as well as a radix and an index given as parameters.
func (mm *MimeMail) buildFileName(part *multipart.Part, radix string, index int) (fileName, mediaType string) {
	if part.FileName() != "" {
		return part.FileName(), ""
	}
	contentTypeHeader := part.Header.Get("Content-Type")
	// some systems don't have a mimetype database; infer most common types
	switch {
	case strings.HasPrefix(contentTypeHeader, "text/plain"):
		return fmt.Sprintf("%s-%d%s", radix, index, ".txt"), contentTypeHeader
	case strings.HasPrefix(mediaType, "text/html"):
		return fmt.Sprintf("%s-%d%s", radix, index, ".html"), contentTypeHeader
	default:
		mediaType, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil {
			return uuid.NewString(), ""
		}
		mime_type, err := mime.ExtensionsByType(mediaType)
		if err != nil {
			return uuid.NewString(), ""
		}
		if len(mime_type) == 0 {
			return uuid.NewString(), ""
		}
		return fmt.Sprintf("%s-%d%s", radix, index, mime_type[0]), mediaType
	}
}

// decodePart decodes the data of MIME part
func (mm *MimeMail) decodePart(part *multipart.Part) ([]byte, error) {
	// Read the data for this MIME part
	part_data, err := io.ReadAll(part)
	if err != nil {
		return nil, err
	}

	content_transfer_encoding := strings.ToUpper(part.Header.Get("Content-Transfer-Encoding"))

	switch {
	case strings.Compare(content_transfer_encoding, "BASE64") == 0:
		decoded_content, err := base64.StdEncoding.DecodeString(string(part_data))
		if err != nil {
			return nil, err
		} else {
			return decoded_content, nil
		}
	case strings.Compare(content_transfer_encoding, "QUOTED-PRINTABLE") == 0:
		decoded_content, err := io.ReadAll(quotedprintable.NewReader(bytes.NewReader(part_data)))
		if err != nil {
			return nil, err
		} else {
			return decoded_content, nil
		}
	default:
		return part_data, nil
	}
}
