package services

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
)

// FileTypeInfo contains information about detected file types
// including MIME type and file extension.
type FileTypeInfo struct {
	MimeType  string
	Extension string
}

var (
	ErrInsufficientData = errors.New("insufficient data to determine file type")
	ErrUnknownFileType  = errors.New("unknown or unsupported file type")
)

// DetectFileType determines the file type from a byte array using magic numbers
// and returns a FileTypeInfo struct containing the MIME type and file extension.
// It checks for common file signatures (magic numbers) to identify the file type.
// The function returns an error if the data is insufficient or if the file type is unknown.
func DetectFileType(data []byte) (*FileTypeInfo, error) {
	if len(data) < 8 {
		return nil, ErrInsufficientData
	}

	// Check for various file signatures
	switch {
	// PDF: %PDF (25 50 44 46)
	case bytes.HasPrefix(data, []byte{0x25, 0x50, 0x44, 0x46}):
		return &FileTypeInfo{MimeType: "application/pdf", Extension: ".pdf"}, nil

	// TIFF (Intel): II* (49 49 2A 00)
	case bytes.HasPrefix(data, []byte{0x49, 0x49, 0x2A, 0x00}):
		return &FileTypeInfo{MimeType: "image/tiff", Extension: ".tiff"}, nil

	// TIFF (Motorola): MM* (4D 4D 00 2A)
	case bytes.HasPrefix(data, []byte{0x4D, 0x4D, 0x00, 0x2A}):
		return &FileTypeInfo{MimeType: "image/tiff", Extension: ".tiff"}, nil

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	case bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
		return &FileTypeInfo{MimeType: "image/png", Extension: ".png"}, nil

	// JPEG: FF D8 FF
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}):
		return &FileTypeInfo{MimeType: "image/jpeg", Extension: ".jpg"}, nil

	// BMP: BM (42 4D)
	case bytes.HasPrefix(data, []byte{0x42, 0x4D}):
		return &FileTypeInfo{MimeType: "image/bmp", Extension: ".bmp"}, nil

	default:
		return nil, ErrUnknownFileType
	}
}

// GetStandardizedExtension takes a filename and returns a standardized file extension.
// It converts the extension to lowercase and maps certain extensions to a standard format.
// For example, it converts ".jpeg", ".jpe", ".jif", and ".jfif" to ".jpg",
// and ".tif" to ".tiff". If the extension is not recognized, it defaults to ".bin".
// This function is useful for ensuring consistent file naming conventions across different file types.
func GetStandardizedExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".jpeg", ".jpe", ".jif", ".jfif":
		return ".jpg"
	case ".tif":
		return ".tiff"
	case ".pdf", ".png", ".jpg", ".tiff", ".bmp":
		return ext
	default:
		return ".bin" // Default binary extension for unknown types
	}
}
