package mupdf_shims

/*
#include <string.h>
#include <stdlib.h>

// Buffer check shims
void * __memmove_chk(void *dest, const void *src, size_t len, size_t destlen) {
    return memmove(dest, src, len);
}

void * __memcpy_chk(void *dest, const void *src, size_t len, size_t destlen) {
    return memcpy(dest, src, len);
}

void * __memset_chk(void *s, int c, size_t n, size_t destlen) {
    return memset(s, c, n);
}

char * __strcat_chk(char *dest, const char *src, size_t destlen) {
    return strcat(dest, src);
}

// Security check fail shim
void __chk_fail(void) {
    abort();
}

// MSVC/Clang setjmp intrinsic shim for MinGW
int __intrinsic_setjmpex(void *env, void *frame) {
    return 0;
}
*/
import "C"
