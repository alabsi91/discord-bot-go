package slashCommands

import (
	"discord-bot/common"
	"discord-bot/utils"

	"github.com/bwmarrin/discordgo"
)

var Log = &utils.Log

var (
	Interactions = []*common.SlashCommand{}
	AppCommands  = []*discordgo.ApplicationCommand{}
)

func registerCommands(c *common.SlashCommand) {
	Interactions = append(Interactions, c)
	AppCommands = append(AppCommands, &c.Command)
}

func ExecuteSlashCommands(s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionData) {
	for _, interaction := range Interactions {
		if data.Name == interaction.Command.Name {
			interaction.Handler(s, i, data)
			break
		}
	}
}
