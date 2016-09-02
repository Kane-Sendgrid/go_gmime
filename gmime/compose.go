package gmime

/*
#cgo pkg-config: gmime-2.6
#include <stdlib.h>
#include <stdarg.h>
#include <string.h>
#include <gmime/gmime.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

var (
	cStringEmpty       = C.CString("")
	cStringAlternative = C.CString("alternative")
	cStringMixed       = C.CString("mixed")
	cStringRelated     = C.CString("related")
	cStringCharset     = C.CString("charset")
	cStringCharsetUTF8 = C.CString("utf-8")

	cStringText  = C.CString("text")
	cStringPlain = C.CString("plain")
	cStringHTML  = C.CString("html")

	cStringContentID    = C.CString("Content-Id")
	cStringApplication  = C.CString("application")
	cStringOctetStream  = C.CString("octet-stream")
	cStringHeaderFormat = C.CString("%s: %s\n")
)

var (
	ErrNoContent = errors.New("No content (text or html)")
)

type EmailAttachment struct {
	FileName    string
	MimeType    string
	ContentID   string
	Disposition string
	Content     []byte
}

type Message struct {
	text     string
	html     string
	embeds   []*EmailAttachment
	attaches []*EmailAttachment
	headers  []*EmailHeader
}

type EmailHeader struct {
	Name  string
	Value string
	Raw   bool
}

type EncodedHeader struct {
	Name  string
	Value string
}

func NewMessage() *Message {
	return &Message{}
}

func (m *Message) SetText(body string) {
	m.text = body
}

func (m *Message) SetHtml(body string) {
	m.html = body
}

func (m *Message) Embed(a *EmailAttachment) {
	m.embeds = append(m.embeds, a)
}

func (m *Message) Attach(a *EmailAttachment) {
	m.attaches = append(m.attaches, a)
}

func (m *Message) AppendHeader(h *EmailHeader) {
	m.headers = append(m.headers, h)
}

func (m *Message) PrependHeader(h *EmailHeader) {
	m.headers = append([]*EmailHeader{h}, m.headers...)
}

func (m *Message) EncodedHeaders() []*EmailHeader {
	message := C.g_mime_message_new(C.TRUE)
	defer C.g_object_unref(message)
	injectHeaders(anyToGMimeObject(unsafe.Pointer(message)), m.headers)
	return encodedHeadersFromGmime(anyToGMimeObject(unsafe.Pointer(message)))
}

func (m *Message) gmimize() error {
	// - mixed
	//     - related
	//         - alternative
	//             - text/plain
	//             - text/html
	//         - embedded image 1
	//         - embedded image 2
	//     - Attachment 1
	//     - Attachment 2
	var contentPart *C.GMimeObject
	contentPart, err := textHTMLPart(m.text, m.html) // need unref
	if err != nil {
		return err
	}
	defer C.g_object_unref(contentPart) // unref

	if len(m.embeds) > 0 {
		// if there are embeds - add "related" part
		relatedPart := newMultiPartWithSubtype(cStringRelated) // need unref
		defer C.g_object_unref(relatedPart)                    // unref
		C.g_mime_multipart_add(relatedPart, contentPart)
		contentPart = anyToGMimeObject(unsafe.Pointer(relatedPart))
		addAttachments(relatedPart, m.embeds)
	}

	if len(m.attaches) > 0 {
		// if there are attaches - add "mixed" part
		mixedPart := newMultiPartWithSubtype(cStringMixed) // need unref
		defer C.g_object_unref(mixedPart)                  // unref
		C.g_mime_multipart_add(mixedPart, contentPart)
		contentPart = anyToGMimeObject(unsafe.Pointer(mixedPart))
		addAttachments(mixedPart, m.attaches)
	}

	message := C.g_mime_message_new(C.TRUE)
	defer C.g_object_unref(message)

	injectHeaders(anyToGMimeObject(unsafe.Pointer(message)), m.headers)

	C.g_mime_message_set_mime_part(message, contentPart)

	ostream := C.g_mime_stream_fs_new(C.dup(C.fileno(C.stdout)))
	defer C.g_object_unref(ostream)

	fmt.Println(">>>>> MESSAGE >>>>>")
	C.g_mime_object_write_to_stream((*C.GMimeObject)(unsafe.Pointer(message)), ostream)
	return nil
}

func (m *Message) String() error {
	return nil
}

func (m *Message) Print() {
	fmt.Println(m.gmimize())
}
