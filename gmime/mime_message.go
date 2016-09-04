package gmime

/*
#cgo pkg-config: gmime-2.6
#include <stdlib.h>
#include <gmime/gmime.h>
*/
import "C"
import "unsafe"

type EncodedHeader struct {
	Name  string
	Value string
}

type MIMEMessage struct {
	EncodedHeaders []*EncodedHeader
	Body           []byte
}

func (m *MIMEMessage) Put() {
	C.g_free(unsafe.Pointer(&m.Body[0]))
}
