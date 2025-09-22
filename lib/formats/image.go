package formats

import (
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
)

type PDNHeader struct {
	Width   int    `xml:"width,attr"`
	Height  int    `xml:"height,attr"`
	Layers  int    `xml:"layers,attr"`
	Version string `xml:"savedWithVersion,attr"`
}

func ParsePDN(reader io.ReadSeeker) (map[string]any, error) {
	if _, err := reader.Seek(4, 0); err != nil {
		return nil, err
	}

	headerSizeBytes := make([]byte, 4)
	if _, err := reader.Read(headerSizeBytes[0:3]); err != nil {
		return nil, err
	}
	headerSizeBytes[3] = 0

	headerSize := binary.LittleEndian.Uint32(headerSizeBytes)

	headerBytes := make([]byte, headerSize)
	if _, err := reader.Read(headerBytes); err != nil {
		return nil, err
	}

	var pdn PDNHeader
	if err := xml.Unmarshal(headerBytes, &pdn); err != nil {
		return nil, err
	}

	return map[string]any{
		"size":               fmt.Sprintf("%dx%d pixels", pdn.Width, pdn.Height),
		"layers":             pdn.Layers,
		"saved with version": pdn.Version,
	}, nil
}

/* ParseICO extracts information from a Windows icon (.ico) file. */
func ParseICO(reader io.ReadSeeker) (map[string]any, error) {
	if _, err := reader.Seek(4, 0); err != nil {
		return nil, err
	}

	nImagesBytes := make([]byte, 2)
	if _, err := reader.Read(nImagesBytes); err != nil {
		return nil, err
	}
	nImages := binary.LittleEndian.Uint16(nImagesBytes)

	iconDescribes := make([]string, nImages)

	for idx := range nImages {
		widthBytes := make([]byte, 1)
		if _, err := reader.Read(widthBytes); err != nil {
			return nil, err
		}
		width := int(widthBytes[0])
		if width == 0 {
			width = 256
		}

		heightBytes := make([]byte, 1)
		if _, err := reader.Read(heightBytes); err != nil {
			return nil, err
		}
		height := int(heightBytes[0])
		if height == 0 {
			height = 256
		}

		colorCountBytes := make([]byte, 1)
		if _, err := reader.Read(colorCountBytes); err != nil {
			return nil, err
		}
		colorCount := int(colorCountBytes[0])

		reservedBytes := make([]byte, 1)
		if _, err := reader.Read(reservedBytes); err != nil {
			return nil, err
		}

		planesBytes := make([]byte, 2)
		if _, err := reader.Read(planesBytes); err != nil {
			return nil, err
		}
		planes := binary.LittleEndian.Uint16(planesBytes)

		bitCountBytes := make([]byte, 2)
		if _, err := reader.Read(bitCountBytes); err != nil {
			return nil, err
		}
		bitCount := binary.LittleEndian.Uint16(bitCountBytes)

		imageSizeBytes := make([]byte, 4)
		if _, err := reader.Read(imageSizeBytes); err != nil {
			return nil, err
		}
		imageSize := binary.LittleEndian.Uint32(imageSizeBytes)

		imageOffsetBytes := make([]byte, 4)
		if _, err := reader.Read(imageOffsetBytes); err != nil {
			return nil, err
		}
		imageOffset := binary.LittleEndian.Uint32(imageOffsetBytes)

		if colorCount == 0 { // no color palette
			iconDescribes[idx] = fmt.Sprintf("%dx%d pixels, %d color plane(s), %d bpp, %d bytes at offset %d", width, height, planes, bitCount, imageSize, imageOffset)
		} else {
			iconDescribes[idx] = fmt.Sprintf("%dx%d pixels, %d color(s), %d color plane(s), %d bpp, %d bytes at offset %d", width, height, colorCount, planes, bitCount, imageSize, imageOffset)
		}
	}

	return map[string]any{"icons": iconDescribes}, nil
}

/* ParseQOI extracts information from Quite Ok Image data. */
func ParseQOI(reader io.ReadSeeker) (map[string]any, error) {
	if _, err := reader.Seek(4, 0); err != nil {
		return nil, err
	}

	widthBytes := make([]byte, 4)
	if _, err := reader.Read(widthBytes); err != nil {
		return nil, err
	}
	width := binary.BigEndian.Uint32(widthBytes)

	heightBytes := make([]byte, 4)
	if _, err := reader.Read(heightBytes); err != nil {
		return nil, err
	}
	height := binary.BigEndian.Uint32(heightBytes)

	channelsBytes := make([]byte, 1)
	if _, err := reader.Read(channelsBytes); err != nil {
		return nil, err
	}
	channels := int(channelsBytes[0])

	colorspaceBytes := make([]byte, 1)
	if _, err := reader.Read(colorspaceBytes); err != nil {
		return nil, err
	}
	colorspace := int(colorspaceBytes[0])

	var csName string
	switch channels {
	case 3:
		csName = "RGB"
	case 4:
		csName = "RGBA"
	default:
		csName = "unknown color space"
	}

	var csKind string
	switch colorspace {
	case 0:
		csKind = "sRGB with linear alpha"
	case 1:
		csKind = "all channels linear"
	default:
		csKind = "unknown type"
	}

	return map[string]any{
		"size":       fmt.Sprintf("%dx%d pixels", width, height),
		"colorspace": fmt.Sprintf("%s (%s)", csName, csKind),
	}, nil
}
