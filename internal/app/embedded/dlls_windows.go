//go:build windows

package embedded

import (
	"fmt"
	"os"
)

// NOTE: This file is a placeholder for DLL embedding on Windows.
//
// Windows cross-compilation from Linux with MuPDF/CGO is currently not supported
// due to MSVC/MinGW incompatibilities in the bundled MuPDF static libraries.
//
// To enable PDF cropping on Windows, you need to:
//
// 1. Build natively on Windows with MSVC toolchain:
//    - Install MSVC Build Tools
//    - Install Go
//    - Run: go build -tags dev -o lyyyra.exe
//
// 2. OR: Obtain Windows DLLs and embed them here:
//    - Download mupdf.dll (v1.24.15 for x64) from https://mupdf.com/releases
//    - Download libffi-8.dll from a reliable source
//    - Place them in this directory
//    - Uncomment the //go:embed directives below
//    - Uncomment the extraction code
//
// 3. OR: Use purego mode (build with -tags nocgo):
//    - Users must install MuPDF DLLs manually
//    - Set PATH to include DLL directory
//
// Example embedding (when you have the DLLs):
//
//   //go:embed mupdf.dll
//   var mupdfDLL []byte
//
//   //go:embed libffi-8.dll
//   var libffiDLL []byte

// EnsureDLLsExtracted is a stub that returns an error indicating DLLs are not embedded
func EnsureDLLsExtracted() (string, error) {
	// Check if DLLs are available in PATH or current directory
	// This would work if user has installed MuPDF system-wide
	if _, err := os.Stat("mupdf.dll"); err == nil {
		dir, _ := os.Getwd()
		return dir, nil
	}

	return "", fmt.Errorf("MuPDF DLLs not embedded in this build. " +
		"To use PDF cropping on Windows: " +
		"(1) Build natively on Windows with MSVC, " +
		"(2) Install MuPDF DLLs system-wide, or " +
		"(3) Place mupdf.dll and libffi-8.dll in the application directory")
}
