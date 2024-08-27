package slashCommands

import (
	"discord-bot/common"
	"discord-bot/discord/interaction"
	"discord-bot/firebase"
	"discord-bot/utils"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var customCommands = common.SlashCommand{
	Command: discordgo.ApplicationCommand{
		Name:        "custom-command",
		Description: "Manage custom commands",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "add",
				Description: "Add a custom command",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "trigger",
						Description: "The trigger word for the command",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
					{
						Name:        "response",
						Description: "The response for the command",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
			{
				Name:        "remove",
				Description: "Remove a custom command",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "trigger",
						Description: "The trigger word for the command to remove",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
			{
				Name:        "list",
				Description: "List all custom commands",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	},

	Handler: customCommandsHandler,
}

func init() {
	registerCommands(&customCommands)
}

type customCommandsOptions struct {
	subcommand string // "set" or "remove"
	trigger    string // required for "add" and "remove"
	response   string // required for "add"
}

func parseCustomCommandsOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (customCommandsOptions, error) {
	results := customCommandsOptions{}

	subcommand := options[0].Name
	subcommandOptions := options[0].Options

	if subcommand == "add" {
		results.subcommand = "add"

		for _, opt := range subcommandOptions {
			switch opt.Name {
			case "trigger":
				val, err := utils.CheckOptionStringValue(opt)
				if err != nil {
					return results, fmt.Errorf("please enter a trigger word")
				}
				results.trigger = val

			case "response":
				val, err := utils.CheckOptionStringValue(opt)
				if err != nil {
					return results, fmt.Errorf("please enter a response")
				}
				results.response = val
			}
		}
	}

	if subcommand == "remove" {
		results.subcommand = "remove"

		for _, opt := range subcommandOptions {
			switch opt.Name {
			case "trigger":
				val, err := utils.CheckOptionStringValue(opt)
				if err != nil {
					return results, fmt.Errorf("please enter a trigger word")
				}
				results.trigger = val
			}
		}
	}

	if subcommand == "list" {
		results.subcommand = "list"
	}

	return results, nil
}

func customCommandsHandler(s *discordgo.Session, i *discordgo.InteractionCreate, appData *discordgo.ApplicationCommandInteractionData) {
	user := utils.GetInteractionAuthor(i.Interaction)

	Log.Debug(Log.Level.Info, `SlashCommand: "custom-command", GuildID:`, i.GuildID, "ChannelID:", i.ChannelID, "UserID:", user.ID, "UserName:", user.Username)

	options, err := parseCustomCommandsOptions(appData.Options)
	if err != nil {
		Log.Debug(Log.Level.Error, `parsing "custom-command" command options:`, err.Error())
		sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while parsing **custom-command** command options:\n`%s`", err.Error()), true)
		if sendError != nil {
			Log.Error("\nCustomCommands:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "custom-command" command:`, sendError.Error())
		}
		return
	}

	guildData, err := firebase.GetGuildData(i.GuildID)
	if err != nil {
		Log.Debug(Log.Level.Error, `getting guild firebase data for "custom-command" command:`, err.Error())
		sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while getting this guild data from firebase:\n`%s`", err.Error()), true)
		if sendError != nil {
			Log.Error("\nCustomCommands:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "custom-command" command:`, sendError.Error())
		}
		return
	}

	// * ADD
	if options.subcommand == "add" {
		newItem := firebase.CustomCommand{When: options.trigger, Say: options.response}
		currentItem, notFoundErr := guildData.CustomCommandsGetItem(options.trigger)
		exists := notFoundErr == nil

		if exists {
			// update
			guildData.CustomCommandsUpdateItem(newItem)
		} else {
			// add
			guildData.CustomCommandsAddItem(newItem)
		}

		customCommandsMap := guildData.CustomCommandsToMap()
		err = firebase.SetCustomCommand(i.GuildID, &customCommandsMap)
		if err != nil {
			// revoke changes on error
			if exists {
				guildData.CustomCommandsUpdateItem(*currentItem)
			} else {
				guildData.CustomCommandsRemoveItem(options.trigger)
			}

			Log.Debug(Log.Level.Error, `uploading "custom-command (add)" data to firebase:`, err.Error())
			sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while uploading **custom-command (add)** data to firebase:\n`%s`", err.Error()), true)
			if sendError != nil {
				Log.Error("\nCustomCommands:", sendError.Error())
				Log.Debug(Log.Level.Error, `sending a respond for "custom-command" command:`, sendError.Error())
			}
			return
		}

		sendError := interaction.RespondWithText(s, i, "**Success:** Custom command added", true)
		if sendError != nil {
			Log.Error("\nCustomCommands:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "custom-command" command:`, sendError.Error())
		}
		return
	}

	// * REMOVE
	if options.subcommand == "remove" {
		currentItem, notFoundError := guildData.CustomCommandsGetItem(options.trigger)
		if notFoundError != nil {
			Log.Debug(Log.Level.Error, `getting a "custom-command" item:`, notFoundError.Error())
			sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while getting a **custom-command** item:\n`%s`", notFoundError.Error()), true)
			if sendError != nil {
				Log.Error("\nCustomCommands:", sendError.Error())
				Log.Debug(Log.Level.Error, `sending a respond for "custom-command" command:`, sendError.Error())
			}
			return
		}

		guildData.CustomCommandsRemoveItem(options.trigger)

		customCommandsMap := guildData.CustomCommandsToMap()
		err = firebase.SetCustomCommand(i.GuildID, &customCommandsMap)
		if err != nil {
			// revoke changes on error
			guildData.CustomCommandsAddItem(*currentItem)

			Log.Debug(Log.Level.Error, `uploading "custom-command (remove)" data to firebase:`, err.Error())
			sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while uploading **custom-command (remove)** data to firebase:\n`%s`", err.Error()), true)
			if sendError != nil {
				Log.Error("\nCustomCommands:", sendError.Error())
				Log.Debug(Log.Level.Error, `sending a respond for "custom-command" command:`, sendError.Error())
			}
			return
		}

		// send confirmation
		sendError := interaction.RespondWithText(s, i, "**Success:** Welcome message removed", true)
		if sendError != nil {
			Log.Error("\nCustomCommands:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "custom-command" command:`, sendError.Error())
		}
		return
	}

	// * LIST
	if options.subcommand == "list" {
		if len(guildData.CustomCommands) == 0 {
			sendError := interaction.RespondWithText(s, i, "⚠️  No custom commands found", true)
			if sendError != nil {
				Log.Error("\nCustomCommands:", sendError.Error())
				Log.Debug(Log.Level.Error, `sending a respond for "custom-command" command:`, sendError.Error())
			}
			return
		}

		formattedResponse := "A list of custom commands:\n"
		for _, item := range guildData.CustomCommands {
			formattedResponse += fmt.Sprintf("🔹 **%s**: %s\n", item.When, item.Say)
		}

		sendError := interaction.RespondWithText(s, i, formattedResponse, true)
		if sendError != nil {
			Log.Error("\nCustomCommands:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "custom-command" command:`, sendError.Error())
		}
		return
	}
}
