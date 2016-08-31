// +build darwin

package gmime

/*
#cgo pkg-config: gmime-2.6
#include <stdlib.h>
#include <gmime/gmime.h>
*/
import "C"

import (
	"crypto/rand"
	"encoding/base64"
	"unsafe"
)

// generateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// generateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomString(s int) string {
	b, _ := generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b)
}

func newMultiPartWithSubtype(subtype *C.char) *C.GMimeMultipart {
	p := C.g_mime_multipart_new_with_subtype(subtype)
	b := C.CString(generateRandomString(20))
	defer C.free(unsafe.Pointer(b))
	C.g_mime_multipart_set_boundary(p, b)
	return p
}
