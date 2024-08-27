package tts

import (
	"bytes"
	"discord-bot/utils"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"

	"github.com/jonas747/ogg"
)

type dca struct {
	pipReader  io.Reader
	framesChan chan []byte
}

func encodeMem(r io.Reader) *dca {
	DCA := dca{
		pipReader:  r,
		framesChan: make(chan []byte, 100),
	}

	go DCA.run()

	return &DCA
}

func (dca *dca) run() (err error) {
	ffmpeg := exec.Command("ffmpeg",
		"-stats",       // Show progress/statistics during processing
		"-i", "pipe:0", // Input from stdin (pipe:0)
		"-reconnect", "1", // Enable automatic reconnection for network streams
		"-reconnect_at_eof", "1", // Reconnect if the stream ends (useful for live streams)
		"-reconnect_streamed", "1", // Reconnect even if the stream is already started
		"-reconnect_delay_max", "2", // Maximum delay between reconnection attempts in seconds
		"-map", "0:a", // Select the audio stream from the input (0:a means first input, audio stream)
		"-acodec", "libopus", // Encode audio using the Opus codec
		"-f", "ogg", // Set the output format to Ogg (Opus requires Ogg container)
		"-vbr", "on", // Enable variable bitrate (VBR) for better quality
		"-compression_level", "10", // Set the Opus compression level (10 is the highest compression)
		"-ar", "48000", // Set the audio sample rate to 48 kHz (standard for Opus)
		"-ac", "2", // Set the number of audio channels to 2 (stereo)
		"-b:a", "64000", // Set the target bitrate to 64 kbps
		"-application", "audio", // Set the Opus application mode to "audio" (optimized for music/voice)
		"-frame_duration", "20", // Set the duration of each Opus frame to 20 ms
		"-packet_loss", "1", // Set packet loss percentage; 1% is typical for handling network issues
		"-threads", "0", // Use the default number of threads (0 lets ffmpeg decide)
		"-ss", "0", // Start at the beginning of the input (0 seconds)
		"pipe:1", // Output to stdout (pipe:1)
	)

	ffmpeg.Stdin = dca.pipReader
	ffmpeg.Stderr = io.Discard // ignore stderr

	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		return
	}

	if err = ffmpeg.Start(); err != nil {
		return
	}

	// read stdout
	defer close(dca.framesChan)
	dca.readStdout(stdout)

	if err = ffmpeg.Wait(); err != nil {
		return
	}

	return
}

// opusFrame implements OpusReader, returning the next opus frame
func (dca *dca) opusFrame() ([]byte, error) {
	frame := <-dca.framesChan
	if frame == nil {
		return nil, io.EOF
	}

	if len(frame) < 2 {
		return nil, fmt.Errorf("invalid frame length: %d", len(frame))
	}

	return frame[2:], nil
}

func (dca *dca) readStdout(stdout io.ReadCloser) {
	decoder := ogg.NewPacketDecoder(ogg.NewDecoder(stdout))

	// the first 2 packets are ogg opus metadata
	skipPackets := 2
	for {
		// Retrieve a packet
		packet, _, err := decoder.Decode()
		if skipPackets > 0 {
			skipPackets--
			continue
		}
		if err != nil {
			if err != io.EOF {
				utils.Log.Error("\nreading ffmpeg stdout:", err.Error())
			}
			break
		}

		err = dca.writeOpusFrame(packet)
		if err != nil {
			utils.Log.Error("\nwriting opus frame:", err.Error())
			break
		}
	}
}

func (dca *dca) writeOpusFrame(opusFrame []byte) error {
	var dcaBuf bytes.Buffer

	err := binary.Write(&dcaBuf, binary.LittleEndian, int16(len(opusFrame)))
	if err != nil {
		return err
	}

	_, err = dcaBuf.Write(opusFrame)
	if err != nil {
		return err
	}

	dca.framesChan <- dcaBuf.Bytes()

	return nil
}
