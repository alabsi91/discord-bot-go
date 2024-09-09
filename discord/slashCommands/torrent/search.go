package torrentCommand

import (
	"discord-bot/common"
	"discord-bot/discord/components"
	"discord-bot/discord/interaction"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly/v2"
)

func searchTorrent(s *discordgo.Session, i *discordgo.InteractionCreate, options *cmdOptions) {
	sendErr := interaction.RespondWithThinking(s, i, false)
	if sendErr != nil {
		Log.Error("\nTorrent:", sendErr.Error())
		Log.Debug(Log.Level.Error, `sending a respond for "torrent" command:`, sendErr.Error())
		return
	}

	// remove previous search message in this channel
	search, ok := searchTmp[i.GuildID+i.ChannelID]
	if ok {
		err := s.ChannelMessageDelete(search.ChannelID, search.MsgID)
		if err != nil {
			Log.Error("\nTorrent:", err.Error())
			Log.Debug(Log.Level.Error, "deleting a message:", err.Error())
		}
	}

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

	searchTmp[i.GuildID+i.ChannelID] = &SearchTmp{Options: options, Results: results, ChannelID: i.ChannelID}

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
	content += fmt.Sprintf("**Page:** `%d` has `%d` results:\n ", options.page, len(results))
	if len(results) > 25 {
		content += "\n (Can only show 25 torrents at once)"
	}

	msg, sendErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Components: components.AddMessageComponents(
			components.NewRow(
				components.NewSelectMenu().SetStringType().SetPlaceholder("Select a result to show details").
					SetCustomID("search_list").
					SetOptions(menuOptions...),
			),
			components.NewRow(
				components.NewButton().SetLabel("Next Page").SetCustomID("next_page"),
			),
		),
	})
	if sendErr != nil {
		Log.Error("\nTorrent:", `sending a respond for "torrent" command:`, sendErr.Error())
	}

	searchTmp[i.GuildID+i.ChannelID].MsgID = msg.ID

	// delete self after 1 minute
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

		delete(searchTmp, i.GuildID+i.ChannelID) // reset
	}()
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

	// URL and Name
	c.OnHTML(`a[href^="/torrent"]`, func(e *colly.HTMLElement) {
		href := baseUrl + e.Attr("href")
		results = append(results, common.SearchResult{Url: href, Name: e.Text})
	})

	// Size
	sizeIndex := 0
	c.OnHTML(`.coll-4.size`, func(e *colly.HTMLElement) {
		if sizeIndex >= len(results) {
			return
		}
		size := e.DOM.Contents().Not("span").Text()
		results[sizeIndex].Size = size
		sizeIndex++
	})

	// Seeds
	seedsIndex := 0
	c.OnHTML(`.coll-2.seeds`, func(e *colly.HTMLElement) {
		if seedsIndex >= len(results) {
			return
		}
		results[seedsIndex].Seeds = e.Text
		seedsIndex++
	})

	// Leeches
	leechesIndex := 0
	c.OnHTML(`.coll-3.leeches`, func(e *colly.HTMLElement) {
		if leechesIndex >= len(results) {
			return
		}
		results[leechesIndex].Leeches = e.Text
		leechesIndex++
	})

	// Date
	dateIndex := 0
	c.OnHTML(`tbody .coll-date`, func(e *colly.HTMLElement) {
		if dateIndex >= len(results) {
			return
		}
		results[dateIndex].Date = e.Text
		dateIndex++
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
