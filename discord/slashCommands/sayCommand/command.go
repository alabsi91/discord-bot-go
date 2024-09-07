package sayCommand

import (
	tts "discord-bot/TTS"
	"discord-bot/common"
	"discord-bot/discord/events"
	"discord-bot/discord/interaction"
	"discord-bot/utils"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

var Log = &utils.Log

var command = common.SlashCommand{
	Command: discordgo.ApplicationCommand{
		Name:        "say",
		Description: "Play a TTS message in the current voice channel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "message",
				Description: "Enter the message you want the bot to say",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "language",
				Description: "Enter the language of the text",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    false,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Afrikaans", Value: "af"},
					{Name: "Armenian", Value: "hy"},
					{Name: "Indonesian", Value: "id"},
					{Name: "German", Value: "de"},
					{Name: "English", Value: "en"},
					{Name: "Spanish", Value: "es"},
					{Name: "French", Value: "fr"},
					{Name: "Italian", Value: "it"},
					{Name: "Dutch", Value: "nl"},
					{Name: "Norwegian", Value: "nb"},
					{Name: "Polish", Value: "pl"},
					{Name: "Portuguese", Value: "pt"},
					{Name: "Romanian", Value: "ro"},
					{Name: "Finnish", Value: "fi"},
					{Name: "Swedish", Value: "sv"},
					{Name: "Turkish", Value: "tr"},
					{Name: "Greek", Value: "el"},
					{Name: "Russian", Value: "ru"},
					{Name: "Ukrainian", Value: "uk"},
					{Name: "Arabic", Value: "ar"},
					{Name: "Persian", Value: "fa"},
					{Name: "Hindi", Value: "hi"},
					{Name: "Korean", Value: "ko"},
					{Name: "Japanese", Value: "ja"},
					{Name: "Chinese", Value: "zh"},
				},
			},
			{
				Name:        "slow",
				Description: "Whether or not to play the message slowly",
				Type:        discordgo.ApplicationCommandOptionBoolean,
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
	message  string // required
	language string // optional
	slow     bool   // optional
}

func parseCmdOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (cmdOptions, error) {
	results := cmdOptions{}
	for _, opt := range options {

		switch opt.Name {
		case "message":
			val, err := utils.CheckOptionStringValue(opt)
			if err != nil {
				return results, fmt.Errorf("please enter a message")
			}
			results.message = val

		case "language":
			val, err := utils.CheckOptionStringValue(opt)
			if err != nil {
				results.language = "en"
			} else {
				results.language = val
			}

		case "slow":
			if opt.Value == nil {
				results.slow = false
			} else {
				results.slow = opt.BoolValue()
			}
		}
	}
	return results, nil
}

func cmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate, appData *discordgo.ApplicationCommandInteractionData) {
	user := utils.GetInteractionAuthor(i.Interaction)

	Log.Debug(Log.Level.Info, `SlashCommand: "say", GuildID:`, i.GuildID, "ChannelID:", i.ChannelID, "UserID:", user.ID, "UserName:", user.Username)

	options, err := parseCmdOptions(appData.Options)
	if err != nil {
		Log.Debug(Log.Level.Error, `parsing "say" command options:`, err.Error())
		sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while parsing **say** command options:\n`%s`", err.Error()), true)
		if sendError != nil {
			Log.Error("\nSay:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "say" command:`, sendError.Error())
		}
		return
	}

	var (
		message = options.message
		lang    = options.language
		slow    = options.slow
	)

	vc, err := interaction.JoinUserVoiceChannel(s, user.ID)
	if err != nil {
		Log.Debug(Log.Level.Error, `joining the voice channel for the command "say":`, err.Error())
		sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while joining the user's voice channel:\n`%s`", err.Error()), true)
		if sendError != nil {
			Log.Error("\nSay:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "say" command:`, sendError.Error())
		}
		return
	}

	time.Sleep(250 * time.Millisecond)
	vc.Speaking(true)

	sendError := interaction.RespondWithText(s, i, "TTS Message will be played in <#"+vc.ChannelID+">", true)
	if sendError != nil {
		Log.Error("\nSay:", sendError.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "say" command:`, sendError.Error())
		return
	}

	err = tts.GenerateAndSendToVoiceChannel(message, vc, tts.TTSOptions{Lang: lang, Slow: slow})
	if err != nil {
		sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while sending the TTS voice message to the voice channel:\n`%s`", err.Error()), true)
		Log.Debug(Log.Level.Error, err.Error())
		if sendError != nil {
			Log.Error("\nSay:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "say" command:`, sendError.Error())
		}
	}

	vc.Speaking(false)
	time.Sleep(250 * time.Millisecond)
	vc.Disconnect()
}
