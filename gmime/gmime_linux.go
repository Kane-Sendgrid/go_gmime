// +build !darwin

package gmime

/*
#cgo pkg-config: gmime-2.6
#include <gmime/gmime.h>
*/
import "C"

func newMultiPartWithSubtype(subtype *C.char) *C.GMimeMultiPart {
	return C.g_mime_multipart_new_with_subtype(subtype)
}
