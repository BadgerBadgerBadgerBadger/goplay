package media

import (
	"bytes"
	"encoding/binary"
	"io"
	"time"

	"github.com/pkg/errors"
)

const (
	MajorBrandM4b = "M4B"

	CompatibleBrandIsom = "isom"
	CompatibleBrandM4B  = "M4B"
	CompatibleBrandMP42 = "MP42"
)

// FileTypeBox (ftyp) - Identifies the file type and compatibility.
type FileTypeBox struct {
	MajorBrand       string   // Major brand for the file format.
	MinorVersion     uint32   // Minor version for the major brand.
	CompatibleBrands []string // List of compatible brands.
}

// MovieBox (moov) - Contains metadata about the overall media file.
type MovieBox struct {
	MovieHeaderBox *MovieHeaderBox
	TrackBox       *TrackBox
	UserDataBox    *UserDataBox
}

// MovieHeaderBox (mvhd) - Contains information about the overall media, such as duration and timescale.
type MovieHeaderBox struct {
	CreationTime     uint32 // Creation time of the media.
	ModificationTime uint32 // Modification time of the media.
	Timescale        uint32 // Timescale used for the media.
	Duration         uint32 // Duration of the media in timescale units.
}

// TrackBox (trak) - Contains information about the audio track.
type TrackBox struct {
	TrackHeaderBox TrackHeaderBox
	MediaBox       MediaBox
}

// TrackHeaderBox (tkhd) - Contains information about the track, such as track ID and duration.
type TrackHeaderBox struct {
	Version          uint8     // Version of the 'tkhd' box, either 0 or 1.
	Flags            uint32    // Flags defining specific properties of the track (e.g., enabled or disabled).
	CreationTime     time.Time // The creation time of the track.
	ModificationTime time.Time // The modification time of the track.
	TrackID          uint32    // A unique identifier for the track within the file.
	Reserved1        uint32    // Reserved for future use, should be set to 0.
	Duration         uint64    // The duration of the track in the movie's timescale units.
	Reserved2        uint64    // Reserved for future use, should be set to 0.
	Layer            int16     // The front-to-back ordering of video tracks; lower numbers are closer to the viewer. For audio tracks, typically set to 0.
	AlternateGroup   int16     // Used to group alternate tracks, such as tracks in different languages. A value of 0 means the track is not part of any group.
	Volume           int16     // The track's audio volume, as a fixed-point 8.8 number. For video tracks, set to 0.
	Reserved3        int16     // Reserved for future use, should be set to 0.
	Matrix           [9]int32  // A 3x3 transformation matrix used to modify the visual presentation of the track. For audio tracks, typically set to the identity matrix.
	Width            uint32    // The track's visual width in pixels, as a fixed-point 16.16 number. For audio tracks, set to 0.
	Height           uint32    // The track's visual height in pixels, as a fixed-point 16.16 number. For audio tracks, set to 0.
}

// MarshalBinary implements the encoding.BinaryMarshaler interface for TkhdBox.
func (t *TrackHeaderBox) MarshalBinary() ([]byte, error) {
	if t == nil {
		return nil, errors.New("TkhdBox: cannot marshal a nil TkhdBox")
	}

	buf := &bytes.Buffer{}

	// Write the size placeholder (4 bytes), which will be updated later.
	_ = binary.Write(buf, binary.BigEndian, uint32(0))

	// Write the box type 'tkhd' (4 bytes).
	_, _ = buf.WriteString("tkhd")

	// Write the version (1 byte).
	_ = binary.Write(buf, binary.BigEndian, t.Version)

	// Write the flags (3 bytes).
	_ = binary.Write(buf, binary.BigEndian, t.Flags&0xFFFFFF)

	// Check the version to determine the date encoding format.
	if t.Version == 1 {
		// Write the creation and modification times as 64-bit unsigned integers (8 bytes each).
		_ = binary.Write(buf, binary.BigEndian, uint64(t.CreationTime.Unix()))
		_ = binary.Write(buf, binary.BigEndian, uint64(t.ModificationTime.Unix()))

		// Write the track ID, reserved1, and duration (4 bytes each).
		_ = binary.Write(buf, binary.BigEndian, t.TrackID)
		_ = binary.Write(buf, binary.BigEndian, t.Reserved1)
		_ = binary.Write(buf, binary.BigEndian, t.Duration)
	} else {
		// Write the creation and modification times as 32-bit unsigned integers (4 bytes each).
		_ = binary.Write(buf, binary.BigEndian, uint32(t.CreationTime.Unix()))
		_ = binary.Write(buf, binary.BigEndian, uint32(t.ModificationTime.Unix()))

		// Write the track ID, reserved1, and duration (4 bytes each).
		_ = binary.Write(buf, binary.BigEndian, t.TrackID)
		_ = binary.Write(buf, binary.BigEndian, t.Reserved1)
		_ = binary.Write(buf, binary.BigEndian, uint32(t.Duration))
	}

	// Write the reserved2, layer, alternate group, volume, and reserved3 fields.
	_ = binary.Write(buf, binary.BigEndian, t.Reserved2)
	_ = binary.Write(buf, binary.BigEndian, t.Layer)
	_ = binary.Write(buf, binary.BigEndian, t.AlternateGroup)
	_ = binary.Write(buf, binary.BigEndian, t.Volume)
	_ = binary.Write(buf, binary.BigEndian, t.Reserved3)

	// Write the transformation matrix.
	for _, v := range t.Matrix {
		_ = binary.Write(buf, binary.BigEndian, v)
	}

	// Write the width and height (fixed-point 16.16 numbers).
	_ = binary.Write(buf, binary.BigEndian, t.Width)
	_ = binary.Write(buf, binary.BigEndian, t.Height)

	// Update the size at the beginning of the buffer.
	boxSize := uint32(buf.Len())
	binary.BigEndian.PutUint32(buf.Bytes()[:4], boxSize)

	return buf.Bytes(), nil
}

// MediaBox (mdia) - Contains information about the media data in the track.
type MediaBox struct {
	MediaHeaderBox      MediaHeaderBox
	HandlerReferenceBox HandlerReferenceBox
	MediaInformationBox *MediaInformationBox
}

func (m MediaBox) MarshalBinary() ([]byte, error) {
	// Marshal Media Header Box (mdhd)
	mdhdData, err := m.MediaHeaderBox.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// Marshal Handler Reference Box (hdlr)
	hdlrData, err := m.HandlerReferenceBox.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// Marshal Media Information Box (minf)
	minfData, err := m.MediaInformationBox.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// Calculate the size of the Media Box, including the 8-byte header
	size := uint32(8) + uint32(len(mdhdData)+len(hdlrData)+len(minfData))

	// Create the Media Box header
	header := make([]byte, 8)
	binary.BigEndian.PutUint32(header[:4], size)
	copy(header[4:], "mdia")

	// Concatenate the header and child atom data
	data := append(header, mdhdData...)
	data = append(data, hdlrData...)
	data = append(data, minfData...)

	return data, nil
}

// MediaHeaderBox (mdhd) - Contains information about the media data, such as duration and timescale.
type MediaHeaderBox struct {
	Version          uint8   // Version of the media header box
	Flags            [3]byte // Flags (set to zero)
	CreationTime     uint64  // Creation time of the media data in seconds since midnight, Jan 1, 1904 (32-bit for version 0, 64-bit for version 1)
	ModificationTime uint64  // Modification time of the media data in seconds since midnight, Jan 1, 1904 (32-bit for version 0, 64-bit for version 1)
	Timescale        uint32  // Number of time units that pass per second
	Duration         uint64  // Length of the media in the defined timescale units (32-bit for version 0, 64-bit for version 1)
	Language         [2]byte // 16-bit packed ISO-639-2 language code
	Quality          uint16  // Reserved for use by Apple (set to zero)
}

func (m MediaHeaderBox) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write the size placeholder (4 bytes, to be updated later)
	if err := binary.Write(buf, binary.BigEndian, uint32(0)); err != nil {
		return nil, err
	}

	// Write the Media Header Box type 'mdhd' (4 bytes)
	if _, err := buf.WriteString("mdhd"); err != nil {
		return nil, err
	}

	// Write the version (1 byte)
	if err := buf.WriteByte(m.Version); err != nil {
		return nil, err
	}

	// Write the flags (3 bytes)
	if _, err := buf.Write(m.Flags[:]); err != nil {
		return nil, err
	}

	// Write the creation time (4 bytes for version 0, 8 bytes for version 1)
	if m.Version == 1 {
		if err := binary.Write(buf, binary.BigEndian, m.CreationTime); err != nil {
			return nil, err
		}
	} else {
		if err := binary.Write(buf, binary.BigEndian, uint32(m.CreationTime)); err != nil {
			return nil, err
		}
	}

	// Write the modification time (4 bytes for version 0, 8 bytes for version 1)
	if m.Version == 1 {
		if err := binary.Write(buf, binary.BigEndian, m.ModificationTime); err != nil {
			return nil, err
		}
	} else {
		if err := binary.Write(buf, binary.BigEndian, uint32(m.ModificationTime)); err != nil {
			return nil, err
		}
	}

	// Write the timescale (4 bytes)
	if err := binary.Write(buf, binary.BigEndian, m.Timescale); err != nil {
		return nil, err
	}

	// Write the duration (4 bytes for version 0, 8 bytes for version 1)
	if m.Version == 1 {
		if err := binary.Write(buf, binary.BigEndian, m.Duration); err != nil {
			return nil, err
		}
	} else {
		if err := binary.Write(buf, binary.BigEndian, uint32(m.Duration)); err != nil {
			return nil, err
		}
	}

	// Write the language code (2 bytes)
	if _, err := buf.Write(m.Language[:]); err != nil {
		return nil, err
	}

	// Write the quality (2 bytes)
	if err := binary.Write(buf, binary.BigEndian, m.Quality); err != nil {
		return nil, err
	}

	// Update the size field with the actual size of the box
	boxSize := uint32(buf.Len())
	binary.BigEndian.PutUint32(buf.Bytes()[0:4], boxSize)

	return buf.Bytes(), nil
}

// HandlerReferenceBox (hdlr) - Specifies the handler type (e.g., audio or video) and the name of the handler.
type HandlerReferenceBox struct {
	HandlerType [4]byte // Handler type, 'soun' for audio tracks
	Name        string  // Name of the handler (optional)
}

func (h HandlerReferenceBox) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	// Write the size (4 bytes, placeholder for now)
	if _, err := buf.Write([]byte{0, 0, 0, 0}); err != nil {
		return nil, err
	}

	// Write the type 'hdlr' (4 bytes)
	if _, err := buf.Write([]byte("hdlr")); err != nil {
		return nil, err
	}

	// Write version and flags (4 bytes: 1-byte version, 3-byte flags)
	if _, err := buf.Write([]byte{0, 0, 0, 0}); err != nil {
		return nil, err
	}

	// Write the pre_defined field (4 bytes)
	if err := binary.Write(buf, binary.BigEndian, uint32(0)); err != nil {
		return nil, err
	}

	// Write the handler_type field (4 bytes)
	if _, err := buf.Write(h.HandlerType[:]); err != nil {
		return nil, err
	}

	// Write the reserved fields (12 bytes, all set to zero)
	reserved := [12]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if _, err := buf.Write(reserved[:]); err != nil {
		return nil, err
	}

	// Write the name field (null-terminated string)
	if _, err := io.WriteString(buf, h.Name); err != nil {
		return nil, err
	}
	if err := buf.WriteByte(0); err != nil { // Null-terminate the string
		return nil, err
	}

	// Update the size field with the actual size of the box
	boxSize := uint32(buf.Len())
	binary.BigEndian.PutUint32(buf.Bytes()[0:4], boxSize)

	return buf.Bytes(), nil
}

// MediaInformationBox (minf) - Contains information about the media data storage.
type MediaInformationBox struct {
	SoundMediaHeaderBox *SoundMediaHeaderBox
	DataInformationBox  *DataInformationBox
	SampleTableBox      *SampleTableBox
}

func (m MediaInformationBox) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	// Write the size (4 bytes, placeholder for now)
	if _, err := buf.Write([]byte{0, 0, 0, 0}); err != nil {
		return nil, err
	}

	// Write the type 'minf' (4 bytes)
	if _, err := buf.Write([]byte("minf")); err != nil {
		return nil, err
	}

	// Write the Sound Media Header Box (smhd)
	if m.SoundMediaHeaderBox != nil {
		smhdData, err := m.SoundMediaHeaderBox.MarshalBinary()
		if err != nil {
			return nil, err
		}
		if _, err := buf.Write(smhdData); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("SoundMediaHeaderBox is required")
	}

	// Write the Data Information Box (dinf)
	if m.DataInformationBox != nil {
		dinfData, err := m.DataInformationBox.MarshalBinary()
		if err != nil {
			return nil, err
		}
		if _, err := buf.Write(dinfData); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("DataInformationBox is required")
	}

	// Write the Sample Table Box (stbl)
	if m.SampleTableBox != nil {
		stblData, err := m.SampleTableBox.MarshalBinary()
		if err != nil {
			return nil, err
		}
		if _, err := buf.Write(stblData); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("SampleTableBox is required")
	}

	// Update the size field with the actual size of the box
	boxSize := uint32(buf.Len())
	binary.BigEndian.PutUint32(buf.Bytes()[0:4], boxSize)

	return buf.Bytes(), nil
}

// SoundMediaHeaderBox (smhd) - Contains audio-specific information.
type SoundMediaHeaderBox struct {
	Version  byte
	Flags    [3]byte
	Balance  int16
	Reserved int16
}

func (s SoundMediaHeaderBox) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	// Write the size (4 bytes, placeholder for now)
	if _, err := buf.Write([]byte{0, 0, 0, 0}); err != nil {
		return nil, err
	}

	// Write the type 'smhd' (4 bytes)
	if _, err := buf.Write([]byte("smhd")); err != nil {
		return nil, err
	}

	// Write the version (1 byte)
	if err := buf.WriteByte(s.Version); err != nil {
		return nil, err
	}

	// Write the flags (3 bytes)
	if _, err := buf.Write(s.Flags[:]); err != nil {
		return nil, err
	}

	// Write the balance (2 bytes)
	if err := binary.Write(buf, binary.BigEndian, s.Balance); err != nil {
		return nil, err
	}

	// Write the reserved field (2 bytes)
	if err := binary.Write(buf, binary.BigEndian, s.Reserved); err != nil {
		return nil, err
	}

	// Update the size field with the actual size of the box
	boxSize := uint32(buf.Len())
	binary.BigEndian.PutUint32(buf.Bytes()[0:4], boxSize)

	return buf.Bytes(), nil
}

// DataInformationBox (dinf) - Contains data reference information for the media.
type DataInformationBox struct {
	DataReferenceBox *DataReferenceBox
}

func (d DataInformationBox) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	// Write the size (4 bytes, placeholder for now)
	if _, err := buf.Write([]byte{0, 0, 0, 0}); err != nil {
		return nil, err
	}

	// Write the type 'dinf' (4 bytes)
	if _, err := buf.Write([]byte("dinf")); err != nil {
		return nil, err
	}

	// Marshal the DataReferenceBox and write its content
	drefData, err := d.DataReferenceBox.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if _, err := buf.Write(drefData); err != nil {
		return nil, err
	}

	// Update the size field with the actual size of the box
	boxSize := uint32(buf.Len())
	binary.BigEndian.PutUint32(buf.Bytes()[0:4], boxSize)

	return buf.Bytes(), nil
}

// DataReferenceBox (dref) - Contains data references for the media.
type DataReferenceBox struct {
	EntryCount uint32          // Number of data entries.
	Entries    []*DataEntryBox // List of data entries.
}

func (d DataReferenceBox) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	// Write the size (4 bytes, placeholder for now)
	if _, err := buf.Write([]byte{0, 0, 0, 0}); err != nil {
		return nil, err
	}

	// Write the type 'dref' (4 bytes)
	if _, err := buf.Write([]byte("dref")); err != nil {
		return nil, err
	}

	// Write the version and flags (4 bytes: 1 byte for version, 3 bytes for flags)
	if _, err := buf.Write([]byte{0, 0, 0, 0}); err != nil {
		return nil, err
	}

	// Write the entry count (4 bytes)
	if err := binary.Write(buf, binary.BigEndian, d.EntryCount); err != nil {
		return nil, err
	}

	// Marshal the DataEntryUrlBox(es) and write their content
	for _, entry := range d.Entries {
		entryData, err := entry.MarshalBinary()
		if err != nil {
			return nil, err
		}
		if _, err := buf.Write(entryData); err != nil {
			return nil, err
		}
	}

	// Update the size field with the actual size of the box
	boxSize := uint32(buf.Len())
	binary.BigEndian.PutUint32(buf.Bytes()[0:4], boxSize)

	return buf.Bytes(), nil
}

// DataEntryBox - Abstract box for data entry, usually a UrlBox or an AliasBox.
type DataEntryBox struct {
	Version  uint8
	Flags    [3]byte
	Location string // Empty string for self-contained media data
}

func (d DataEntryBox) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	// Write the size (4 bytes, placeholder for now)
	if _, err := buf.Write([]byte{0, 0, 0, 0}); err != nil {
		return nil, err
	}

	// Write the type 'url ' (4 bytes)
	if _, err := buf.Write([]byte("url ")); err != nil {
		return nil, err
	}

	// Write the version and flags (4 bytes: 1 byte for version, 3 bytes for flags)
	if _, err := buf.Write(append([]byte{d.Version}, d.Flags[:]...)); err != nil {
		return nil, err
	}

	// Write the location (null-terminated string)
	if _, err := buf.WriteString(d.Location + "\x00"); err != nil {
		return nil, err
	}

	// Update the size field with the actual size of the box
	boxSize := uint32(buf.Len())
	binary.BigEndian.PutUint32(buf.Bytes()[0:4], boxSize)

	return buf.Bytes(), nil
}

// SampleTableBox (stbl) - Contains information about the samples (audio frames) in the media data.
type SampleTableBox struct {
	SampleDescriptionBox *SampleDescriptionBox
	TimeToSampleBox      *TimeToSampleBox
	SampleSizeBox        *SampleSizeBox
	SampleToChunkBox     *SampleToChunkBox
	ChunkOffsetBox       *ChunkOffsetBox
}

func (box *SampleTableBox) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	stsd, err := box.SampleDescriptionBox.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(stsd)

	stts, err := box.TimeToSampleBox.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(stts)

	stsc, err := box.SampleToChunkBox.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(stsc)

	stsz, err := box.SampleSizeBox.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(stsz)

	stco, err := box.ChunkOffsetBox.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(stco)

	return buf.Bytes(), nil
}

// SampleDescriptionBox (stsd) - Contains codec-specific information (MP3).
type SampleDescriptionBox struct {
	EntryCount uint32                 // Number of sample entries.
	Entries    []*AudioSampleEntryBox // List of audio sample entries.
}

func (box *SampleDescriptionBox) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	binary.Write(buf, binary.BigEndian, uint32(0)) // Placeholder for box size
	buf.WriteString("stsd")
	binary.Write(buf, binary.BigEndian, uint32(0)) // Version and flags
	binary.Write(buf, binary.BigEndian, box.EntryCount)

	for _, entry := range box.Entries {
		entryData, err := entry.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(entryData)
	}

	// Update box size
	boxSize := uint32(buf.Len())
	binary.BigEndian.PutUint32(buf.Bytes(), boxSize)

	return buf.Bytes(), nil
}

type AudioSampleEntryBox struct {
	// Common fields for all sample entries
	DataFormat         [4]byte // Format type of the sample (e.g., 'mp4a')
	Reserved           [6]byte // Reserved, set to 0
	DataReferenceIndex uint16  // Indicates the data reference to use to retrieve the media data

	// Audio-specific fields
	Version          uint16  // Format version
	RevisionLevel    uint16  // Format revision level
	Vendor           [4]byte // Vendor identifier
	NumberOfChannels uint16  // Number of audio channels
	SampleSize       uint16  // Number of bits per audio sample
	CompressionID    uint16  // Compression scheme identifier (0 for version 0)
	PacketSize       uint16  // Reserved, set to 0
	SampleRate       uint32  // Audio sample rate, fixed point 16.16 number
}

func (entry *AudioSampleEntryBox) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	buf.Write(entry.DataFormat[:])
	buf.Write(entry.Reserved[:])
	binary.Write(buf, binary.BigEndian, entry.DataReferenceIndex)

	binary.Write(buf, binary.BigEndian, entry.Version)
	binary.Write(buf, binary.BigEndian, entry.RevisionLevel)
	buf.Write(entry.Vendor[:])
	binary.Write(buf, binary.BigEndian, entry.NumberOfChannels)
	binary.Write(buf, binary.BigEndian, entry.SampleSize)
	binary.Write(buf, binary.BigEndian, entry.CompressionID)
	binary.Write(buf, binary.BigEndian, entry.PacketSize)
	binary.Write(buf, binary.BigEndian, entry.SampleRate)

	return buf.Bytes(), nil
}

// CodecSpecificBox - Contains codec-specific information.
type CodecSpecificBox struct {
	Data []byte // Raw codec-specific data.
}

// TimeToSampleBox (stts) - Contains information about the timing of the samples.
type TimeToSampleBox struct {
	EntryCount uint32
	Entries    []*TimeToSampleEntry
}

func (box *TimeToSampleBox) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	binary.Write(buf, binary.BigEndian, uint32(0)) // Placeholder for box size
	buf.WriteString("stts")
	binary.Write(buf, binary.BigEndian, uint32(0)) // Version and flags
	binary.Write(buf, binary.BigEndian, box.EntryCount)

	for _, entry := range box.Entries {
		binary.Write(buf, binary.BigEndian, entry.SampleCount)
		binary.Write(buf, binary.BigEndian, entry.SampleDelta)
	}

	// Update box size
	boxSize := uint32(buf.Len())
	binary.BigEndian.PutUint32(buf.Bytes(), boxSize)

	return buf.Bytes(), nil
}

// TimeToSampleEntry - Represents a single entry in a TimeToSampleBox.
type TimeToSampleEntry struct {
	SampleCount uint32 // Number of consecutive samples with the same duration.
	SampleDelta uint32 // Duration of the samples in the entry, in timescale units.
}

// SampleSizeBox (stsz) - Contains the size of each sample.
type SampleSizeBox struct {
	SampleSize      uint32   // Uniform size of all samples, or 0 if samples have different sizes.
	SampleCount     uint32   // Number of samples.
	SampleSizeTable []uint32 // Sizes of individual samples, if SampleSize is 0.
}

func (box *SampleSizeBox) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	binary.Write(buf, binary.BigEndian, uint32(0)) // Placeholder for box size
	buf.WriteString("stsz")
	binary.Write(buf, binary.BigEndian, uint32(0)) // Version and flags
	binary.Write(buf, binary.BigEndian, box.SampleSize)
	binary.Write(buf, binary.BigEndian, box.SampleCount)

	if box.SampleSize == 0 {
		for _, sampleSize := range box.SampleSizeTable {
			binary.Write(buf, binary.BigEndian, sampleSize)
		}
	}

	// Update box size
	boxSize := uint32(buf.Len())
	binary.BigEndian.PutUint32(buf.Bytes(), boxSize)

	return buf.Bytes(), nil
}

// SampleToChunkBox (stsc) - Contains information about how samples are stored in chunks.
type SampleToChunkBox struct {
	EntryCount uint32
	Entries    []*SampleToChunkEntry
}

func (box *SampleToChunkBox) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	binary.Write(buf, binary.BigEndian, uint32(0)) // Placeholder for box size
	buf.WriteString("stsc")
	binary.Write(buf, binary.BigEndian, uint32(0)) // Version and flags
	binary.Write(buf, binary.BigEndian, box.EntryCount)

	for _, entry := range box.Entries {
		binary.Write(buf, binary.BigEndian, entry.FirstChunk)
		binary.Write(buf, binary.BigEndian, entry.SamplesPerChunk)
		binary.Write(buf, binary.BigEndian, entry.SampleDescriptionIndex)
	}

	// Update box size
	boxSize := uint32(buf.Len())
	binary.BigEndian.PutUint32(buf.Bytes(), boxSize)

	return buf.Bytes(), nil
}

// SampleToChunkEntry - Represents a single entry in a SampleToChunkBox.
type SampleToChunkEntry struct {
	FirstChunk             uint32 // Index of the first chunk in this run of chunks with the same number of samples.
	SamplesPerChunk        uint32 // Number of samples in each chunk of this run.
	SampleDescriptionIndex uint32 // Index of the sample description for the samples in this run of chunks.
}

// ChunkOffsetBox (stco) - Contains the byte offset of each chunk in the media data.
type ChunkOffsetBox struct {
	EntryCount   uint32   // Number of chunk offsets.
	ChunkOffsets []uint32 // Byte offsets of individual chunks.
}

func (box *ChunkOffsetBox) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	binary.Write(buf, binary.BigEndian, uint32(0)) // Placeholder for box size
	buf.WriteString("stco")
	binary.Write(buf, binary.BigEndian, uint32(0)) // Version and flags
	binary.Write(buf, binary.BigEndian, box.EntryCount)

	for _, chunkOffset := range box.ChunkOffsets {
		binary.Write(buf, binary.BigEndian, chunkOffset)
	}

	// Update box size
	boxSize := uint32(buf.Len())
	binary.BigEndian.PutUint32(buf.Bytes(), boxSize)

	return buf.Bytes(), nil
}

// UserDataBox (udta) - Contains user-defined metadata, such as title, author, album, and album art.
type UserDataBox struct {
	MetaBox *MetaBox
}

// MetaBox (meta) - Contains metadata for the file.
type MetaBox struct {
	HandlerBox  *HandlerReferenceBox
	ItemListBox *ItemListBox
}

// ItemListBox (ilst) - Contains the actual metadata items, such as title, author, album, and album art.
// Only one of these should be set.
type ItemListBox struct {
	Title    *TitleBox
	Author   *AuthorBox
	Album    *AlbumBox
	AlbumArt *AlbumArtBox
}

// TitleBox (©nam) - Contains the title of the audiobook.
type TitleBox struct {
	Data string // Title text.
}

// AuthorBox (©ART) - Contains the author of the audiobook.
type AuthorBox struct {
	Data string // Author text.
}

// AlbumBox (©alb) - Contains the album name.
type AlbumBox struct {
	Data string // Album name text.
}

// AlbumArtBox (covr) - Contains the album art image.
type AlbumArtBox struct {
	Data []byte // Binary data of the album art image (e.g., JPEG or PNG).
}

// MediaDataBox (mdat) - Contains the actual MP3 audio data.
type MediaDataBox struct {
	Data []byte // Raw MP3 audio data.
}
