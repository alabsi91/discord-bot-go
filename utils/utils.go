package utils

import (
	"discord-bot/common"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
)

// GetInteractionAuthor returns the author of the interaction
func GetInteractionAuthor(i *discordgo.Interaction) *discordgo.User {
	if i.Member != nil {
		return i.Member.User
	}
	return i.User
}

// IsMessageLatest Function to check if a message is the latest in the channel
func IsMessageLatest(s *discordgo.Session, channelID, messageIDToCheck string) (bool, error) {
	// Fetch the latest messages in the channel
	messages, err := s.ChannelMessages(channelID, 1, "", "", "")
	if err != nil {
		return false, err
	}

	if len(messages) == 0 {
		// No messages found in the channel
		return false, nil
	}

	// Check if the message ID to check is the latest one
	latestMessageID := messages[0].ID
	return latestMessageID == messageIDToCheck, nil
}

// RandomInt generates a random int between min and max
func RandomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

// Contains checks if a string is present in a slice of strings.
func Contains(arr *[]string, target string) bool {
	for _, item := range *arr {
		if item == target {
			return true
		}
	}
	return false
}

// CheckOptionStringValue check if a discord option has a string value (not nil and not empty)
func CheckOptionStringValue(options *discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	if options.Value == nil {
		return "", fmt.Errorf("please provide a value")
	}

	val := options.StringValue()
	if val == "" {
		return "", fmt.Errorf("please provide a value")
	}

	return val, nil
}

// ZipFolder compresses the specified folder into a zip file at the given destination path.
// func ZipFolder(sourceFolder, destinationZip string) error {
// 	zipFile, err := os.Create(destinationZip)
// 	if err != nil {
// 		return err
// 	}
// 	defer zipFile.Close()

// 	zipWriter := zip.NewWriter(zipFile)
// 	defer zipWriter.Close()

// 	err = filepath.Walk(sourceFolder, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		relativePath, err := filepath.Rel(sourceFolder, path)
// 		if err != nil {
// 			return err
// 		}

// 		if relativePath == "." {
// 			return nil
// 		}

// 		header, err := zip.FileInfoHeader(info)
// 		if err != nil {
// 			return err
// 		}
// 		header.Name = relativePath

// 		if info.IsDir() {
// 			header.Name += "/"
// 			_, err := zipWriter.CreateHeader(header)
// 			return err
// 		}

// 		writer, err := zipWriter.CreateHeader(header)
// 		if err != nil {
// 			return err
// 		}

// 		file, err := os.Open(path)
// 		if err != nil {
// 			return err
// 		}
// 		defer file.Close()

// 		_, err = io.Copy(writer, file)
// 		return err
// 	})

// 	return err
// }

// IsVideoFile checks if a file is a video type based on MIME type.
func IsVideoFile(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read the first 512 bytes to detect MIME type
	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		return false, err
	}

	mimeType := http.DetectContentType(buf)

	return mimeType[:6] == "video/", nil
}

var Config *common.Config

func PrepareAppConfig(path string) (*common.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&Config)
	if err != nil {
		return Config, err
	}

	return Config, nil
}

func GetAppConfig() *common.Config {
	return Config
}

func GenerateConfigJsonTemplate() error {
	template := `{
  "log": {
    "enabled": true,
    "path": "./discordBot.log"
  },
  "torrent": {
    "downloadDir": "./downloads",
    "zipDir": "./zips"
  },
  "http": {
	"domain": "http://localhost:3000",
    "host": "",
    "port": 3000,
    "routes": {
      "video": "/stream/",
      "zip": "/zip/"
    }
  }
}
`

	if FileExists(".config.json") {
		return fmt.Errorf("file .config.json already exists")
	}

	file, err := os.Create(".config.json")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(template)
	if err != nil {
		return err
	}

	return nil
}

func GenerateEnvFileTemplate() error {
	template := `TOKEN=YOUR_BOT_TOKEN
APP_ID=YOUR_APP_ID
SERVICE_ACCOUNT_KEY=FIREBASE_SERVICE_ACCOUNT_KEY_JSON
`

	if FileExists(".env") {
		return fmt.Errorf("file .env already exists")
	}

	file, err := os.Create(".env")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(template)
	if err != nil {
		return err
	}

	return nil
}
