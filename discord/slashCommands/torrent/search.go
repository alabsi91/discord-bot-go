package torrentCommand

import (
	"discord-bot/common"
	"discord-bot/discord/components"
	"discord-bot/discord/interaction"
	"fmt"
	"net/url"
	"strconv"
	"strings"

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
