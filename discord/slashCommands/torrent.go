package slashCommands

import (
	"discord-bot/common"
	"discord-bot/discord/components"
	"discord-bot/discord/events"
	"discord-bot/discord/interaction"
	"discord-bot/torrentClient"
	"discord-bot/utils"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cenkalti/rain/torrent"
	"github.com/gocolly/colly/v2"
)

type SearchTmp struct {
	Options *torrentOptions
	Results []common.SearchResult
}

var searchTmp *SearchTmp

var torrentCommand = common.SlashCommand{
	Command: discordgo.ApplicationCommand{
		Name:        "torrent",
		Description: "Manage and download torrents",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "add",
				Description: "Adds a torrent to be downloaded",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "uri",
						Description: "The URI of the torrent",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
			{
				Name:        "list",
				Description: "List all torrents",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        "search",
				Description: "Search for torrents on `https://1337x.to` and download them",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "query",
						Description: "The query to search for",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
					{
						Name:        "category",
						Description: "The category to search in (optional)",
						Type:        discordgo.ApplicationCommandOptionInteger,
						Required:    false,
						Choices: []*discordgo.ApplicationCommandOptionChoice{
							{Name: common.CategoryAll.String(), Value: common.CategoryAll},
							{Name: common.CategoryMovie.String(), Value: common.CategoryMovie},
							{Name: common.CategoryTV.String(), Value: common.CategoryTV},
							{Name: common.CategoryAnime.String(), Value: common.CategoryAnime},
							{Name: common.CategoryDocumentaries.String(), Value: common.CategoryDocumentaries},
							{Name: common.CategoryGames.String(), Value: common.CategoryGames},
							{Name: common.CategoryMusic.String(), Value: common.CategoryMusic},
							{Name: common.CategoryOther.String(), Value: common.CategoryOther},
							{Name: common.CategoryXXX.String(), Value: common.CategoryXXX},
						},
					},
					{
						Name:        "sort",
						Description: "The sort order (optional)",
						Type:        discordgo.ApplicationCommandOptionInteger,
						Required:    false,
						Choices: []*discordgo.ApplicationCommandOptionChoice{
							{Name: common.SortNone.String(), Value: common.SortNone},
							{Name: common.SortTimeDesc.String(), Value: common.SortTimeDesc},
							{Name: common.SortTimeAsc.String(), Value: common.SortTimeAsc},
							{Name: common.SortSizeDesc.String(), Value: common.SortSizeDesc},
							{Name: common.SortSizeAsc.String(), Value: common.SortSizeAsc},
							{Name: common.SortSeedersDesc.String(), Value: common.SortSeedersDesc},
							{Name: common.SortSeedersAsc.String(), Value: common.SortSeedersAsc},
							{Name: common.SortLeechersDesc.String(), Value: common.SortLeechersDesc},
							{Name: common.SortLeechersAsc.String(), Value: common.SortLeechersAsc},
						},
					},
					{
						Name:        "page",
						Description: "The page number (optional)",
						Type:        discordgo.ApplicationCommandOptionInteger,
						Required:    false,
					},
				},
			},
		},
	},

	Handler: torrentHandler,
}

func init() {
	registerCommands(&torrentCommand)
	events.RegisterComponentReactionEvent(onComponentReaction)
}

type torrentOptions struct {
	subcommand string                // "add" or "list"
	uri        string                // required for "add"
	query      string                // required for "search"
	category   common.X1337xCategory // optional for "search"
	sort       common.X1337xSort     // optional for "search"
	page       int                   // optional for "search"
}

func parseTorrentOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (torrentOptions, error) {
	results := torrentOptions{}

	subcommand := options[0].Name
	subcommandOptions := options[0].Options

	if subcommand == "add" {
		results.subcommand = subcommand

		for _, option := range subcommandOptions {
			switch option.Name {
			case "uri":
				val, err := utils.CheckOptionStringValue(option)
				if err != nil {
					return results, fmt.Errorf("please enter a URI")
				}
				results.uri = val
			}
		}
	}

	if subcommand == "list" {
		results.subcommand = subcommand
	}

	if subcommand == "search" {
		results.subcommand = subcommand

		for _, opt := range subcommandOptions {
			switch opt.Name {
			case "query":
				val, err := utils.CheckOptionStringValue(opt)
				if err != nil {
					return results, fmt.Errorf("please enter a query")
				}
				results.query = val
			case "category":
				results.category = common.X1337xCategory(opt.IntValue())
			case "sort":
				results.sort = common.X1337xSort(opt.IntValue())
			case "page":
				results.page = int(opt.IntValue())
				if results.page < 1 {
					results.page = 1
				}
			}
		}
	}

	return results, nil
}

func torrentHandler(s *discordgo.Session, i *discordgo.InteractionCreate, appData *discordgo.ApplicationCommandInteractionData) {
	user := utils.GetInteractionAuthor(i.Interaction)

	Log.Debug(Log.Level.Info, `SlashCommand: "torrent", GuildID:`, i.GuildID, "ChannelID:", i.ChannelID, "UserID:", user.ID, "UserName:", user.Username)

	options, err := parseTorrentOptions(appData.Options)
	if err != nil {
		Log.Debug(Log.Level.Error, `parsing "torrent" command options:`, err.Error())
		sendError := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while parsing **torrent** command options:\n`%s`", err.Error()), true)
		if sendError != nil {
			Log.Error("\nTorrent:", sendError.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendError.Error())
		}
		return
	}

	// * ADD
	if options.subcommand == "add" {
		addTorrent(s, i, &options.uri, nil)
		return
	}

	// * LIST
	if options.subcommand == "list" {
		listTorrents(s, i)
		return
	}

	// * SEARCH
	if options.subcommand == "search" {
		searchTorrent(s, i, &options)
	}
}

func onComponentReaction(s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.MessageComponentInteractionData) {
	// add torrent button from yts search command
	if data.CustomID == "yts_torrents" {
		selectedValue := data.Values[0]
		addTorrent(s, i, &selectedValue, nil)
		return
	}

	// stop button while torrent is downloading
	if data.CustomID == "torrent_stop" {
		torrentClient.Stop(false)

		sendErr := interaction.RespondWithNothing(s, i)
		if sendErr != nil {
			Log.Error("\ntorrent:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	// stop and remove button while torrent is downloading
	if data.CustomID == "torrent_stop_remove" {
		torrentClient.Stop(true)

		sendErr := interaction.RespondWithNothing(s, i)
		if sendErr != nil {
			Log.Error("\ntorrent:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	// remove torrent button from torrents list with actions
	if strings.HasPrefix(data.CustomID, "torrent_remove") {
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
		return
	}

	// resume torrent button from torrents list with actions
	if strings.HasPrefix(data.CustomID, "torrent_resume") {
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
		return
	}

	// generate links button from torrents list with actions
	if strings.HasPrefix(data.CustomID, "torrent_links") {
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
		return
	}

	// reshow torrents list from torrents list with actions
	if data.CustomID == "show_torrents_list" {
		msgDeleteErr := s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
		if msgDeleteErr != nil {
			Log.Error("\ntorrentList:", msgDeleteErr.Error())
			Log.Debug(Log.Level.Error, `deleting a message for "torrent" command:`, msgDeleteErr.Error())
		}
		listTorrents(s, i)
		return
	}

	// when selecting a torrent from the torrents list
	if data.CustomID == "torrent_list" {
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

		return
	}

	// Next Page button when using torrent search
	if data.CustomID == "next_page" {
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

		return
	}

	if data.CustomID == "search_list" {
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

		return
	}
}

func addTorrent(s *discordgo.Session, i *discordgo.InteractionCreate, uri *string, tor *torrent.Torrent) {
	sendErr := interaction.RespondWithThinking(s, i, false)
	if sendErr != nil {
		Log.Error("\nTorrent:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		return
	}

	var (
		status func() torrentClient.TorrentInfo
		err    error
	)

	// download from a uri, or resume a torrent
	if uri != nil {
		status, err = torrentClient.Download(*uri)
	} else {
		status, err = torrentClient.Resume(tor)
	}

	// error while starting downloading
	if err != nil {
		Log.Debug(Log.Level.Error, "starting downloading a torrent:", err.Error())
		sendErr := interaction.RespondEdit(s, i, fmt.Sprintf("**Error:** while starting downloading a torrent:\n`%s`", err.Error()))
		if sendErr != nil {
			Log.Error("\nTorrent:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	// print the status
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			state := status()

			if state.Completed {
				sendErr := interaction.RespondEdit(s, i, "Download complete!")
				if sendErr != nil {
					Log.Error("\nTorrent:", sendErr.Error())
					Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
				}

				torrentClient.Stop(false) // make sure the torrent is stopped
				break
			}

			// if there is an error
			if state.Error != nil {
				Log.Debug(Log.Level.Error, "downloading a torrent:", state.Error.Error())
				sendErr := interaction.RespondEdit(s, i, fmt.Sprintf("**Error:** while downloading a torrent:\n`%s`", state.Error.Error()))
				if sendErr != nil {
					Log.Error("\nTorrent:", sendErr.Error())
					Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
				}
				break
			}

			// if the torrent is stopped
			if state.Stopped {
				sendErr := interaction.RespondEdit(s, i, "Download stopped!")
				if sendErr != nil {
					Log.Error("\nTorrent:", sendErr.Error())
					Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
				}
				break
			}

			format := fmt.Sprint(
				"\u200b\n",
				"**Status:** _", state.Status, "_\n",
				"**Progress:** `", state.Progress, "`\n",
				"**Downloaded:** `", state.Downloaded, " / ", state.TotalSize, "`\n",
				"**Download Speed:** `", state.DownloadSpeed+"s", "`\n",
				"**Upload Speed:** `", state.UploadSpeed+"s", "`\n",
				"**Peers:** `", state.Peers, "`\n",
				"**ETA:** `", state.ETA, "`\n\u200b",
			)

			sendErr := interaction.RespondEditWithComponents(s, i, &format,
				components.AddMessageComponents(
					components.NewRow(
						components.NewButton().SetLabel("Stop").SetCustomID("torrent_stop").SetStylePrimary(),
						components.NewButton().SetLabel("Stop and remove").SetCustomID("torrent_stop_remove").SetStyleDanger(),
					),
				),
			)
			if sendErr != nil {
				Log.Error("\nTorrent:", sendErr.Error())
				Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
			}

			<-ticker.C // wait for ticker
		}
	}()
}

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

func searchTorrent(s *discordgo.Session, i *discordgo.InteractionCreate, options *torrentOptions) {
	sendErr := interaction.RespondWithThinking(s, i, false)
	if sendErr != nil {
		Log.Error("\nTorrent:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		return
	}

	searchTmp = nil

	if options.page < 1 {
		options.page = 1
	}

	results := search_1337x(options.query, options.category, options.sort, options.page)

	if len(results) == 0 {
		sendErr = interaction.RespondEdit(s, i, "No results found.")
		if sendErr != nil {
			Log.Error("\nTorrent:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		}
		return
	}

	searchTmp = &SearchTmp{Options: options, Results: results}

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

	content := fmt.Sprintf("Found %d results\n **Page:** `%d`", len(results), options.page)
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

func getVideoUrls(tor *torrent.Torrent) ([]string, error) {
	config := utils.GetAppConfig()

	paths, err := tor.FilePaths()
	if err != nil {
		return nil, err
	}

	var results []string
	for _, filePath := range paths {
		fullPath := filepath.Join(config.Torrent.DownloadDir, filePath)
		isVideo, err := utils.IsVideoFile(fullPath)
		if err != nil {
			return nil, err
		}
		if isVideo {
			videoUrl := url.PathEscape(filePath)
			results = append(results, videoUrl)
		}
	}

	return results, err
}

func search_1337x(query string, category common.X1337xCategory, sort common.X1337xSort, page int) (results []common.SearchResult) {
	if page < 1 {
		page = 1
	}

	query = strings.ReplaceAll(query, " ", "+")
	query = strings.ReplaceAll(query, ".", "+")

	baseUrl := "https://1337x.to"
	searchCMD := "search"
	pageStr := strconv.Itoa(page)

	if category != common.CategoryAll {
		searchCMD = fmt.Sprintf("category-%s", searchCMD)
	}
	if sort != common.SortNone {
		searchCMD = fmt.Sprintf("sort-%s", searchCMD)
	}

	fullUrl, err := url.JoinPath(baseUrl, searchCMD, query, category.Parse(), sort.Parse(), pageStr, "/")

	if err != nil {
		panic(err)
	}

	c := colly.NewCollector()

	c.OnHTML(`a[href^="/torrent"]`, func(e *colly.HTMLElement) {
		href := baseUrl + e.Attr("href")
		results = append(results, common.SearchResult{Url: href, Name: e.Text})
	})

	c.Visit(fullUrl)

	return
}

func getMagnet(url string) (magnet string) {
	c := colly.NewCollector()

	c.OnHTML(`a[href^="magnet:?"]`, func(e *colly.HTMLElement) {
		magnet = e.Attr("href")
	})

	c.Visit(url)

	return
}
