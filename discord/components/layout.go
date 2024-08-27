package components

import "github.com/bwmarrin/discordgo"

func NewRow(components ...discordgo.MessageComponent) *discordgo.ActionsRow {
	return &discordgo.ActionsRow{
		Components: components,
	}
}

func AddMessageComponents(components ...discordgo.MessageComponent) *[]discordgo.MessageComponent {
	return &components
}
