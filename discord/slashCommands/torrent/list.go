package torrentCommand

import (
	"discord-bot/discord/components"
	"discord-bot/discord/interaction"
	"discord-bot/torrentClient"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// send a list of all torrents to the user in shape of a select menu
func listTorrents(s *discordgo.Session, i *discordgo.InteractionCreate) {
	allTorrents := torrentClient.GetAllTorrents()

	// no torrents found
	if len(allTorrents) == 0 {
		sendErr := interaction.RespondWithText(s, i, "No torrents found.", true)
		if sendErr != nil {
			Log.Error("\nTorrent:", `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	// create select menu options for each torrent
	var menuOptions []*components.SelectMenuOption
	for i, torrent := range allTorrents {
		if i >= 24 {
			break
		}
		menuOptions = append(menuOptions, components.NewMenuOption().SetLabel(torrent.Name()).SetValue(torrent.ID()))
	}

	content := fmt.Sprintf("Found %d torrents", len(allTorrents))
	if len(allTorrents) > 25 {
		content += ", (Can only show 25 torrents at once)"
	}

	sendErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "\u200b\n" + content,
			Components: *components.AddMessageComponents(
				components.NewRow(
					components.NewSelectMenu().SetStringType().SetPlaceholder("Select a torrent to show actions").
						SetCustomID("torrent_list").
						SetOptions(menuOptions...),
				),
			),
		},
	})

	if sendErr != nil {
		Log.Error("\nTorrent:", `sending a respond for "torrent" command:`, sendErr.Error())
	}
}
