package main

/*
#cgo pkg-config: gmime-2.6
#include <stdlib.h>
#include <string.h>
#include <gmime/gmime.h>

// pull out some glib guts
static gboolean object_is_object(GTypeInstance *obj) {
        return G_IS_OBJECT(obj);
}

*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/sendgrid/go_gmime/gmime"
)

func main() {
	println(">>> gmime")
	msg := gmime.NewMessage()
	msg.SetText("test")
	msg.SetHtml("html")
	msg.Embed(&gmime.EmailAttachment{
		Content: []byte("test"),
	})
	msg.Attach(&gmime.EmailAttachment{
		Content: []byte("test attach"),
	})
	msg.AppendHeader(&gmime.EmailHeader{
		Name:  "test",
		Value: "testtesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttest",
		Raw:   true,
	})
	msg.AppendHeader(&gmime.EmailHeader{
		Name:  "Subject",
		Value: "test subject привет test subject привет test subject привет test subject привет test subject привет test subject привет",
	})
	msg.AppendHeader(&gmime.EmailHeader{
		Name:  "custom header",
		Value: "some value here",
	})

	msg.AddRecipient(&gmime.EmailAddress{
		AddressType: gmime.AddressTo,
		Name:        "John Doe",
		Address:     "john@example.com",
	})
	msg.AddRecipient(&gmime.EmailAddress{
		AddressType: gmime.AddressTo,
		Name:        "Кто то странный",
		Address:     "john_jr@example.com",
	})

	for _, h := range msg.EncodedHeaders() {
		fmt.Println(h)
	}

	b, _ := msg.BytesBorrow()
	fmt.Println(string(b))
	msg.BytesReturn(b)
	return

	subtype := C.CString("alternative")
	multipart := C.g_mime_multipart_new_with_subtype(subtype)

	text := C.CString("test")
	mem := C.g_mime_stream_mem_new_with_buffer(text, C.strlen(text))
	content := C.g_mime_data_wrapper_new_with_stream(mem, C.GMIME_CONTENT_ENCODING_DEFAULT)
	C.g_object_unref(mem)
	part := C.g_mime_part_new_with_type(C.CString("text"), C.CString("plain"))
	C.g_mime_object_set_content_type_parameter((*C.GMimeObject)(unsafe.Pointer(part)), C.CString("charset"), C.CString("utf-8"))
	C.g_mime_part_set_content_encoding(part, C.GMIME_CONTENT_ENCODING_QUOTEDPRINTABLE)
	C.g_mime_part_set_content_object(part, content)
	C.g_object_unref(content)

	//add part to multipart message
	C.g_mime_multipart_add(multipart, (*C.GMimeObject)(unsafe.Pointer(part)))

	text = C.CString("test html 웹문서, 이미지, 뉴스그룹, 디렉토리 검색, 한글 페이지 검색.")
	mem = C.g_mime_stream_mem_new_with_buffer(text, C.strlen(text))
	content = C.g_mime_data_wrapper_new_with_stream(mem, C.GMIME_CONTENT_ENCODING_DEFAULT)
	C.g_object_unref(mem)
	part = C.g_mime_part_new_with_type(C.CString("text"), C.CString("html"))
	C.g_mime_object_set_content_type_parameter((*C.GMimeObject)(unsafe.Pointer(part)), C.CString("charset"), C.CString("utf-8"))
	C.g_mime_part_set_content_encoding(part, C.GMIME_CONTENT_ENCODING_QUOTEDPRINTABLE)
	C.g_mime_part_set_content_object(part, content)
	C.g_object_unref(content)

	//add part to multipart message
	C.g_mime_multipart_add(multipart, (*C.GMimeObject)(unsafe.Pointer(part)))

	message := C.g_mime_message_new(C.TRUE)
	C.g_mime_message_set_mime_part(message, (*C.GMimeObject)(unsafe.Pointer(multipart)))
	C.g_object_unref(multipart)

	fmt.Println(message)
	ostream := C.g_mime_stream_fs_new(C.dup(C.fileno(C.stdout)))
	C.g_mime_object_write_to_stream((*C.GMimeObject)(unsafe.Pointer(message)), ostream)
	C.g_object_unref(ostream)
	// f, err := os.Open("testdata/test3.eml")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// data, err := ioutil.ReadAll(f)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // fmt.Println(string(data))

	// var cBuffer = C.CString(string(data))
	// // defer C.free(unsafe.Pointer(cBuffer))

	// strLen := len(data)
	// gb := C.g_byte_array_new_take((*C.guint8)(unsafe.Pointer(cBuffer)), (C.gsize)(strLen))
	// cStream := C.g_mime_stream_mem_new_with_byte_array(gb)
	// parser := C.g_mime_parser_new_with_stream(cStream)
	// message := C.g_mime_parser_construct_message(parser)
	// subject := C.g_mime_message_get_subject(message)
	// s := C.GoString(subject)
	// fmt.Println(s)
	// C.g_object_unref(cStream)
	// C.g_byte_array_unref(gb)
}
