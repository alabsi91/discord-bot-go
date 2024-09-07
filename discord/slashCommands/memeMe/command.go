package memeMe

import (
	"bytes"
	"discord-bot/common"
	"discord-bot/discord/events"
	"discord-bot/discord/interaction"
	"discord-bot/utils"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

var Log = &utils.Log

var command = common.SlashCommand{
	Command: discordgo.ApplicationCommand{
		Name:        "mememe",
		Description: "Get random meme image with a text on it",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "first_line",
				Description: "Write a text on top of the image",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    false,
			},
			{
				Name:        "second_line",
				Description: "Write a text on the bottom of the image",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    false,
			},
		},
	},

	Handler: cmdHandler,
}

func init() {
	events.RegisterSlashCommand(&command)
}

type cmdOptions struct {
	first_line  string // optional
	second_line string // optional
}

func memeCmdOptions(options []*discordgo.ApplicationCommandInteractionDataOption) cmdOptions {
	results := cmdOptions{}

	for _, opt := range options {
		switch opt.Name {
		case "first_line":
			val, _ := utils.CheckOptionStringValue(opt)
			results.first_line = val

		case "second_line":
			val, _ := utils.CheckOptionStringValue(opt)
			results.second_line = val
		}
	}

	return results
}

func cmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate, appData *discordgo.ApplicationCommandInteractionData) {
	user := utils.GetInteractionAuthor(i.Interaction)

	Log.Debug(Log.Level.Info, `SlashCommand: "mememe", GuildID:`, i.GuildID, "ChannelID:", i.ChannelID, "UserID:", user.ID, "UserName:", user.Username)

	options := memeCmdOptions(appData.Options)

	firstLine := options.first_line
	secondLine := options.second_line

	sendingErr := interaction.RespondWithThinking(s, i, false)
	if sendingErr != nil {
		Log.Error("\nMemeMe:", sendingErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "mememe" command:`, sendingErr.Error())
		return
	}

	memeMeData, err := fetchMemeData()
	if err != nil {
		Log.Debug(Log.Level.Error, "fetching a meme image:", err.Error())
		interactionErr := interaction.RespondEdit(s, i, fmt.Sprintf("**Error:**: while fetching a meme image:\n`%s`", err.Error()))
		if interactionErr != nil {
			Log.Error("\nMemeMe:", interactionErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "mememe" command:`, interactionErr.Error())
		}
		return
	}

	// random meme
	meme := memeMeData.Data.Memes[utils.RandomInt(0, len(memeMeData.Data.Memes)-1)]

	//  Create meme image
	buf, err := createMemeImage(&meme, &firstLine, &secondLine)
	if err != nil {
		Log.Debug(Log.Level.Error, "adding text on a meme image:", err.Error())
		interactionErr := interaction.RespondEdit(s, i, fmt.Sprintf("**Error:**: while adding text on a meme image:\n`%s`", err.Error()))
		if interactionErr != nil {
			Log.Error("\nMemeMe:", interactionErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "mememe" command:`, interactionErr.Error())
		}
		return
	}

	_, sendingErr = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Files: []*discordgo.File{{Name: "meme.png", Reader: buf}},
	})
	if sendingErr != nil {
		Log.Error("\nMemeMe:", sendingErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "mememe" command:`, sendingErr.Error())
	}
}

// fetchMemeData fetches a list of meme images data
func fetchMemeData() (*common.MemeResponse, error) {
	// Make the HTTP GET request
	resp, err := http.Get("https://api.imgflip.com/get_memes")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the JSON response
	var response common.MemeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, errors.New("response returned a non-successful status code")
	}

	return &response, nil
}

// loadImgFromUrl loads an image from a URL
func loadImgFromUrl(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// createMemeImage creates a meme image with text on it
func createMemeImage(meme *common.Meme, firstLine *string, secondLine *string) (*bytes.Buffer, error) {
	var imageBuffer bytes.Buffer

	// load the image from url
	img, err := loadImgFromUrl(meme.Url)
	if err != nil {
		return &imageBuffer, err
	}

	// load the font
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return &imageBuffer, err
	}

	// create the canvas
	dc := gg.NewContext(meme.Width, meme.Height)

	// draw the image to the canvas
	dc.DrawImage(img, 0, 0)

	face := truetype.NewFace(font, &truetype.Options{Size: float64(meme.Width) * 0.075})
	dc.SetFontFace(face)

	// draw top text

	// stroke
	dc.SetLineWidth(10) // stroke width
	dc.SetRGB(0, 0, 0)  // stroke color
	dc.DrawStringAnchored(*firstLine, float64(meme.Width/2)-1, dc.FontHeight()-1, 0.5, 0.5)
	dc.Stroke()

	// fill
	dc.SetRGB255(255, 255, 255)
	dc.DrawStringAnchored(*firstLine, float64(meme.Width/2), dc.FontHeight(), 0.5, 0.5)
	dc.Fill()

	// draw the bottom text

	// stroke
	dc.SetLineWidth(10)
	dc.SetRGB(0, 0, 0)
	dc.DrawStringAnchored(*secondLine, float64(meme.Width/2)-1, float64(meme.Height)-dc.FontHeight()-1, 0.5, 0.5)
	dc.Stroke()

	// fill
	dc.SetRGB255(255, 255, 255)
	dc.DrawStringAnchored(*secondLine, float64(meme.Width/2), float64(meme.Height)-dc.FontHeight(), 0.5, 0.5)
	dc.Fill()

	// create file attachment
	err = dc.EncodePNG(&imageBuffer)
	if err != nil {
		return &imageBuffer, err
	}

	return &imageBuffer, nil
}
