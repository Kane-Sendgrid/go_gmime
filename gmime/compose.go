package gmime

/*
#cgo pkg-config: gmime-2.6
#include <stdlib.h>
#include <stdarg.h>
#include <string.h>
#include <gmime/gmime.h>

struct CustomGMimeHeaderList {
	GMimeStream *stream;
	GHashTable *writers;
};

char *call_g_mime_utils_header_printf(const char *format, const char *name, const char *value) {
	return g_mime_utils_header_printf(format, name, value);
}

ssize_t call_writer(GMimeHeaderWriter writer, GMimeStream *stream, const char *name, const char *value) {
	return writer(stream, name, value);
}

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
	"strings"
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

	cStringContentID   = C.CString("Content-Id")
	cStringApplication = C.CString("application")
	cStringOctetStream = C.CString("octet-stream")
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

func (m *Message) EncodedHeaders() {

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
	contentPart, err := textHTMLPart(m.text, m.html)
	if err != nil {
		return err
	}
	defer C.g_object_unref(contentPart)

	if len(m.embeds) > 0 {
		// if there are embeds - add "related" part
		relatedPart := newMultiPartWithSubtype(cStringRelated)
		defer C.g_object_unref(relatedPart)
		C.g_mime_multipart_add(relatedPart, contentPart)
		contentPart = anyToGMimeObject(unsafe.Pointer(relatedPart))

		for _, e := range m.embeds {
			text := (*C.char)(unsafe.Pointer(&e.Content[0]))
			mem := C.g_mime_stream_mem_new_with_buffer(text, C.strlen(text))                        // needs unref
			content := C.g_mime_data_wrapper_new_with_stream(mem, C.GMIME_CONTENT_ENCODING_DEFAULT) // needs unref
			C.g_object_unref(mem)

			var part *C.GMimePart
			mimeSplit := strings.Split(e.MimeType, "/")
			if len(mimeSplit) == 2 {
				cStringMimeType := C.CString(mimeSplit[0])
				cStringMimeSubType := C.CString(mimeSplit[1])
				part = C.g_mime_part_new_with_type(cStringMimeType, cStringMimeSubType) // needs unref
				C.free(unsafe.Pointer(cStringMimeType))
				C.free(unsafe.Pointer(cStringMimeSubType))
			} else {
				part = C.g_mime_part_new_with_type(cStringApplication, cStringOctetStream) // needs unref
			}
			C.g_mime_part_set_content_encoding(part, C.GMIME_CONTENT_ENCODING_BASE64)
			C.g_mime_part_set_content_object(part, content)
			C.g_object_unref(content)
			C.g_mime_object_append_header(anyToGMimeObject(unsafe.Pointer(part)), cStringContentID, C.CString("test"))
			C.g_mime_multipart_add(relatedPart, anyToGMimeObject(unsafe.Pointer(part)))
			C.g_object_unref(part)
		}
	}

	message := C.g_mime_message_new(C.TRUE)
	defer C.g_object_unref(message)

	headerList := C.g_mime_object_get_header_list(anyToGMimeObject(unsafe.Pointer(message)))
	for _, h := range m.headers {
		name := C.CString(h.Name)   // needs free
		value := C.CString(h.Value) // needs free
		if h.Raw {
			C.g_mime_header_list_register_writer(headerList, name, (C.GMimeHeaderWriter)(unsafe.Pointer(C.raw_header_writer)))
			C.g_mime_object_prepend_header(anyToGMimeObject(unsafe.Pointer(message)), name, value)
		} else {
			encodedValue := C.g_mime_utils_header_encode_text(value) // needs g_free
			//TODO: support append/prepend
			C.g_mime_object_prepend_header(anyToGMimeObject(unsafe.Pointer(message)), name, encodedValue)
			C.g_free(C.gpointer(encodedValue))
		}

		C.free(unsafe.Pointer(name))
		C.free(unsafe.Pointer(value))
	}

	var iter C.GMimeHeaderIter
	ok := C.g_mime_header_list_get_iter(headerList, &iter)
	if ok == C.TRUE {
		writers := (*C.struct_CustomGMimeHeaderList)(unsafe.Pointer(headerList)).writers
		for {
			if val := C.g_mime_header_iter_get_value(&iter); val != nil {
				name := C.g_mime_header_iter_get_name(&iter)
				writer := (C.GMimeHeaderWriter)(C.g_hash_table_lookup(writers, name))
				if writer != nil {
					stream := C.g_mime_stream_mem_new()
					C.call_writer(writer, stream, name, val)
					byteArray := C.g_mime_stream_mem_get_byte_array((*C.GMimeStreamMem)(unsafe.Pointer(stream)))
					C.g_byte_array_append(byteArray, (*C.guint8)(unsafe.Pointer(cStringEmpty)), 1)
					str := C.GoString((*C.char)(unsafe.Pointer(byteArray.data)))
					fmt.Println(str)
					C.g_object_unref(stream)
				} else {
					data := C.call_g_mime_utils_header_printf(C.CString("%s: %s\n"), name, val)
					str := C.GoString((*C.char)(unsafe.Pointer(data)))
					fmt.Println(str)
					C.g_free(data)
				}
			}
			more := C.g_mime_header_iter_next(&iter)
			if more == C.FALSE {
				break
			}
		}
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
