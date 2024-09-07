package torrentCommand

import (
	"discord-bot/discord/components"
	"discord-bot/discord/interaction"
	"discord-bot/torrentClient"
	"discord-bot/utils"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func ytsListOnSelect(data *discordgo.MessageComponentInteractionData, s *discordgo.Session, i *discordgo.InteractionCreate) {
	selectedValue := data.Values[0]
	addTorrent(s, i, &selectedValue, nil)
}

func torrentStopAndRemoveButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	torrentClient.Stop(true)

	sendErr := interaction.RespondWithNothing(s, i)
	if sendErr != nil {
		Log.Error("\ntorrent:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
	}
}

func torrentStopButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	torrentClient.Stop(false)

	sendErr := interaction.RespondWithNothing(s, i)
	if sendErr != nil {
		Log.Error("\ntorrent:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
	}
}

func onTorrentListSelect(s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.MessageComponentInteractionData) {
	sendErr := interaction.RespondWithNothing(s, i)
	if sendErr != nil {
		Log.Error("\ntorrentList:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		return
	}

	torrentID := data.Values[0]
	torrentName := fmt.Sprintf("**%s**\n\u200b", torrentClient.GetTorrentName(torrentID))
	exists := torrentClient.Exists(torrentID)

	if !exists {
		errMsg := "Could not find the torrent with ID: " + torrentID
		Log.Debug(Log.Level.Error, errMsg)
		sendErr = interaction.RespondEdit(s, i, errMsg)
		if sendErr != nil {
			Log.Error("\ntorrentList:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	sendErr = interaction.RespondEditWithComponents(s, i, &torrentName,
		components.AddMessageComponents(
			components.NewRow(
				components.NewButton().SetLabel("Generate Video Links").SetCustomID("torrent_links:"+torrentID).SetStyleSecondary(),
			),
			components.NewRow(
				components.NewButton().SetLabel("Resume").SetCustomID("torrent_resume:"+torrentID).SetStylePrimary(),
				components.NewButton().SetLabel("Remove").SetCustomID("torrent_remove:"+torrentID).SetStyleDanger(),
			),
			components.NewRow(
				components.NewButton().SetLabel("Show Torrents List").SetCustomID("show_torrents_list").SetStyleSecondary(),
			),
		),
	)

	if sendErr != nil {
		Log.Error("\ntorrentList:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		return
	}
}

func generateLinksButton(data *discordgo.MessageComponentInteractionData, s *discordgo.Session, i *discordgo.InteractionCreate) {
	torrentID := strings.Split(data.CustomID, ":")[1]

	sendErr := interaction.RespondWithNothing(s, i)
	if sendErr != nil {
		Log.Error("\ntorrentList:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		return
	}

	tor, err := torrentClient.GetTorrentByID(torrentID)
	if err != nil {
		Log.Debug(Log.Level.Error, "getting a torrent by ID:", err.Error())
		sendErr = interaction.RespondEdit(s, i, fmt.Sprintf("**Error:**: while getting a torrent by ID:\n`%s`", err.Error()))
		if sendErr != nil {
			Log.Error("\ntorrentList:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	urls, err := getVideoUrls(tor)
	if err != nil {
		Log.Debug(Log.Level.Error, "getting a torrent video urls:", err.Error())
		sendErr = interaction.RespondEdit(s, i, fmt.Sprintf("**Error:**: while getting a torrent video urls:\n`%s`", err.Error()))
		if sendErr != nil {
			Log.Error("\ntorrentList:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	content := "No videos found for **" + tor.Name() + "**."

	if len(urls) > 0 {
		content = "Video links for **" + tor.Name() + "**:\n"
		config := utils.GetAppConfig()
		for _, url := range urls {
			content += fmt.Sprint(config.Http.Domain, config.Http.Routes.Video, url, "\n")
		}
	}

	sendErr = interaction.RespondEdit(s, i, content)
	if sendErr != nil {
		Log.Error("\ntorrentList:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
	}
}

func torrentResumeButton(data *discordgo.MessageComponentInteractionData, s *discordgo.Session, i *discordgo.InteractionCreate) {
	torrentID := strings.Split(data.CustomID, ":")[1]

	tor, err := torrentClient.GetTorrentByID(torrentID)
	if err != nil {
		sendErr := interaction.RespondWithNothing(s, i)
		if sendErr != nil {
			Log.Error("\ntorrentList:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
			return
		}

		Log.Debug(Log.Level.Error, "getting a torrent by ID;", err.Error())
		sendErr = interaction.RespondEdit(s, i, fmt.Sprintf("**Error:** while getting a torrent by ID:\n`%s`", err.Error()))
		if sendErr != nil {
			Log.Error("\ntorrentList:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	deleteMsgErr := s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
	if deleteMsgErr != nil {
		Log.Error("\ntorrentList:", deleteMsgErr.Error())
		Log.Debug(Log.Level.Error, `deleting a message for "torrent" command:`, deleteMsgErr.Error())
		return
	}

	addTorrent(s, i, nil, tor)
}

func TorrentDeleteButton(data *discordgo.MessageComponentInteractionData, s *discordgo.Session, i *discordgo.InteractionCreate) {
	torrentID := strings.Split(data.CustomID, ":")[1]

	sendErr := interaction.RespondWithNothing(s, i)
	if sendErr != nil {
		Log.Error("\ntorrentList:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		return
	}

	err := torrentClient.Remove(torrentID)
	if err != nil {
		Log.Debug(Log.Level.Error, "removing a torrent", err.Error())
		sendErr := interaction.RespondEdit(s, i, fmt.Sprintf("**Error:** while removing a torrent:\n`%s`", err.Error()))
		if sendErr != nil {
			Log.Error("\ntorrentList:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	sendErr = interaction.RespondEdit(s, i, "**Success:** Torrent has been removed.")
	if sendErr != nil {
		Log.Error("\ntorrentList:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
	}
}

func showTorrentListButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	msgDeleteErr := s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
	if msgDeleteErr != nil {
		Log.Error("\ntorrentList:", msgDeleteErr.Error())
		Log.Debug(Log.Level.Error, `deleting a message for "torrent" command:`, msgDeleteErr.Error())
	}
	listTorrents(s, i)
}

func onSearchListSelect(data *discordgo.MessageComponentInteractionData, s *discordgo.Session, i *discordgo.InteractionCreate) {
	searchIndexStr := data.Values[0]
	searchIndex, err := strconv.Atoi(searchIndexStr)

	if searchTmp == nil || err != nil || searchIndex < 0 || searchIndex >= len(searchTmp.Results) {
		sendErr := interaction.RespondEdit(s, i, "No search results found.")
		if sendErr != nil {
			Log.Error("\nTorrent:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	searchResult := searchTmp.Results[searchIndex]
	magnet := getMagnet(searchResult.Url)

	if magnet == "" {
		sendErr := interaction.RespondEdit(s, i, "No magnet link found.")
		if sendErr != nil {
			Log.Error("\nTorrent:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	addTorrent(s, i, &magnet, nil)
}

func nextPageButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	sendErr := interaction.RespondWithNothing(s, i)
	if sendErr != nil {
		Log.Error("\ntorrentList:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		return
	}

	if searchTmp == nil {
		sendErr = interaction.RespondEdit(s, i, "No previous page")
		if sendErr != nil {
			Log.Error("\nTorrent:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	searchTmp.Options.page += 1

	results := search_1337x(searchTmp.Options.query, searchTmp.Options.category, searchTmp.Options.sort, searchTmp.Options.page)

	if len(results) == 0 {
		sendErr = interaction.RespondEdit(s, i, "No results found.")
		if sendErr != nil {
			Log.Error("\nTorrent:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	searchTmp.Results = results

	menuOptions := []*components.SelectMenuOption{}
	for i, result := range results {
		if i >= 24 {
			break
		}

		name := result.Name
		if name == "" {
			name = "Error"
		}
		if len(name) > 100 {
			name = name[:97] + "..."
		}

		menuOptions = append(menuOptions, components.NewMenuOption().SetLabel(name).SetValue(strconv.Itoa(i)))
	}

	content := fmt.Sprintf("Found %d results\n **Page:** `%d`", len(results), searchTmp.Options.page)
	if len(results) > 25 {
		content += "\n (Can only show 25 torrents at once)"
	}

	sendErr = interaction.RespondEditWithComponents(s, i, &content,
		components.AddMessageComponents(
			components.NewRow(
				components.NewSelectMenu().SetStringType().SetPlaceholder("Select a result to download").
					SetCustomID("search_list").
					SetOptions(menuOptions...),
			),
			components.NewRow(
				components.NewButton().SetLabel("Next Page").SetCustomID("next_page"),
			),
		),
	)

	if sendErr != nil {
		Log.Error("\nTorrent:", `sending a respond for "torrent" command:`, sendErr.Error())
	}
}
