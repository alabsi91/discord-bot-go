package discord

import (
	"discord-bot/discord/dmsCommands"
	"discord-bot/discord/events"
	_ "discord-bot/discord/slashCommands" // to initialize slash commands
	"discord-bot/utils"
	"os"

	"github.com/bwmarrin/discordgo"
)

var Log = utils.Log

// StartDiscordBotSession starts the Discord session
func StartDiscordBotSession() *discordgo.Session {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + os.Getenv("TOKEN"))
	if err != nil {
		Log.Fatal("\nerror creating Discord session:", err.Error())
		Log.Debug(Log.Level.Fatal, err.Error())
	}

	// * Register Handlers.

	// On Ready
	dg.AddHandler(events.OnReady)

	// Register Slash Commands
	_, err = dg.ApplicationCommandBulkOverwrite(os.Getenv("APP_ID"), "", events.SlashCommands)
	if err != nil {
		Log.Fatal("\ncould not register commands:", err.Error())
		Log.Debug(Log.Level.Fatal, err.Error())
	}

	// On Interaction
	dg.AddHandler(events.OnInteraction)

	// On DM
	events.RegisterOnDmMessageEvent(dmsCommands.ExecuteDmsCommands)
	dg.AddHandler(events.OnDM)

	// Register the voice state update handler
	dg.AddHandler(events.VoiceStateUpdate)

	// * Begin listening.
	err = dg.Open()
	if err != nil {
		Log.Fatal("\nerror opening Discord session:", err.Error())
		Log.Debug(Log.Level.Fatal, err.Error())
	}

	return dg
}
