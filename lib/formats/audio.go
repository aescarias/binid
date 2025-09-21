package formats

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

/* ParseQOA extracts information from a Quite Okay Audio (QOA) file. */
func ParseQOA(rs io.ReadSeeker) (map[string]any, error) {
	if _, err := rs.Seek(4, 0); err != nil {
		return nil, err
	}

	samplesBytes := make([]byte, 4)
	if _, err := rs.Read(samplesBytes); err != nil {
		return nil, err
	}

	samples := binary.BigEndian.Uint32(samplesBytes)
	frames := uint32(math.Ceil(float64(samples) / (256 * 20)))

	frameDescribes := make([]string, frames)

	for idx := range frames {
		numChannelsBytes := make([]byte, 1)
		if _, err := rs.Read(numChannelsBytes); err != nil {
			return nil, err
		}

		numChannels := numChannelsBytes[0]

		sampleRateBytes := make([]byte, 3)
		if _, err := rs.Read(sampleRateBytes); err != nil {
			return nil, err
		}

		sampleRate := uint32(sampleRateBytes[0])<<16 | uint32(sampleRateBytes[1])<<8 | uint32(sampleRateBytes[2])

		frameSamplesBytes := make([]byte, 2)
		if _, err := rs.Read(frameSamplesBytes); err != nil {
			return nil, err
		}
		frameSamples := binary.BigEndian.Uint16(frameSamplesBytes)

		frameSizeBytes := make([]byte, 2)
		if _, err := rs.Read(frameSizeBytes); err != nil {
			return nil, err
		}
		frameSize := binary.BigEndian.Uint16(frameSizeBytes)

		const headerSize uint16 = 8
		if _, err := rs.Seek(int64(frameSize-headerSize), 1); err != nil {
			return nil, err
		}

		frameDescribes[idx] = fmt.Sprintf("%d channel(s) @ %d hz, %d sample(s) per channel", numChannels, sampleRate, frameSamples)
	}

	return map[string]any{
		"samples per channel": samples,
		"frames":              frameDescribes,
	}, nil
}
