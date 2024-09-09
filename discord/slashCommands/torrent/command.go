package torrentCommand

import (
	"discord-bot/common"
	"discord-bot/discord/events"
	"discord-bot/discord/interaction"
	"discord-bot/utils"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/cenkalti/rain/torrent"
)

var Log = &utils.Log

type SearchTmp struct {
	Options    *cmdOptions
	Results    []common.SearchResult
	ChannelID  string
	MsgID      string
	EmbedMsgID string
}

// store the key as [guildID+channelID]
var searchTmp = map[string]*SearchTmp{}

var command = common.SlashCommand{
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

	Handler: cmdHandler,
}

func init() {
	events.RegisterSlashCommand(&command)
	events.RegisterComponentReactionEvent(onComponentReaction)
}

type cmdOptions struct {
	subcommand string                // "add" or "list"
	uri        string                // required for "add"
	query      string                // required for "search"
	category   common.X1337xCategory // optional for "search"
	sort       common.X1337xSort     // optional for "search"
	page       int                   // optional for "search"
}

func parseCmdOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (cmdOptions, error) {
	results := cmdOptions{}

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

func cmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate, appData *discordgo.ApplicationCommandInteractionData) {
	user := utils.GetInteractionAuthor(i.Interaction)

	Log.Debug(Log.Level.Info, `SlashCommand: "torrent", GuildID:`, i.GuildID, "ChannelID:", i.ChannelID, "UserID:", user.ID, "UserName:", user.Username)

	options, err := parseCmdOptions(appData.Options)
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
	if data.CustomID == "yts_torrents" {
		ytsListOnSelect(data, s, i)
		return
	}

	if data.CustomID == "torrent_stop" {
		torrentStopButton(s, i)
		return
	}

	if data.CustomID == "torrent_stop_remove" {
		torrentStopAndRemoveButton(s, i)
		return
	}

	if strings.HasPrefix(data.CustomID, "torrent_remove") {
		TorrentDeleteButton(data, s, i)
		return
	}

	if strings.HasPrefix(data.CustomID, "torrent_resume") {
		torrentResumeButton(data, s, i)
		return
	}

	if strings.HasPrefix(data.CustomID, "torrent_links") {
		generateLinksButton(data, s, i)
		return
	}

	if data.CustomID == "show_torrents_list" {
		showTorrentListButton(s, i)
		return
	}

	if data.CustomID == "torrent_list" {
		onTorrentListSelect(s, i, data)
		return
	}

	if data.CustomID == "search_list" {
		onSearchListSelect(data, s, i)
		return
	}

	if data.CustomID == "next_page" {
		nextPageButton(s, i)
		return
	}

	if strings.HasPrefix(data.CustomID, "search_download") {
		searchDownloadButton(data, s, i)
		return
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
