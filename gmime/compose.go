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
	"reflect"
	"unsafe"
)

var (
	cStringEmpty       = C.CString("")
	cStringAlternative = C.CString("alternative")
	cStringMixed       = C.CString("mixed")
	cStringRelated     = C.CString("related")
	cStringCharset     = C.CString("charset")
	cStringCharsetUTF8 = C.CString("utf-8")

	cStringText   = C.CString("text")
	cStringPlain  = C.CString("plain")
	cStringHTML   = C.CString("html")
	cStringBase64 = C.CString("base64")

	cStringContentID               = C.CString("Content-Id")
	cStringHeaderFormat            = C.CString("%s: %s\n")
	cStringContentTransferEncoding = C.CString("Content-Transfer-Encoding")
)

var (
	ErrNoContent = errors.New("No content (text or html)")
	ErrWrite     = errors.New("Error writing message to stream")
)

type AddressType int
type EncodingType int

const (
	AddressTo      AddressType = C.GMIME_RECIPIENT_TYPE_TO
	AddressCC                  = C.GMIME_RECIPIENT_TYPE_CC
	AddressFrom                = 100 + iota
	AddressReplyTo             = 100 + iota
)

var (
	EncodingDefault EncodingType = C.GMIME_CONTENT_ENCODING_DEFAULT
	EncodingBase64  EncodingType = C.GMIME_CONTENT_ENCODING_BASE64
)

type EmailAttachment struct {
	FileName       string
	MimeType       string
	ContentID      string
	Disposition    string
	Content        []byte
	InputEncoding  *EncodingType
	OutputEncoding *EncodingType
}

type Message struct {
	text      []byte
	html      []byte
	embeds    []*EmailAttachment
	attaches  []*EmailAttachment
	headers   []*EmailHeader
	addresses []*EmailAddress
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

type EmailAddress struct {
	AddressType AddressType
	Name        string
	Address     string
}

func NewMessage() *Message {
	return &Message{}
}

func (m *Message) SetText(body []byte) {
	m.text = body
}

func (m *Message) SetHtml(body []byte) {
	m.html = body
}

func (m *Message) Embed(a *EmailAttachment) {
	if a.Disposition == "" {
		a.Disposition = C.GMIME_DISPOSITION_INLINE
	}
	m.embeds = append(m.embeds, a)
}

func (m *Message) Attach(a *EmailAttachment) {
	if a.Disposition == "" {
		a.Disposition = C.GMIME_DISPOSITION_ATTACHMENT
	}
	m.attaches = append(m.attaches, a)
}

func (m *Message) AppendHeader(h *EmailHeader) {
	m.headers = append(m.headers, h)
}

func (m *Message) PrependHeader(h *EmailHeader) {
	m.headers = append([]*EmailHeader{h}, m.headers...)
}

func (m *Message) AddAddress(a *EmailAddress) {
	m.addresses = append(m.addresses, a)
}

func (m *Message) EncodedHeaders() []*EmailHeader {
	message := C.g_mime_message_new(C.TRUE)
	defer C.g_object_unref(message)
	injectHeaders(anyToGMimeObject(unsafe.Pointer(message)), m.headers, m.addresses)
	return encodedHeadersFromGmime(anyToGMimeObject(unsafe.Pointer(message)))
}

func (m *Message) Export() ([]byte, error) {
	message, err := m.gmimize()
	if err != nil {
		return nil, err
	}

	stream := C.g_mime_stream_mem_new() // need unref
	defer C.g_object_unref(stream)      // unref
	nWritten := C.g_mime_object_write_to_stream((*C.GMimeObject)(unsafe.Pointer(message)), stream)
	if nWritten <= 0 {
		return nil, ErrWrite
	}
	// byteArray is owned by stream and will be freed with it
	byteArray := C.g_mime_stream_mem_get_byte_array((*C.GMimeStreamMem)(unsafe.Pointer(stream)))
	return C.GoBytes(unsafe.Pointer(byteArray.data), (C.int)(nWritten)), nil
}

// BytesBorrow returns byte slice that caller has to return with BytesReturn
func (m *Message) Bytes() ([]byte, error) {
	message, err := m.gmimize()
	if err != nil {
		return nil, err
	}

	stream := C.g_mime_stream_mem_new() // need unref
	defer C.g_object_unref(stream)      // unref
	nWritten := C.g_mime_object_write_to_stream((*C.GMimeObject)(unsafe.Pointer(message)), stream)
	if nWritten <= 0 {
		return nil, ErrWrite
	}
	// byteArray is owned by stream and will be freed with it
	byteArray := C.g_mime_stream_mem_get_byte_array((*C.GMimeStreamMem)(unsafe.Pointer(stream)))
	C.g_mime_stream_mem_set_owner((*C.GMimeStreamMem)(unsafe.Pointer(stream)), C.FALSE) // tell stream that we own GByteArray
	h := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(byteArray.data)),
		Len:  (int)(nWritten),
		Cap:  (int)(nWritten),
	}
	s := *(*[]byte)(unsafe.Pointer(&h))
	C.g_byte_array_free(byteArray, C.FALSE) // free GByteArray structure, but keep byteArray->data allocated, we will free it in BytesReturn
	return s, nil
}

func (m *Message) Put(b []byte) {
	C.g_free(unsafe.Pointer(&b[0]))
}

func (m *Message) Print() error {
	message, err := m.gmimize()
	if err != nil {
		return err
	}
	ostream := C.g_mime_stream_fs_new(C.dup(C.fileno(C.stdout)))
	defer C.g_object_unref(ostream)

	fmt.Println(">>>>> MESSAGE >>>>>")
	C.g_mime_object_write_to_stream((*C.GMimeObject)(unsafe.Pointer(message)), ostream)
	return nil
}

// returns *GMimeMessage, need unref
func (m *Message) gmimize() (*C.GMimeMessage, error) {
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
		return nil, err
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

	message := C.g_mime_message_new(C.TRUE) // this message is returned, caller to unref

	injectHeaders(anyToGMimeObject(unsafe.Pointer(message)), m.headers, m.addresses)

	C.g_mime_message_set_mime_part(message, contentPart)

	return message, nil
}
