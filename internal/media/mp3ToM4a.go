package media

import (
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/xfrr/goffmpeg/transcoder"
)

func Mp3ToM4A(mp3Reader io.Reader, out io.Writer) error {
	// Read the MP3 data into a buffer.
	mp3Data, err := ioutil.ReadAll(mp3Reader)
	if err != nil {
		return errors.Wrap(err, "failed to read MP3 data")
	}

	// Set up the transcoder.
	trans := new(transcoder.Transcoder)

	// Create the input pipe for the MP3 data.
	inputPipe, err := trans.CreateInputPipe()
	if err != nil {
		return errors.Wrap(err, "failed to create input pipe")
	}

	// Create the output pipe for the M4A data.
	outputPipe, err := trans.CreateOutputPipe("m4a")
	if err != nil {
		return errors.Wrap(err, "failed to create output pipe")
	}

	// Set up the conversion process.
	mp3WriteDone := make(chan error)
	go func() {
		defer close(mp3WriteDone)

		// Write the MP3 data to the input pipe.
		if _, err := inputPipe.Write(mp3Data); err != nil {
			mp3WriteDone <- err
			return
		}
		if err := inputPipe.Close(); err != nil {
			mp3WriteDone <- err
			return
		}
	}()

	transcodeDone := trans.Run(false)
	err = <-transcodeDone
	if err != nil {
		return errors.Wrap(err, "failed to transcode MP3 to M4A")
	}

	// Wait for the writing process to finish.
	if err = <-mp3WriteDone; err != nil {
		return errors.Wrap(err, "failed to write MP3 data to input pipe")
	}

	// Read the M4A data from the output pipe.
	_, err = io.Copy(out, outputPipe)
	if err != nil {
		return errors.Wrap(err, "failed to copy M4A data from output pipe")
	}

	return nil
}
