// Package gmime implements the MIME, port of C GMime
package gmime

/*
#cgo pkg-config: gmime-2.6
#include <stdlib.h>
#include <gmime/gmime.h>

// pull out some glib guts
static gboolean object_is_object(GTypeInstance *obj) {
        return G_IS_OBJECT(obj);
}

*/
import "C"
import "unsafe"

// This function call automatically by runtime
func init() {
	println(">>> init gmime")
	C.g_mime_init(0)
}

//Shutdown This function really needed only for valgrind
func Shutdown() {
	C.g_mime_shutdown()
}

// convert from Go bool to C gboolean
func gbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}

// convert from C gboolean to Go bool
func gobool(b C.gboolean) bool {
	return b != C.gboolean(0)
}

// free up memory
func unref(referee C.gpointer) {
	C.g_object_unref(referee)
}

func anyToGMimeObject(obj unsafe.Pointer) *C.GMimeObject {
	return (*C.GMimeObject)(obj)
}