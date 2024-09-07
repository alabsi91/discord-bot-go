package events

import (
	"discord-bot/common"

	"github.com/bwmarrin/discordgo"
)

var (
	Interactions  = []*common.SlashCommand{}
	SlashCommands = []*discordgo.ApplicationCommand{}
)

func init() {
	RegisterOnSlashCommandEvent(ExecuteSlashCommands)
}

func RegisterSlashCommand(c *common.SlashCommand) {
	Interactions = append(Interactions, c)
	SlashCommands = append(SlashCommands, &c.Command)
}

func ExecuteSlashCommands(s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionData) {
	for _, interaction := range Interactions {
		if data.Name == interaction.Command.Name {
			interaction.Handler(s, i, data)
			break
		}
	}
}
