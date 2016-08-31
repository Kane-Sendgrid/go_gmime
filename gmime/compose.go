package gmime

/*
#cgo pkg-config: gmime-2.6
#include <stdlib.h>
#include <string.h>
#include <gmime/gmime.h>

ssize_t raw_header_writer(GMimeStream *stream, const char *name, const char *value)
{
	ssize_t nwritten;
	char *val;

	val = g_strdup_printf("%s: %s\n", name, value);
	nwritten = g_mime_stream_write_string (stream, val);
	g_free (val);

	return nwritten;
}
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

var (
	cStringAlternative = C.CString("alternative")
	cStringMixed       = C.CString("mixed")
	cStringRelated     = C.CString("related")
	cStringCharset     = C.CString("charset")
	cStringCharsetUTF8 = C.CString("utf-8")

	cStringText  = C.CString("text")
	cStringPlain = C.CString("plain")
	cStringHTML  = C.CString("html")
)

var (
	ErrNoContent = errors.New("No content (text or html)")
)

type EmailAttachment struct {
	fileName    string
	mimeType    string
	contentID   string
	disposition string
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

func (m *Message) gmimize() error {
	var contentPart *C.GMimeObject
	contentPart, err := textHTMLPart(m.text, m.html)
	if err != nil {
		return err
	}
	defer C.g_object_unref(contentPart)

	if len(m.embeds) > 0 {
		// should be multipart/related(wrapping multipart/alternative and embeds)
		// - related
		//     - alternative
		//         - text/plain
		//         - text/html
		//     - embedded image 1
		//     - embedded image 2
		relatedPart := newMultiPartWithSubtype(cStringRelated)
		defer C.g_object_unref(relatedPart)
		C.g_mime_multipart_add(relatedPart, contentPart)
		contentPart = anyToGMimeObject(unsafe.Pointer(relatedPart))

		for _, e := range m.embeds {
			text := (*C.char)(unsafe.Pointer(&e.Content[0]))
			mem := C.g_mime_stream_mem_new_with_buffer(text, C.strlen(text))
			content := C.g_mime_data_wrapper_new_with_stream(mem, C.GMIME_CONTENT_ENCODING_DEFAULT)
			C.g_object_unref(mem)

			//TODO: set mime type
			part := C.g_mime_part_new_with_type(C.CString("application"), C.CString("octet-stream"))
			C.g_mime_part_set_content_encoding(part, C.GMIME_CONTENT_ENCODING_BASE64)
			C.g_mime_part_set_content_object(part, content)
			C.g_object_unref(content)
			C.g_mime_object_append_header(anyToGMimeObject(unsafe.Pointer(part)), C.CString("Content-Id"), C.CString("test"))
			C.g_mime_multipart_add(relatedPart, anyToGMimeObject(unsafe.Pointer(part)))
			C.g_object_unref(part)
		}
	}

	message := C.g_mime_message_new(C.TRUE)
	defer C.g_object_unref(message)

	headerList := C.g_mime_object_get_header_list(anyToGMimeObject(unsafe.Pointer(message)))
	C.g_mime_header_list_register_writer(headerList, C.CString("test"), (C.GMimeHeaderWriter)(unsafe.Pointer(C.raw_header_writer)))

	for _, h := range m.headers {
		C.g_mime_object_append_header(anyToGMimeObject(unsafe.Pointer(message)), C.CString(h.Name), C.CString(h.Value))
	}

	C.g_mime_message_set_mime_part(message, contentPart)

	ostream := C.g_mime_stream_fs_new(C.dup(C.fileno(C.stdout)))
	defer C.g_object_unref(ostream)

	C.g_mime_object_write_to_stream((*C.GMimeObject)(unsafe.Pointer(message)), ostream)
	return nil
}

func (m *Message) String() error {
	return nil
}

func (m *Message) Print() {
	fmt.Println(m.gmimize())
}
