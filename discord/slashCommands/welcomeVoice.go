package slashCommands

import (
	"discord-bot/common"
	"discord-bot/discord/interaction"
	"discord-bot/firebase"
	"discord-bot/utils"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var welcomeVoiceMessage = common.SlashCommand{
	Command: discordgo.ApplicationCommand{
		Name:        "welcome-voice-message",
		Description: "Set a welcome TTS message every time someone joins the voice channel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "set",
				Description: "Set/update a welcome TTS message",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "user",
						Description: "The user to set the welcome message for",
						Type:        discordgo.ApplicationCommandOptionUser,
						Required:    true,
					},
					{
						Name:        "message",
						Description: "The welcome message to set",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
					{
						Name:        "language",
						Description: "Enter the language of the text",
						Type:        discordgo.ApplicationCommandOptionString,
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
						Required: false,
					},
				},
			},
			{
				Name:        "remove",
				Description: "Remove a welcome TTS message",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "user",
						Description: "The user to remove the welcome message for",
						Type:        discordgo.ApplicationCommandOptionUser,
						Required:    true,
					},
				},
			},
		},
	},

	Handler: welcomeVoiceMessageHandler,
}

func init() {
	registerCommands(&welcomeVoiceMessage)
}

type welcomeVoiceMessageOptions struct {
	subcommand string                                             // "set" or "remove"
	user       *discordgo.ApplicationCommandInteractionDataOption // required for "set" and "remove"
	message    string                                             // required for "set"
	language   string                                             // optional
}

func parseWelcomeVoiceMessageOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (welcomeVoiceMessageOptions, error) {
	results := welcomeVoiceMessageOptions{}

	subcommand := options[0].Name
	subcommandOptions := options[0].Options

	if subcommand == "remove" {
		results.subcommand = "remove"

		for _, opt := range subcommandOptions {
			switch opt.Name {
			case "user":
				if opt.Value == nil {
					return results, fmt.Errorf("please choose a user")
				}
				results.user = opt
			}
		}
	}

	if subcommand == "set" {
		results.subcommand = "set"

		for _, opt := range subcommandOptions {
			switch opt.Name {
			case "user":
				if opt.Value == nil {
					return results, fmt.Errorf("please choose a user")
				}
				results.user = opt

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
			}
		}
	}

	return results, nil
}

func welcomeVoiceMessageHandler(s *discordgo.Session, i *discordgo.InteractionCreate, appData *discordgo.ApplicationCommandInteractionData) {
	user := utils.GetInteractionAuthor(i.Interaction)

	Log.Debug(Log.Level.Info, `SlashCommand: "welcome-voice-message", GuildID:`, i.GuildID, "ChannelID:", i.ChannelID, "UserID:", user.ID, "UserName:", user.Username)

	options, err := parseWelcomeVoiceMessageOptions(appData.Options)
	if err != nil {
		Log.Debug(Log.Level.Error, `parsing "welcome-voice-message" command options:`, err.Error())
		sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while parsing **welcome-voice-message** command options:\n`%s`", err.Error()), true)
		if sendError != nil {
			Log.Error("\nWelcomeVoiceMessage:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "welcome-voice-message" command:`, sendError.Error())
		}
		return
	}

	setForUser := options.user.UserValue(s)
	if setForUser.Bot {
		sendError := interaction.RespondWithText(s, i, "**Error:** Cannot set welcome message for bots", true)
		if sendError != nil {
			Log.Error("\nWelcomeVoiceMessage:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "welcome-voice-message" command:`, sendError.Error())
		}
		return
	}

	guildData, err := firebase.GetGuildData(i.GuildID)
	if err != nil {
		Log.Debug(Log.Level.Error, `getting guild firebase data for "welcome-voice-message" command:`, err.Error())
		sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while getting this guild data from firebase:\n`%s`", err.Error()), true)
		if sendError != nil {
			Log.Error("\nWelcomeVoiceMessage:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "welcome-voice-message" command:`, sendError.Error())
		}
		return
	}

	// * SET
	if options.subcommand == "set" {
		newItem := firebase.VoiceWelcomeMessage{
			Id:      setForUser.ID,
			Message: options.message,
			Lang:    options.language,
		}

		currentItem, notFoundErr := guildData.VoiceMessagesGetItem(setForUser.ID)
		exists := notFoundErr == nil
		if exists {
			// update item
			guildData.VoiceMessagesUpdateItem(newItem)
		} else {
			// add item
			guildData.VoiceMessagesAddItem(newItem)
		}

		voiceMessagesMap := guildData.VoiceMessagesToMap()
		err = firebase.SetVoiceMessages(i.GuildID, &voiceMessagesMap)
		if err != nil {
			// revoke changes on error
			if exists {
				guildData.VoiceMessagesUpdateItem(*currentItem)
			} else {
				guildData.VoiceMessagesRemoveItem(newItem.Id)
			}

			Log.Debug(Log.Level.Error, `uploading "welcome-voice-message (set)" data to firebase:`, err.Error())
			sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while uploading **welcome-voice-message (set)** data to firebase:\n`%s`", err.Error()), true)
			if sendError != nil {
				Log.Error("\nWelcomeVoiceMessage:", sendError.Error())
				Log.Debug(Log.Level.Error, `sending a respond for "welcome-voice-message" command:`, sendError.Error())
			}
			return
		}

		sendError := interaction.RespondWithText(s, i, "**Success:** Welcome message set", true)
		if sendError != nil {
			Log.Error("\nWelcomeVoiceMessage:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "welcome-voice-message" command:`, sendError.Error())
		}
		return
	}

	// * REMOVE
	if options.subcommand == "remove" {
		currentItem, notFoundError := guildData.VoiceMessagesGetItem(setForUser.ID)
		if notFoundError != nil {
			sendError := interaction.RespondWithText(s, i, "**Error:** No welcome message set for this user", true)
			Log.Debug(Log.Level.Error, notFoundError.Error())
			if sendError != nil {
				Log.Error("\nWelcomeVoiceMessage:", sendError.Error())
				Log.Debug(Log.Level.Error, `sending a respond for "welcome-voice-message" command:`, sendError.Error())
			}
			return
		}

		guildData.VoiceMessagesRemoveItem(setForUser.ID)

		voiceMessagesMap := guildData.VoiceMessagesToMap()
		err = firebase.SetVoiceMessages(i.GuildID, &voiceMessagesMap)
		if err != nil {
			// revoke changes on error
			guildData.VoiceMessagesAddItem(*currentItem)

			Log.Debug(Log.Level.Error, `uploading "welcome-voice-message (remove)" data to firebase:`, err.Error())
			sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while uploading **welcome-voice-message (remove)** data to firebase:\n`%s`", err.Error()), true)
			if sendError != nil {
				Log.Error("\nWelcomeVoiceMessage:", sendError.Error())
				Log.Debug(Log.Level.Error, `sending a respond for "welcome-voice-message" command:`, sendError.Error())
			}
			return
		}

		// send confirmation
		sendError := interaction.RespondWithText(s, i, "**Success:** Welcome message removed", true)
		if sendError != nil {
			Log.Error("\nWelcomeVoiceMessage:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "welcome-voice-message" command:`, sendError.Error())
		}
		return
	}
}
