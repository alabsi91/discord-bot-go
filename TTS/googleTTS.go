package tts

import (
	"discord-bot/utils"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

type TTSOptions struct {
	Lang     string
	Host     string
	Slow     bool
	FilePath string
}

// GenerateAndSaveToFile generates TTS audio from the provided text and saves it to a file.
func GenerateAndSaveToFile(text string, options ...TTSOptions) error {
	option, err := validateOptions(text, options...)
	if err != nil {
		return err
	}

	url := getAudioUrl(text, option)

	audioData, err := downloadAudioData(url)
	if err != nil {
		return err
	}
	defer (audioData).Close()

	err = saveAudioToFile(audioData, option.FilePath)
	if err != nil {
		return err
	}

	return nil
}

// GenerateAndSendToVoiceChannel generates TTS audio from the provided text and sends it to the specified Discord voice channel.
func GenerateAndSendToVoiceChannel(text string, voiceConnection *discordgo.VoiceConnection, options ...TTSOptions) error {
	option, err := validateOptions(text, options...)
	if err != nil {
		return err
	}

	url := getAudioUrl(text, option)

	audioData, err := downloadAudioData(url)
	if err != nil {
		return err
	}
	defer (audioData).Close()

	dca := encodeMem(audioData)

	for {
		frame, err := dca.opusFrame()
		if err == io.EOF {
			break
		}
		if err != nil {
			utils.Log.Error("Error getting Opus frame:", err.Error())
			break
		}

		select {
		case voiceConnection.OpusSend <- frame:
		case <-time.After(time.Second):
			return errors.New("failed to send frame in time")
		}
	}

	return nil
}

// validateOptions validates and sets default options for TTS generation.
func validateOptions(text string, options ...TTSOptions) (TTSOptions, error) {
	defaultOptions := TTSOptions{
		Lang:     "en-us",
		Host:     "https://translate.google.com",
		Slow:     false,
		FilePath: "output.mp3",
	}

	if len(options) > 0 {
		if options[0].Lang != "" {
			defaultOptions.Lang = options[0].Lang
		}
		if options[0].Host != "" {
			defaultOptions.Host = options[0].Host
		}
		if options[0].Slow {
			defaultOptions.Slow = options[0].Slow
		}
		if options[0].FilePath != "" {
			defaultOptions.FilePath = options[0].FilePath
		}
	}

	if text == "" {
		return defaultOptions, errors.New("text should be a string")
	}
	if len(text) > 200 {
		return defaultOptions, errors.New("text should be less than 200 characters")
	}

	return defaultOptions, nil
}

// GetAudioUrl constructs the Google Translate TTS URL based on the provided options and text.
func getAudioUrl(text string, options TTSOptions) string {
	speed := "1.0"
	if options.Slow {
		speed = "0.24"
	}

	query := url.Values{}
	query.Set("ie", "UTF-8")
	query.Set("q", text)
	query.Set("tl", options.Lang)
	query.Set("total", "1")
	query.Set("idx", "0")
	query.Set("textlen", strconv.Itoa(len(text)))
	query.Set("client", "tw-ob")
	query.Set("prev", "input")
	query.Set("ttsspeed", speed)

	return options.Host + "/translate_tts?" + query.Encode()
}

// DownloadAudioData downloads the audio data from the specified URL.
func downloadAudioData(audioURL string) (io.ReadCloser, error) {
	resp, err := http.Get(audioURL)
	if err != nil {
		return resp.Body, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return resp.Body, fmt.Errorf("failed to download file: %s", resp.Status)
	}

	return resp.Body, nil
}

// SaveAudioToFile saves the downloaded audio data to the specified file path.
func saveAudioToFile(audioData io.ReadCloser, filePath string) error {
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, audioData)
	return err
}
