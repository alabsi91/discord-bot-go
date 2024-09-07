package prefixCommand

import (
	"discord-bot/common"
	"discord-bot/discord/events"
	"discord-bot/discord/interaction"
	"discord-bot/firebase"
	"discord-bot/utils"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var Log = &utils.Log

var command = common.SlashCommand{
	Command: discordgo.ApplicationCommand{
		Name:        "prefix",
		Description: "Change/get the prefix for message commands",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "set",
				Description: "Set/update the prefix for message commands",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "prefix",
						Description: "The prefix to set",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
			{
				Name:        "get",
				Description: "Show the current prefix",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	},

	Handler: cmdHandler,
}

func init() {
	events.RegisterSlashCommand(&command)
}

type cmdOptions struct {
	subcommand string // "set" or "get"
	prefix     string // required for "set"
}

func parseCmdOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (cmdOptions, error) {
	results := cmdOptions{}

	subcommand := options[0].Name
	subcommandOptions := options[0].Options

	if subcommand == "set" {
		results.subcommand = "set"

		for _, opt := range subcommandOptions {
			switch opt.Name {
			case "prefix":
				val, err := utils.CheckOptionStringValue(opt)
				if err != nil {
					return results, fmt.Errorf("please enter a prefix")
				}
				results.prefix = val
			}
		}
	}

	if subcommand == "get" {
		results.subcommand = "get"
	}

	return results, nil
}

func cmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate, appData *discordgo.ApplicationCommandInteractionData) {
	user := utils.GetInteractionAuthor(i.Interaction)

	Log.Debug(Log.Level.Info, `SlashCommand: "prefix", GuildID:`, i.GuildID, "ChannelID:", i.ChannelID, "UserID:", user.ID, "UserName:", user.Username)

	options, err := parseCmdOptions(appData.Options)
	if err != nil {
		Log.Debug(Log.Level.Error, `parsing "prefix" command options:`, err.Error())
		sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while parsing **prefix** command options:\n`%s`", err.Error()), true)
		if sendError != nil {
			Log.Error("\nprefix:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "prefix" command:`, sendError.Error())
		}
		return
	}

	guildData, err := firebase.GetGuildData(i.GuildID)
	if err != nil {
		Log.Debug(Log.Level.Error, `getting guild firebase data for "prefix" command:`, err.Error())
		sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while getting this guild data from firebase:\n`%s`", err.Error()), true)
		if sendError != nil {
			Log.Error("\nprefixCommand:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "prefix" command:`, sendError.Error())
		}
		return
	}

	// * SET
	if options.subcommand == "set" {
		err := firebase.SetCommandPrefix(i.GuildID, options.prefix)
		if err != nil {
			Log.Debug(Log.Level.Error, `uploading "prefix (set)" data to firebase:`, err.Error())
			sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while uploading **prefix (set)** data to firebase:\n`%s`", err.Error()), true)
			if sendError != nil {
				Log.Error("\nprefixCommand:", sendError.Error())
				Log.Debug(Log.Level.Error, `sending a respond for "prefix" command:`, sendError.Error())
				return
			}
		}

		sendError := interaction.RespondWithText(s, i, "**Success:** Prefix set to: "+options.prefix, true)
		if sendError != nil {
			Log.Error("\nprefixCommand:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "prefix" command:`, sendError.Error())
		}

		return
	}

	// * GET
	if options.subcommand == "get" {
		sendError := interaction.RespondWithText(s, i, fmt.Sprintf("The current prefix is: `%s`", guildData.Prefix), true)
		if sendError != nil {
			Log.Error("\nprefixCommand:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "prefix" command:`, sendError.Error())
		}
	}
}
