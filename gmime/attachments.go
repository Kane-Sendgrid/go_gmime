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
	"mime"
	"path/filepath"
	"strings"
	"unsafe"
)

func addAttachments(obj *C.GMimeMultipart, attaches []*EmailAttachment) {
	for _, e := range attaches {
		text := (*C.char)(unsafe.Pointer(&e.Content[0]))
		inputEncoding := e.InputEncoding
		outputEncoding := e.OutputEncoding
		if inputEncoding == nil {
			inputEncoding = &EncodingDefault
		}
		if outputEncoding == nil {
			outputEncoding = &EncodingBase64
		}
		mem := C.g_mime_stream_mem_new_with_buffer(text, C.strlen(text))                                // needs unref
		content := C.g_mime_data_wrapper_new_with_stream(mem, (C.GMimeContentEncoding)(*inputEncoding)) // needs unref
		C.g_object_unref(mem)                                                                           // unref

		mediaType := e.MimeType
		if mediaType == "" {
			mediaType = mime.TypeByExtension(filepath.Ext(e.FileName))
		}
		if mediaType == "" {
			mediaType = "application/octet-stream"
		}

		mimeSplit := strings.Split(mediaType, "/")
		cStringMimeType := C.CString(mimeSplit[0])                               // needs free
		cStringMimeSubType := C.CString(mimeSplit[1])                            // needs free
		part := C.g_mime_part_new_with_type(cStringMimeType, cStringMimeSubType) // needs unref
		partObject := anyToGMimeObject(unsafe.Pointer(part))
		C.free(unsafe.Pointer(cStringMimeType))    // free
		C.free(unsafe.Pointer(cStringMimeSubType)) //free

		C.g_mime_part_set_content_encoding(part, (C.GMimeContentEncoding)(*outputEncoding))
		C.g_mime_part_set_content_object(part, content)
		C.g_object_unref(content) // unref
		if cid := e.ContentID; cid != "" {
			if cid[0] != '<' {
				cid = "<" + cid + ">"
			}
			contentID := C.CString(cid) // needs free
			C.g_mime_object_append_header(partObject, cStringContentID, contentID)
			C.free(unsafe.Pointer(contentID)) // free
		}

		disposition := C.CString(e.Disposition) // needs free
		C.g_mime_object_set_disposition(partObject, disposition)
		C.free(unsafe.Pointer(disposition)) // free

		if e.FileName != "" {
			fileName := C.CString(e.FileName) // needs free
			C.g_mime_part_set_filename(part, fileName)
			C.free(unsafe.Pointer(fileName)) // free
		}

		C.g_mime_multipart_add(obj, partObject)
		C.g_object_unref(part) // unref
	}

}
