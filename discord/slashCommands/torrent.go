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
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cenkalti/rain/torrent"
)

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
		},
	},

	Handler: torrentHandler,
}

func init() {
	registerCommands(&torrentCommand)
	events.RegisterComponentReactionEvent(onComponentReaction)
}

type torrentOptions struct {
	subcommand string // "add" or "list"
	uri        string // required for "add"
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
