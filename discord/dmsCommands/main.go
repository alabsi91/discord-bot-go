package dmsCommands

import (
	"discord-bot/common"
	"discord-bot/firebase"
	"discord-bot/utils"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var Log = &utils.Log
var dms = []*common.DmsCommand{}

func registerDmsCommands(c *common.DmsCommand) {
	dms = append(dms, c)
}

func ExecuteDmsCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	// get guild data for command prefix and custom commands
	guildData, err := firebase.GetGuildData(m.GuildID)
	if err != nil {
		_, sendError := s.ChannelMessageSendReply(m.ChannelID, "❗ Something went wrong while getting the guild data.", m.Reference())
		if sendError != nil {
			Log.Error("\nOnDM:", sendError.Error())
		}

		Log.Error("\nOnDM:", err.Error())
		return
	}

	content := strings.Trim(m.Content, " ")

	for _, command := range dms {
		withoutCommand, isCommand := strings.CutPrefix(content, guildData.Prefix+command.Name)
		if isCommand {
			args := strings.Split(withoutCommand, " ")
			args = args[1:] // remove empty space
			subcommand := ""

			if len(args) != 0 && utils.Contains(&command.Subcommands, args[0]) {
				subcommand = args[0]
				args = args[1:] // remove subcommand
			}

			command.Handler(s, m, subcommand, &args)
			return
		}
	}

	// try custom commands
	for _, command := range guildData.CustomCommands {
		if content == command.When {
			_, err := s.ChannelMessageSend(m.ChannelID, command.Say)
			if err != nil {
				Log.Error("\nCustomCommands:", err.Error())
			}
			return
		}
	}
}
