package slashCommands

import (
	"discord-bot/common"
	"discord-bot/discord/interaction"
	"discord-bot/firebase"
	"discord-bot/utils"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var botActivityCommand = common.SlashCommand{
	Command: discordgo.ApplicationCommand{
		Name:        "bot-activity",
		Description: "Change bot activity",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "status",
				Description: "Write the status message",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "activity",
				Description: "Current activity",
				Type:        discordgo.ApplicationCommandOptionInteger,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Playing", Value: discordgo.ActivityTypeGame},
					{Name: "Streaming", Value: discordgo.ActivityTypeStreaming},
					{Name: "Listening", Value: discordgo.ActivityTypeListening},
					{Name: "Watching", Value: discordgo.ActivityTypeWatching},
					{Name: "Competing", Value: discordgo.ActivityTypeCompeting},
					{Name: "Custom", Value: discordgo.ActivityTypeCustom},
				},
				Required: true,
			},
		},
	},

	Handler: botActivityHandler,
}

func init() {
	registerCommands(&botActivityCommand)
}

type botActivityOptions struct {
	status   string                 // required
	activity discordgo.ActivityType // required
}

func parseBotActivityOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (botActivityOptions, error) {
	results := botActivityOptions{}

	for _, opt := range options {
		switch opt.Name {

		case "status":
			val, err := utils.CheckOptionStringValue(opt)
			if err != nil {
				return results, fmt.Errorf("please enter a status message")
			}
			results.status = val

		case "activity":
			if opt.Value == nil {
				return results, fmt.Errorf("please choose an activity")
			}
			results.activity = discordgo.ActivityType(opt.IntValue())
		}
	}

	return results, nil
}

func botActivityHandler(s *discordgo.Session, i *discordgo.InteractionCreate, appData *discordgo.ApplicationCommandInteractionData) {
	user := utils.GetInteractionAuthor(i.Interaction)

	Log.Debug(Log.Level.Info, `SlashCommand: "bot-activity", GuildID:`, i.GuildID, "ChannelID:", i.ChannelID, "UserID:", user.ID, "UserName:", user.Username)

	options, err := parseBotActivityOptions(appData.Options)
	if err != nil {
		Log.Debug(Log.Level.Error, `parsing "bot-activity" command options:`, err.Error())
		sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while parsing **bot-activity** command options:\n`%s`", err.Error()), true)
		if sendError != nil {
			Log.Error("\nbotActivity:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "bot-activity" command:`, sendError.Error())
		}
		return
	}

	err = firebase.SetBotActivity(firebase.BotActivity{
		Activity:     options.status,
		ActivityType: options.activity,
	})
	if err != nil {
		sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while uploading bot activity data to firebase::\n`%s`", err.Error()), true)
		Log.Debug(Log.Level.Error, `uploading "bot-activity" data to firebase:`, err.Error())
		if sendError != nil {
			Log.Error("\nbotActivity:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "bot-activity" command:`, sendError.Error())
		}
		return
	}

	err = s.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{{Name: options.status, Type: options.activity}},
	})
	if err != nil {
		sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while updating bot activity status:\n`%s`", err.Error()), true)
		Log.Debug(Log.Level.Error, `updating "bot-activity" in discord:`, err.Error())
		if sendError != nil {
			Log.Error("\nbotActivity:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "bot-activity" command:`, sendError.Error())
		}
		return
	}

	sendError := interaction.RespondWithText(s, i, "**Success:** Bot activity updated", true)
	if sendError != nil {
		Log.Error("\nbotActivity:", sendError.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "bot-activity" command:`, sendError.Error())
	}
}
