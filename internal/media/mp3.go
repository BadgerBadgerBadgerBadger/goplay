package media

import (
	"io"

	"github.com/pkg/errors"
)

type MP3Frame struct {
	Header        MP3Header
	SideInfo      MP3SideInfo
	MainData      []byte
	AncillaryData []byte
}

type MP3Header struct {
	// SyncWord is a 11-bit field that is always set to "11111111111" (0x7FF) to identify the beginning of a valid frame.
	SyncWord uint16

	// MPEGVersion is the MPEG version for the frame. Possible values are 1 (MPEG-1) and 2 (MPEG-2).
	MPEGVersion uint8

	// Layer is the audio layer for the frame. Possible values are 1, 2, or 3 (Layer I, II, or III).
	Layer uint8

	// ProtectionBit indicates whether the frame has CRC protection. A value of 1 means no CRC, while 0 means CRC is present.
	ProtectionBit bool

	// BitrateIndex is the index used to determine the bitrate. Use a lookup table based on the MPEG version and layer to find the actual bitrate.
	BitrateIndex uint8

	// SamplingRateIndex is the index used to determine the sampling rate. Use a lookup table based on the MPEG version to find the actual sampling rate.
	SamplingRateIndex uint8

	// PaddingBit indicates whether the frame has padding. A value of 1 means there is a padding byte at the end of the frame.
	PaddingBit bool

	// PrivateBit is a private bit that has no specified meaning and can be used for application-specific purposes.
	PrivateBit bool

	// ChannelMode is the channel mode for the frame. Possible values are 0 (Stereo), 1 (Joint Stereo), 2 (Dual Channel), and 3 (Mono).
	ChannelMode uint8

	// ModeExtension contains additional information for Joint Stereo and Dual Channel modes. It is a 2-bit field.
	ModeExtension uint8

	// CopyrightBit indicates whether the frame is copyrighted. A value of 1 means the frame is copyrighted, while 0 means it is not.
	CopyrightBit bool

	// OriginalBit indicates whether the frame is an original or a copy. A value of 1 means the frame is original, while 0 means it is a copy.
	OriginalBit bool

	// Emphasis is the de-emphasis applied to the frame. Possible values are 0 (None), 1 (50/15 microseconds), and 3 (CCITT J.17).
	Emphasis uint8
}

type MP3SideInfo struct {
	// MainDataBegin is a 9-bit value (MPEG-1) or 8-bit value (MPEG-2) that specifies the number of bytes in the main_data() field.
	MainDataBegin uint16

	// PrivateBits is a private bit field that has no specified meaning and can be used for application-specific purposes. For MPEG-1: 5 bits (3 bits in mono), for MPEG-2: 3 bits (1 bit in mono).
	PrivateBits uint8

	// ChannelData contains an array of ChannelData structs with side info for each channel.
	ChannelData [2][2]ChannelData // [granule][channel]
}

type ChannelData struct {
	// Part2_3Length is a 12-bit value that specifies the number of bits used for scalefactors and Huffman data in the main_data() field.
	Part2_3Length uint16

	// BigValues is a 9-bit value that specifies the number of Huffman code words whose length is <= 16 bits (two region boundary).
	BigValues uint16

	// GlobalGain is an 8-bit value that specifies the global gain factor for the granule/channel combination.
	GlobalGain uint8

	// ScalefacCompress is a 4-bit (MPEG-1) or 9-bit (MPEG-2) value that specifies the method of Huffman code table selection.
	ScalefacCompress uint16

	// WindowSwitchingFlag is a 1-bit value that specifies whether window switching is enabled for the granule/channel combination.
	WindowSwitchingFlag bool

	// BlockType is a 2-bit value that specifies the block type for the granule/channel combination when window switching is enabled.
	BlockType uint8

	// MixedBlockFlag is a 1-bit value that specifies whether the block is a mixed block type when window switching is enabled.
	MixedBlockFlag bool

	// TableSelect contains an array of 3 5-bit values that specify the Huffman code table numbers for the big values region.
	TableSelect [3]uint8

	// SubblockGain contains an array of 3 3-bit values that specify the subblock gains for the short windows when window switching is enabled.
	SubblockGain [3]uint8

	// Region0Count is a 4-bit value that specifies the number of Huffman code words in the first region.
	Region0Count uint8

	// Region1Count is a 3-bit value that specifies the number of Huffman code words in the second region.
	Region1Count uint8

	// PreFlag is a 1-bit value that specifies whether the pre-emphasis flag is set for the granule/channel combination (MPEG-1 only).
	PreFlag bool

	// ScalefacScale is a 1-bit value that specifies the scale for the scalefactors.
	ScalefacScale uint8

	// Count1TableSelect is a 1-bit value that specifies the Huffman code table number for the count1 region.
	Count1TableSelect uint8
}

func UnmarshalMP3Frames(r io.Reader) ([]*MP3Frame, error) {
	var frames []*MP3Frame

	for {
		header, err := readMP3Header(r)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		sideInfo, err := readMP3SideInfo(r, header)
		if err != nil {
			return nil, err
		}

		// Decode main data, ancillary data, and other fields in the frame
		// ...

		frame := &MP3Frame{
			Header:   *header,
			SideInfo: *sideInfo,
			// Populate MainData and AncillaryData once decoded
		}
		frames = append(frames, frame)
	}

	return frames, nil
}

func readMP3Header(r io.Reader) (*MP3Header, error) {
	var headerData [4]byte
	if _, err := io.ReadFull(r, headerData[:]); err != nil {
		return nil, err
	}

	header := &MP3Header{}

	// Check for the sync word (0x7FF)
	if headerData[0]&0xFF != 0xFF || headerData[1]&0xE0 != 0xE0 {
		return nil, errors.New("invalid sync word in MP3 header")
	}
	header.SyncWord = 0x7FF

	// Parse the rest of the header fields
	header.MPEGVersion = (headerData[1] >> 3) & 0x01
	header.Layer = (headerData[1] >> 1) & 0x03
	header.ProtectionBit = (headerData[1] & 0x01) == 0
	header.BitrateIndex = (headerData[2] >> 4) & 0x0F
	header.SamplingRateIndex = (headerData[2] >> 2) & 0x03
	header.PaddingBit = (headerData[2] & 0x02) != 0
	header.PrivateBit = (headerData[2] & 0x01) != 0
	header.ChannelMode = (headerData[3] >> 6) & 0x03
	header.ModeExtension = (headerData[3] >> 4) & 0x03
	header.CopyrightBit = (headerData[3] & 0x08) != 0
	header.OriginalBit = (headerData[3] & 0x04) != 0
	header.Emphasis = headerData[3] & 0x03

	return header, nil
}

func readMP3SideInfo(r io.Reader, header *MP3Header) (*MP3SideInfo, error) {
	return &MP3SideInfo{}, nil
}
