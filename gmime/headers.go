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
	"strings"
	"unsafe"
)

func injectHeaders(obj *C.GMimeObject, headers []*EmailHeader, addresses []*EmailAddress) {
	headerList := C.g_mime_object_get_header_list(anyToGMimeObject(unsafe.Pointer(obj)))
	for _, h := range headers {
		name := C.CString(h.Name)   // needs free
		value := C.CString(h.Value) // needs free
		if h.Raw {
			C.g_mime_header_list_register_writer(headerList, name, (C.GMimeHeaderWriter)(unsafe.Pointer(C.raw_header_writer)))
			C.g_mime_object_prepend_header(anyToGMimeObject(unsafe.Pointer(obj)), name, value)
		} else {
			encodedValue := C.g_mime_utils_header_encode_text(value) // needs g_free
			//TODO: support append/prepend
			C.g_mime_object_prepend_header(anyToGMimeObject(unsafe.Pointer(obj)), name, encodedValue)
			C.g_free(C.gpointer(encodedValue))
		}

		C.free(unsafe.Pointer(name))
		C.free(unsafe.Pointer(value))
	}

	message := (*C.GMimeMessage)(unsafe.Pointer(obj))
	for _, a := range addresses {
		switch a.AddressType {
		case AddressTo, AddressCC:
			name := C.CString(a.Name)       // needs free
			address := C.CString(a.Address) // needs free
			C.g_mime_message_add_recipient(message, (C.GMimeRecipientType)(a.AddressType), name, address)
			C.free(unsafe.Pointer(name))
			C.free(unsafe.Pointer(address))
		case AddressFrom:
			addr := C.CString(a.Name + " " + "<" + a.Address + ">")
			C.g_mime_message_set_sender(message, addr)
			C.free(unsafe.Pointer(addr))
		case AddressReplyTo:
			addr := C.CString(a.Name + " " + "<" + a.Address + ">")
			C.g_mime_message_set_reply_to(message, addr)
			C.free(unsafe.Pointer(addr))
		}
	}
}

func encodedHeadersFromGmime(obj *C.GMimeObject) []*EmailHeader {
	var iter C.GMimeHeaderIter
	var returnHeaders []*EmailHeader
	headerList := C.g_mime_object_get_header_list(anyToGMimeObject(unsafe.Pointer(obj)))
	if C.g_mime_header_list_get_iter(headerList, &iter) == C.TRUE {
		writers := (*C.struct_CustomGMimeHeaderList)(unsafe.Pointer(headerList)).writers
		for {
			if val := C.g_mime_header_iter_get_value(&iter); val != nil {
				encodedHeaderValue := ""
				name := C.g_mime_header_iter_get_name(&iter)
				writer := (C.GMimeHeaderWriter)(C.g_hash_table_lookup(writers, name))
				if writer != nil {
					stream := C.g_mime_stream_mem_new()
					C.call_writer(writer, stream, name, val)
					byteArray := C.g_mime_stream_mem_get_byte_array((*C.GMimeStreamMem)(unsafe.Pointer(stream)))
					C.g_byte_array_append(byteArray, (*C.guint8)(unsafe.Pointer(cStringEmpty)), 1)
					encodedHeaderValue = C.GoString((*C.char)(unsafe.Pointer(byteArray.data)))
					C.g_object_unref(stream)
				} else {
					data := C.call_g_mime_utils_header_printf(cStringHeaderFormat, name, val)
					encodedHeaderValue = C.GoString((*C.char)(unsafe.Pointer(data)))
					C.g_free(data)
				}

				headerNameValue := strings.SplitN(encodedHeaderValue, ": ", 2)
				returnHeaders = append(returnHeaders, &EmailHeader{
					Name:  headerNameValue[0],
					Value: headerNameValue[1],
				})
			}
			if C.g_mime_header_iter_next(&iter) == C.FALSE {
				break
			}
		}
	}
	return returnHeaders
}
