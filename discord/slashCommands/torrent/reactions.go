package torrentCommand

import (
	"discord-bot/common"
	"discord-bot/discord/components"
	"discord-bot/discord/interaction"
	"discord-bot/torrentClient"
	"discord-bot/utils"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

	search, ok := searchTmp[i.GuildID+i.ChannelID]

	if !ok || err != nil || searchIndex < 0 || searchIndex >= len(search.Results) {
		sendErr := interaction.RespondWithText(s, i, "No search results found.", false)
		if sendErr != nil {
			Log.Error("\nTorrent:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	// should edit or send new embed
	shouldEdit := false
	messageId := ""
	messages, err := s.ChannelMessages(i.ChannelID, 1, "", "", "")
	if err == nil && len(messages) > 0 {
		embeds := messages[0].Embeds
		for _, embeds := range embeds {
			if strings.Contains(embeds.Footer.Text, "1337x.to") {
				shouldEdit = true
				messageId = messages[0].ID
				break
			}
		}
	}

	searchResult := search.Results[searchIndex]

	if shouldEdit {
		sendErr := interaction.RespondWithNothing(s, i)
		if sendErr != nil {
			Log.Error("\nTorrent:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}

		content := "_This message will be deleted also._\n"
		_, sendErr = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel: i.ChannelID,
			ID:      messageId,
			Content: &content,
			Embeds:  &[]*discordgo.MessageEmbed{createSearchDetailEmbed(searchResult)},
			Components: components.AddMessageComponents(
				components.NewRow(
					components.NewButton().SetLabel("Download").SetCustomID("search_download:" + searchIndexStr).SetStyleSecondary(),
				),
			),
		})
		if sendErr != nil {
			Log.Error("\nTorrent:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}

		return
	}

	content := "_This message will be deleted also_\n"
	sendErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Embeds:  []*discordgo.MessageEmbed{createSearchDetailEmbed(searchResult)},
			Components: *components.AddMessageComponents(
				components.NewRow(
					components.NewButton().SetLabel("Download").SetCustomID("search_download:" + searchIndexStr).SetStyleSecondary(),
				),
			),
		},
	})
	if sendErr != nil {
		Log.Error("\nTorrent:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
	}

	// get message id to remove later
	msg, followupErr := s.InteractionResponse(i.Interaction)
	if followupErr != nil {
		Log.Error("Failed to retrieve response message: ", followupErr.Error())
		return
	}
	search.EmbedMsgID = msg.ID
}

func nextPageButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	sendErr := interaction.RespondWithNothing(s, i)
	if sendErr != nil {
		Log.Error("\ntorrentList:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		return
	}

	search, ok := searchTmp[i.GuildID+i.ChannelID]
	if !ok {
		sendErr = interaction.RespondEdit(s, i, "No previous page")
		if sendErr != nil {
			Log.Error("\nTorrent:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	search.Options.page += 1

	results := search_1337x(search.Options.query, search.Options.category, search.Options.sort, search.Options.page)

	if len(results) == 0 {
		sendErr = interaction.RespondEdit(s, i, "No results found.")
		if sendErr != nil {
			Log.Error("\nTorrent:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	search.Results = results

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

	content := "_This message Will be deleted in `1` minute_\n\u200b\n"
	content += fmt.Sprintf("**Page:** `%d` has `%d` results:\n ", search.Options.page, len(results))
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

func searchDownloadButton(data *discordgo.MessageComponentInteractionData, s *discordgo.Session, i *discordgo.InteractionCreate) {
	indexStr := strings.Split(data.CustomID, ":")[1]
	searchIndex, err := strconv.Atoi(indexStr)

	search, ok := searchTmp[i.GuildID+i.ChannelID]
	if !ok || err != nil || searchIndex < 0 || searchIndex >= len(search.Results) {
		sendErr := interaction.RespondWithText(s, i, "No search results found.", false)
		if sendErr != nil {
			Log.Error("\nTorrent:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	searchResult := search.Results[searchIndex]
	magnet := getMagnet(searchResult.Url)

	if magnet == "" {
		sendErr := interaction.RespondEdit(s, i, "No magnet link found.")
		if sendErr != nil {
			Log.Error("\nTorrent:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	err = s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
	if err != nil {
		Log.Error("\nTorrent:", err.Error())
		Log.Debug(Log.Level.Error, "deleting a message:", err.Error())
	}

	addTorrent(s, i, &magnet, nil)
}

// utils

func createSearchDetailEmbed(result common.SearchResult) *discordgo.MessageEmbed {
	embed := components.NewEmbed().
		SetColor(0xf14e13).
		SetTitle(result.Name).
		SetURL(result.Url).
		AddField("\u200b\n"+cleanTitle(result.Name), "\u200b", false).
		AddField("Resolution", extractResolution(result.Name), true).
		AddField("Quality", extractQuality(result.Name), true).
		AddField("Date", result.Date, true).
		AddField("\u200b", "", false).
		AddField("Size", result.Size, true).
		AddField("Seeds", result.Seeds, true).
		AddField("Leeches", result.Leeches, true).
		AddField("\u200b", "", false).
		SetFooter("Powered by 1337x.to")

	return embed.Into()
}

func cleanTitle(name string) string {
	re := regexp.MustCompile(`(?i)(?:(?:\s|\.)?.+?(?:\s|\.)?)+(?:(?:\bS\d{2}E\d{2}\b)|(?:\bS\d{2}\b)|(?:\b(?:19|20)\d{2}\b))`)
	match := re.FindString(name)

	// Remove special characters
	spacialChars := regexp.MustCompile(`[^\w\s]`)
	match = spacialChars.ReplaceAllString(match, " ")
	match = strings.ReplaceAll(match, "  ", " ")

	// Capitalize first letter
	match = cases.Title(language.English).String(match)

	// Capitalize season and episode
	re = regexp.MustCompile(`(?i)\bS\d{2}E\d{2}\b`)
	SE := re.FindString(match)
	if SE != "" {
		match = strings.ReplaceAll(match, SE, strings.ToUpper(SE))
	}

	// trim
	match = strings.TrimSpace(match)

	return match
}

func extractResolution(name string) string {
	resolutions := []string{"2160p", "1080p", "720p", "480p", "360p", "240p", "144p"}
	for _, resolution := range resolutions {
		if strings.Contains(name, resolution) {
			return resolution
		}
	}
	return "unknown"
}

func extractQuality(name string) string {
	qualities := []string{"WEBRip", "WEB-DL", "WEB", "BluRay", "DVDRip", "HDTV", "HDTS", "HDRip", "CAM", "TVRip"}
	for _, quality := range qualities {
		if strings.Contains(strings.ToLower(name), strings.ToLower(quality)) {
			return quality
		}
	}
	return "unknown"
}
