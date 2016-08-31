package gmime

/*
#cgo pkg-config: gmime-2.6
#include <stdlib.h>
#include <string.h>
#include <gmime/gmime.h>
*/
import "C"
import "unsafe"

// returns MimePart as GMimeObject
// caller is responsible for unref
func mimePartFromString(body string) *C.GMimeObject {
	text := C.CString(body)
	defer C.free(unsafe.Pointer(text))

	mem := C.g_mime_stream_mem_new_with_buffer(text, C.strlen(text))
	defer C.g_object_unref(mem)

	content := C.g_mime_data_wrapper_new_with_stream(mem, C.GMIME_CONTENT_ENCODING_DEFAULT)
	defer C.g_object_unref(content)

	part := C.g_mime_part_new_with_type(cStringText, cStringPlain)
	textPart := anyToGMimeObject(unsafe.Pointer(part))

	C.g_mime_object_set_content_type_parameter(textPart, cStringCharset, cStringCharsetUTF8)
	C.g_mime_part_set_content_encoding(part, C.GMIME_CONTENT_ENCODING_QUOTEDPRINTABLE)
	C.g_mime_part_set_content_object(part, content)

	return textPart
}

// returns GMimeObject
// caller responsible for unref
func textHTMLPart(text, html string) (*C.GMimeObject, error) {
	var textPart, htmlPart *C.GMimeObject

	if text != "" {
		textPart = mimePartFromString(text)
	}

	if html != "" {
		htmlPart = mimePartFromString(html)
	}

	switch {
	case text != "" && html != "":
		// should be multipart/alternative
		multipart := C.g_mime_multipart_new_with_subtype(cStringAlternative)
		C.g_mime_multipart_add(multipart, textPart)
		C.g_mime_multipart_add(multipart, htmlPart)
		return anyToGMimeObject(unsafe.Pointer(multipart)), nil
	case text != "":
		// only text part
		return textPart, nil
	case html != "":
		// only html part
		return htmlPart, nil
	default:
		return nil, ErrNoContent
	}
}
