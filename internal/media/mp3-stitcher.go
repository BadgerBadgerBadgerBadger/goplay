package media

import (
	"bytes"
	"fmt"
	"io"

	"github.com/bogem/id3v2"
	"github.com/dmulholl/mp3lib"
	"github.com/pkg/errors"
)

type Stitcher struct {
	tag     *id3v2.Tag
	readers []io.Reader
}

type StitcherOption func(*Stitcher)

func WithId3v2(tag *id3v2.Tag) StitcherOption {
	return func(s *Stitcher) {
		s.tag = tag
	}
}

func NewStitcher(options ...StitcherOption) *Stitcher {
	s := &Stitcher{}
	for _, option := range options {
		option(s)
	}
	return s
}

func (s *Stitcher) Add(r io.Reader) {
	s.readers = append(s.readers, r)
}

func (s *Stitcher) Stitch(w io.Writer) error {

	outbuffer := bytes.NewBuffer([]byte{})

	if s.tag != nil {
		if _, err := s.tag.WriteTo(outbuffer); err != nil {
			return err
		}
	}

	var totalFrames uint32
	var totalBytes uint32
	var totalFiles int
	var firstBitRate int
	var isVBR bool

	for _, r := range s.readers {

		isFirstFrame := true

		for {
			// Read the next frame from the input file.
			frame := mp3lib.NextFrame(r)
			if frame == nil {
				break
			}

			// If this is the first frame, record the bit rate.
			if isFirstFrame {
				isFirstFrame = false
				if mp3lib.IsXingHeader(frame) || mp3lib.IsVbriHeader(frame) {
					continue
				}
			}

			// If we detect more than one bitrate we'll need to add a VBR header to the output file.
			if firstBitRate == 0 {
				firstBitRate = frame.BitRate
			} else if frame.BitRate != firstBitRate {
				isVBR = true
			}

			// Write the frame to the io.Writer.
			_, err := outbuffer.Write(frame.RawBytes)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to write frame %d", totalFrames))
			}

			totalFrames += 1
			totalBytes += uint32(len(frame.RawBytes))
		}

		totalFiles += 1
	}

	// If we detected multiple bitrates, prepend a VBR header to the file.
	if isVBR {
		outbuffer = prependXingHeader(totalFrames, totalBytes, outbuffer)
	}

	_, err := outbuffer.WriteTo(w)
	if err != nil {
		return errors.Wrap(err, "failed to write output")
	}

	return nil
}

// prependXingHeader prepends a Xing header to the output file. This is necessary
// when the input files have different bitrates. The Xing header is a VBR header
// that is used to indicate that the file is VBR. It is also used to store the
// total number of frames and total number of bytes in the file. This is useful
// for players that want to know the total length of the file.
// We need to return a new buffer because the Xing header is written to the
// beginning of the file, and we need to return the entire file.
func prependXingHeader(totalFrames uint32, totalBytes uint32, sofar *bytes.Buffer) *bytes.Buffer {
	xingHeader := mp3lib.NewXingHeader(totalFrames, totalBytes)

	out := bytes.NewBuffer([]byte{})
	out.Write(xingHeader.RawBytes)
	out.Write(sofar.Bytes())

	return out
}
