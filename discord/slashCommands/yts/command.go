package slashCommands

import (
	"discord-bot/common"
	"discord-bot/discord/components"
	"discord-bot/discord/events"
	"discord-bot/discord/interaction"
	"discord-bot/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var Log = &utils.Log

var command = common.SlashCommand{
	Command: discordgo.ApplicationCommand{
		Name:        "yts",
		Description: "Search for a movie on yts database",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "movie_name",
				Description: "Enter the movie name to search for",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	},

	Handler: cmdHandler,
}

type SearchTmp struct {
	Movies     []common.Movie
	ChannelID  string
	MsgID      string
	EmbedMsgID string
}

// the key is [guildID+channelID]
var searchTmp = map[string]*SearchTmp{}

func init() {
	events.RegisterSlashCommand(&command)
	events.RegisterComponentReactionEvent(ytsOnSelect)
}

type cmdOptions struct {
	movie_name string // required
}

func parseCmdOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (cmdOptions, error) {
	results := cmdOptions{}

	for _, opt := range options {

		if opt.Name == "movie_name" {
			val, err := utils.CheckOptionStringValue(opt)
			if err != nil {
				return results, fmt.Errorf("please enter a movie name")
			}
			results.movie_name = val

			break // only one field is required
		}
	}

	return results, nil
}

func cmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate, appData *discordgo.ApplicationCommandInteractionData) {
	user := utils.GetInteractionAuthor(i.Interaction)

	Log.Debug(Log.Level.Info, `SlashCommand: "yts", GuildID:`, i.GuildID, "ChannelID:", i.ChannelID, "UserID:", user.ID, "UserName:", user.Username)

	options, err := parseCmdOptions(appData.Options)
	if err != nil {
		Log.Debug(Log.Level.Error, `parsing "yts" command options:`, err.Error())
		sendErr := interaction.RespondWithText(s, i, fmt.Sprintf("**Error:** while parsing **yts** command options:\n`%s`", err.Error()), true)
		if sendErr != nil {
			Log.Error("\nYTS:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "yts" command:`, sendErr.Error())
		}
		return
	}

	sendErr := interaction.RespondWithThinking(s, i, false)
	if sendErr != nil {
		Log.Error("\nYTS:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "yts" command:`, sendErr.Error())
		return
	}

	// remove previous messages
	search, ok := searchTmp[i.GuildID+i.ChannelID]
	if ok {
		err := s.ChannelMessageDelete(search.ChannelID, search.MsgID)
		if err != nil {
			Log.Error("\nTorrent:", err.Error())
			Log.Debug(Log.Level.Error, "deleting a message:", err.Error())
		}
	}

	ytsResponse, err := fetchYtsMovies(options.movie_name)
	if err != nil {
		Log.Debug(Log.Level.Error, "fetching yts search data:", err.Error())
		sendErr := interaction.RespondEdit(s, i, fmt.Sprintf("**Error:** while fetching yts search data:\n`%s`", err.Error()))
		if sendErr != nil {
			Log.Error("\nYTS:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "yts" command:`, sendErr.Error())
		}
		return
	}

	moviesLength := len(ytsResponse.Data.Movies)

	if moviesLength == 0 {
		sendErr := interaction.RespondEdit(s, i, "No results found for:", options.movie_name)
		if sendErr != nil {
			Log.Error("\nYTS:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "yts" command:`, sendErr.Error())
		}
		return
	}

	msg, err := createSelectMenu(s, i, ytsResponse)
	if err != nil {
		Log.Error("\nYTS:", err.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "yts" command:`, err.Error())
	}

	// save for later when the user selects a movie from the select menu
	searchTmp[i.GuildID+i.ChannelID] = &SearchTmp{Movies: ytsResponse.Data.Movies, ChannelID: i.ChannelID, MsgID: msg.ID}

	go func() {
		<-time.After(time.Minute)

		search, ok := searchTmp[i.GuildID+i.ChannelID]
		if !ok {
			return
		}

		// remove embed message also
		if search.EmbedMsgID != "" {
			err := s.ChannelMessageDelete(search.ChannelID, search.EmbedMsgID)
			if err != nil {
				Log.Error("\nTorrent:", err.Error())
				Log.Debug(Log.Level.Error, "deleting a message:", err.Error())
			}
		}

		err := s.ChannelMessageDelete(search.ChannelID, search.MsgID)
		if err != nil {
			Log.Error("\nTorrent:", err.Error())
			Log.Debug(Log.Level.Error, "deleting a message:", err.Error())
		}

		delete(searchTmp, i.GuildID+i.ChannelID)
	}()
}

// ytsOnSelect is called when the user selects a movie from the select menu
func ytsOnSelect(s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.MessageComponentInteractionData) {
	if data.CustomID != "yts" {
		return
	}

	selectedValue := data.Values[0]

	movieIndex, err := strconv.Atoi(selectedValue)
	if err != nil {
		Log.Error("\nYTS:", err.Error())
		Log.Debug(Log.Level.Error, `converting a string value to int "yts" on select:`, err.Error())
		return
	}

	search, ok := searchTmp[i.GuildID+i.ChannelID]

	if !ok || search.Movies == nil || movieIndex < 0 || movieIndex >= len(search.Movies) {
		sendErr := interaction.RespondWithText(s, i, "**Error:** No results found", true)
		if sendErr != nil {
			Log.Error("\nYTS:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "yts" command:`, sendErr.Error())
		}
		return
	}

	selectedMovie := &search.Movies[movieIndex]

	generateLinksOptions := func() []*components.SelectMenuOption {
		var options []*components.SelectMenuOption
		for _, torrent := range selectedMovie.Torrents {
			label := fmt.Sprintf("%s.%s(%s)", torrent.Quality, torrent.Type, torrent.Size)
			options = append(options, components.NewMenuOption().SetLabel(label).SetValue(torrent.URL))
		}
		return options
	}

	// should edit or send new embed
	shouldEdit := false
	messageId := ""
	messages, err := s.ChannelMessages(i.ChannelID, 1, "", "", "")
	if err == nil && len(messages) > 0 {
		embeds := messages[0].Embeds
		for _, embeds := range embeds {
			if strings.Contains(embeds.Footer.Text, "yts.mx") {
				shouldEdit = true
				messageId = messages[0].ID
				break
			}
		}
	}

	// check if should we respond with new embed or update the current
	if shouldEdit {
		sendErr := interaction.RespondWithNothing(s, i)
		if sendErr != nil {
			Log.Error("\nYTS:", sendErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "yts" command:`, sendErr.Error())
			return
		}

		// update the embed
		content := "_This message will be deleted also._"
		_, editMsgErr := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel: i.ChannelID,
			ID:      messageId,
			Content: &content,
			Embeds:  &[]*discordgo.MessageEmbed{createMovieEmbed(selectedMovie)},
			Components: components.AddMessageComponents(
				components.NewRow(
					components.NewSelectMenu().SetStringType().SetCustomID("yts_torrents").
						SetPlaceholder("Select a torrent to download").
						SetOptions(generateLinksOptions()...),
				),
			),
		})
		if editMsgErr != nil {
			Log.Error("\nYTS:", editMsgErr.Error())
			Log.Debug(Log.Level.Error, `sending a respond for "yts" command:`, editMsgErr.Error())
		}
		return
	}

	// Respond with embed
	content := "_This message will be deleted also._"
	sendErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Embeds:  []*discordgo.MessageEmbed{createMovieEmbed(selectedMovie)},
			Components: *components.AddMessageComponents(
				components.NewRow(
					components.NewSelectMenu().SetStringType().SetCustomID("yts_torrents").
						SetPlaceholder("Select a torrent to download").
						SetOptions(generateLinksOptions()...),
				),
			),
		},
	})
	if sendErr != nil {
		Log.Error("\nYTS:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "yts" command:`, sendErr.Error())
		return
	}

	// Get the message ID to delete later
	msg, sendErr := s.InteractionResponse(i.Interaction)
	if sendErr != nil {
		Log.Error("\nYTS:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "yts" command:`, sendErr.Error())
		return
	}

	search.EmbedMsgID = msg.ID
}

// fetchYtsMovies search the yts database for a movie by name
func fetchYtsMovies(movieName string) (*common.YtsResponse, error) {
	movieName = strings.ToLower(movieName)
	movieName = strings.TrimSpace(movieName)
	movieName = strings.ReplaceAll(movieName, " ", "+")

	apiURL := fmt.Sprintf("https://yts.mx/api/v2/list_movies.json?query_term=%s", url.QueryEscape(movieName))

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ytsResponse common.YtsResponse
	err = json.Unmarshal(body, &ytsResponse)
	if err != nil {
		return nil, err
	}

	return &ytsResponse, nil
}

// createSelectMenu send a select menu of movies to the user
func createSelectMenu(s *discordgo.Session, i *discordgo.InteractionCreate, ytsData *common.YtsResponse) (*discordgo.Message, error) {
	length := len(ytsData.Data.Movies)

	menuOptions := make([]*components.SelectMenuOption, length)
	for i := 0; i < length; i++ {
		label := fmt.Sprintf("%d. %s", i+1, ytsData.Data.Movies[i].TitleLong)
		value := fmt.Sprintf("%d", i)
		menuOptions[i] = components.NewMenuOption().SetLabel(label).SetValue(value)
	}

	content := "_This message will be deleted in `1` minute._\n\u200b\n"
	content += fmt.Sprintf("Found `%d` results", length)

	msg, sendErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Components: components.AddMessageComponents(
			components.NewRow(
				components.NewSelectMenu().
					SetPlaceholder("Select a movie").
					SetCustomID("yts").
					SetStringType().
					SetOptions(menuOptions...),
			),
		),
	})

	return msg, sendErr
}

// createMovieEmbed create the embed for a specific movie
func createMovieEmbed(movie *common.Movie) *discordgo.MessageEmbed {
	embed := components.NewEmbed().
		SetColor(0x0099ff).
		SetTitle(movie.TitleLong).
		SetURL(movie.URL).
		SetDescription(movie.Summary).
		SetThumbnail(movie.SmallCoverImage).
		AddEmptyField().
		AddField("Rating:", fmt.Sprintf(":star:   %.1f", movie.Rating), true).
		AddField("Genres:", ":movie_camera:  "+strings.Join(movie.Genres, ", "), true).
		AddField("Runtime:", fmt.Sprintf(":clock1: %d minutes", movie.Runtime), true).
		AddEmptyField().
		AddField("Trailer:", fmt.Sprintf(":link:   [Youtube](https://www.youtube.com/watch?v=%s)", movie.YTtrailerCode), true).
		AddField("IMDb:", fmt.Sprintf(":link:   [IMDb](https://www.imdb.com/title/%s)", movie.IMDBCode), true).
		AddEmptyField()

	for _, torrent := range movie.Torrents {
		embed = embed.AddField("Torrent link:", fmt.Sprintf("[%s.%s(%s)](%s)", torrent.Quality, torrent.Type, torrent.Size, torrent.URL), true)
	}

	embed = embed.SetImage(movie.LargeCoverImage).
		SetTimestamp(time.Now().Format(time.RFC3339)).
		SetFooter("https://yts.mx")

	return embed.Into()
}
