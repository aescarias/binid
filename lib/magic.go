package lib

import (
	"bytes"
	"io"

	"github.com/aescarias/binid/lib/formats"
)

type FileNarrower func(io.ReadSeeker) (bool, error)
type FileExtractor func(io.ReadSeeker) (map[string]any, error)

// A Tag describes the method used to identify a file type.
type Tag struct {
	Name      string        // The common name of the file type.
	Mime      string        // The MIME or media type assigned to this file type.
	Narrower  FileNarrower  // A function that determines whether a file matches the tag.
	Extractor FileExtractor // A function that extracts more information from a file once it's matched.
}

// ReadBlock reads a block from a reader starting from offset and ending at offset + length.
func ReadBlock(reader io.ReadSeeker, length int, offset int64) ([]byte, error) {
	if _, err := reader.Seek(0, 0); err != nil {
		return nil, err
	}

	block := make([]byte, length)
	if _, err := reader.Read(block); err != nil {
		return nil, err
	}

	return block, nil
}

// NarrowMagic produces a narrower for files with tag at position start.
func NarrowMagic(tag []byte, start int64) FileNarrower {
	return func(rs io.ReadSeeker) (bool, error) {
		block, err := ReadBlock(rs, len(tag), start)
		if err != nil {
			return false, err
		}

		return bytes.Equal(block, tag), nil
	}
}

// NarrowRiffWithIdent produces a narrower for RIFF containers with tag.
func NarrowRiffWithIdent(tag []byte) FileNarrower {
	return func(rs io.ReadSeeker) (bool, error) {
		block, err := ReadBlock(rs, 16, 0)
		if err != nil {
			return false, err
		}

		return bytes.Equal(block[0:4], []byte("RIFF")) && bytes.Equal(block[8:12], tag), nil
	}
}

// NarrowJpeg determines whether a file is some form of JPEG.
func NarrowJpeg(rs io.ReadSeeker) (bool, error) {
	jpegMagic := [][]byte{
		[]byte("\xff\xd8\xff\xdb"),
		[]byte("\xff\xd8\xff\xe0\x00\x10\x4a\x46\x49\x46\x00\x01"), // JFIF
		[]byte("\xff\xd8\xff\xee"),
		[]byte("\xff\xd8\xff\xe0"),
		[]byte("\x00\x00\x00\x0c\x6a\x50\x20\x20\x0d\x0a\x87\x0a"), // JPEG 2000
		[]byte("\xff\x4f\xff\x51"),                                 // ...
	}

	for _, magic := range jpegMagic {
		block, err := ReadBlock(rs, len(magic), 0)
		if err != nil {
			return false, err
		}

		if bytes.Equal(block, magic) {
			return true, nil
		}
	}

	// EXIF variant
	block, err := ReadBlock(rs, 12, 0)
	if err != nil {
		return false, err
	}

	if bytes.Equal(block[0:4], []byte("\xff\xd8\xff\xe1")) && bytes.Equal(block[6:12], []byte("Exif\x00\x00")) {
		return true, nil
	}

	return false, nil
}

var MagicTags = []Tag{
	{
		Name:      "Quite Ok Image (QOI) data",
		Mime:      "image/x-qoi",
		Narrower:  NarrowMagic([]byte("qoif"), 0),
		Extractor: formats.ParseQOI,
	},
	{
		Name:      "Quite Ok Audio (QOA) data",
		Mime:      "audio/x-qoa",
		Narrower:  NarrowMagic([]byte("qoaf"), 0),
		Extractor: formats.ParseQOA,
	},
	// RIFF-based formats
	{
		Name:     "Waveform Audio file",
		Mime:     "audio/wav",
		Narrower: NarrowRiffWithIdent([]byte("WAVE")),
	},
	{
		Name:     "Audio Video Interleave (AVI) file",
		Mime:     "video/x-msvideo",
		Narrower: NarrowRiffWithIdent([]byte("AVI\x20")),
	},
	{
		Name:     "WebP image",
		Mime:     "image/webp",
		Narrower: NarrowRiffWithIdent([]byte("WEBP")),
	},
	{
		Name:     "Generic RIFF container",
		Mime:     "application/x-riff",
		Narrower: NarrowMagic([]byte("RIFF"), 0),
	}, // The generic RIFF container shall be after the specific ones.
	{
		Name:     "Compound File Binary Format",
		Mime:     "application/x-ole-storage",
		Narrower: NarrowMagic([]byte("\xd0\xcf\x11\xe0\xa1\xb1\x1a\xe1"), 0),
	},
	{
		Name:     "Adobe Portable Document Format",
		Mime:     "application/pdf",
		Narrower: NarrowMagic([]byte("%PDF-"), 0),
	},
	{
		Name:     "MZ Portable Executable",
		Mime:     "application/x-msdownload",
		Narrower: NarrowMagic([]byte("MZ"), 0),
	},
	{
		Name:     "Windows Bitmap file",
		Mime:     "image/bmp",
		Narrower: NarrowMagic([]byte("BM"), 0),
	},
	{
		Name:     "Adobe Photoshop Document",
		Mime:     "image/vnd.adobe.photoshop",
		Narrower: NarrowMagic([]byte("8BPS"), 0),
	},
	{
		Name:     "Executable and Linkable Format",
		Mime:     "application/x-executable",
		Narrower: NarrowMagic([]byte("\x7fELF"), 0),
	},
	{
		Name:     "ZIP compressed archive",
		Mime:     "application/zip",
		Narrower: NarrowMagic([]byte("PK\x03\x04"), 0),
	},
	{
		Name:     "Portable Network Graphics (PNG) image",
		Mime:     "image/png",
		Narrower: NarrowMagic([]byte("\x89\x50\x4e\x47\x0d\x0a\x1a\x0a"), 0),
	},
	{
		Name:     "GZip compressed file",
		Mime:     "application/gzip",
		Narrower: NarrowMagic([]byte("\x1f\x8b"), 0),
	},
	{
		Name:     "Roshal Archive (RAR) v1.5+",
		Mime:     "application/x-rar-compressed",
		Narrower: NarrowMagic([]byte("Rar!\x1a\x07\x00"), 0),
	},
	{
		Name:     "Roshal Archive (RAR) v5.0+",
		Mime:     "application/x-rar-compressed",
		Narrower: NarrowMagic([]byte("Rar!\x1a\x07\x01\x00"), 0),
	},
	{
		Name:     "Extended Module (XM)",
		Mime:     "audio/x-xm",
		Narrower: NarrowMagic([]byte("Extended Module: "), 0),
	},
	{
		Name:     "TrueType Font",
		Mime:     "font/ttf",
		Narrower: NarrowMagic([]byte("\x00\x01\x00\x00\x00"), 0),
	},
	{
		Name:     "OpenType Font",
		Mime:     "font/otf",
		Narrower: NarrowMagic([]byte("OTTO"), 0),
	},
	{
		Name:     "TIFF (little-endian)",
		Mime:     "image/tiff",
		Narrower: NarrowMagic([]byte("II\x2a\x00"), 0),
	},
	{
		Name:     "TIFF (big-endian)",
		Mime:     "image/tiff",
		Narrower: NarrowMagic([]byte("MM\x00\x2a"), 0),
	},
	{
		Name:     "Ogg Container",
		Mime:     "application/ogg",
		Narrower: NarrowMagic([]byte("OggS"), 0),
	},
	{
		Name:      "Windows Icon",
		Mime:      "image/x-icon",
		Narrower:  NarrowMagic([]byte("\x00\x00\x01\x00"), 0),
		Extractor: formats.ParseICO,
	},
	{
		Name:     "Graphics Interchange Format (GIF) version 87",
		Mime:     "image/gif",
		Narrower: NarrowMagic([]byte("GIF87a"), 0),
	},
	{
		Name:     "Graphics Interchange Format (GIF) version 89",
		Mime:     "image/gif",
		Narrower: NarrowMagic([]byte("GIF89a"), 0),
	},
	{
		Name:     "Sphinx Objects Inventory version 2",
		Mime:     "application/x-intersphinx",
		Narrower: NarrowMagic([]byte("# Sphinx inventory version 2"), 0),
	},
	{
		Name:     "JPEG image",
		Mime:     "image/jpeg",
		Narrower: NarrowJpeg,
	},
	{
		Name:     "SQLite database file",
		Mime:     "application/vnd.sqlite3",
		Narrower: NarrowMagic([]byte("SQLite format 3\x00"), 0),
	},
}
