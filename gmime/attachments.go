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
	"strings"
	"unsafe"
)

func addAttachments(obj *C.GMimeMultipart, attaches []*EmailAttachment) {
	for _, e := range attaches {
		text := (*C.char)(unsafe.Pointer(&e.Content[0]))
		mem := C.g_mime_stream_mem_new_with_buffer(text, C.strlen(text))                        // needs unref
		content := C.g_mime_data_wrapper_new_with_stream(mem, C.GMIME_CONTENT_ENCODING_DEFAULT) // needs unref
		C.g_object_unref(mem)                                                                   // unref

		var part *C.GMimePart // needs unref
		mimeSplit := strings.Split(e.MimeType, "/")
		if len(mimeSplit) == 2 {
			cStringMimeType := C.CString(mimeSplit[0])
			cStringMimeSubType := C.CString(mimeSplit[1])
			part = C.g_mime_part_new_with_type(cStringMimeType, cStringMimeSubType)
			C.free(unsafe.Pointer(cStringMimeType))
			C.free(unsafe.Pointer(cStringMimeSubType))
		} else {
			part = C.g_mime_part_new_with_type(cStringApplication, cStringOctetStream)
		}
		C.g_mime_part_set_content_encoding(part, C.GMIME_CONTENT_ENCODING_BASE64)
		C.g_mime_part_set_content_object(part, content)
		C.g_object_unref(content) // unref
		C.g_mime_object_append_header(anyToGMimeObject(unsafe.Pointer(part)), cStringContentID, C.CString("test"))
		C.g_mime_multipart_add(obj, anyToGMimeObject(unsafe.Pointer(part)))
		C.g_object_unref(part) // unref
	}

}
