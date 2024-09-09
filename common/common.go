package common

import "github.com/bwmarrin/discordgo"

type VoiceState = int

const (
	VoiceStateNone = iota
	VoiceJoined
	VoiceLeft
	VoiceSwitched
)

type DmsCommand struct {
	Name        string
	Subcommands []string
	Handler     func(s *discordgo.Session, m *discordgo.MessageCreate, subcommand string, args *[]string)
}

type SlashCommand struct {
	Command discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate, appData *discordgo.ApplicationCommandInteractionData)
}

type Torrent struct {
	URL              string `json:"url"`
	Hash             string `json:"hash"`
	Quality          string `json:"quality"`
	Type             string `json:"type"`
	Seeds            int    `json:"seeds"`
	Peers            int    `json:"peers"`
	Size             string `json:"size"`
	DateUploaded     string `json:"date_uploaded"`
	DateUploadedUnix int    `json:"date_uploaded_unix"`
}

type Movie struct {
	ID                      int       `json:"id"`
	URL                     string    `json:"url"`
	IMDBCode                string    `json:"imdb_code"`
	Title                   string    `json:"title"`
	TitleEnglish            string    `json:"title_english"`
	TitleLong               string    `json:"title_long"`
	Slug                    string    `json:"slug"`
	Year                    int       `json:"year"`
	Rating                  float64   `json:"rating"`
	Runtime                 int       `json:"runtime"`
	Genres                  []string  `json:"genres"`
	Summary                 string    `json:"summary"`
	DescriptionFull         string    `json:"description_full"`
	Synopsis                string    `json:"synopsis"`
	YTtrailerCode           string    `json:"yt_trailer_code"`
	Language                string    `json:"language"`
	MPARating               string    `json:"mpa_rating"`
	BackgroundImage         string    `json:"background_image"`
	BackgroundImageOriginal string    `json:"background_image_original"`
	SmallCoverImage         string    `json:"small_cover_image"`
	MediumCoverImage        string    `json:"medium_cover_image"`
	LargeCoverImage         string    `json:"large_cover_image"`
	State                   string    `json:"state"`
	Torrents                []Torrent `json:"torrents"`
	DateUploaded            string    `json:"date_uploaded"`
	DateUploadedUnix        int       `json:"date_uploaded_unix"`
}

type YtsResponse struct {
	Status        string `json:"status"`
	StatusMessage string `json:"status_message"`
	Meta          struct {
		ServerTime     int64  `json:"server_time"`
		ServerTimeZone string `json:"server_timezone"`
		ApiVersion     int    `json:"api_version"`
		ExecutionTime  string `json:"execution_time"`
	} `json:"@meta"`
	Data struct {
		MovieCount int     `json:"movie_count"`
		Limit      int     `json:"limit"`
		PageNumber int     `json:"page_number"`
		Movies     []Movie `json:"movies"`
	} `json:"data"`
}

type Meme struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Url      string `json:"url"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	BoxCount int    `json:"box_count"`
	Captions int    `json:"captions"`
}

type MemeResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Memes []Meme `json:"memes"`
	} `json:"data"`
}

type Config struct {
	Log struct {
		Enabled bool   `json:"enabled"`
		Path    string `json:"path"`
	} `json:"log"`

	Torrent struct {
		DownloadDir string `json:"downloadDir"`
		ZipDir      string `json:"zipDir"`
	} `json:"torrent"`

	Http struct {
		Domain string `json:"domain"`
		Host   string `json:"host"`
		Port   int    `json:"port"`
		Routes struct {
			Video string `json:"video"`
			Zip   string `json:"zip"`
		} `json:"routes"`
	} `json:"http"`
}

type X1337xCategory int

func (x X1337xCategory) String() string {
	categories := []string{
		"All",
		"Movies",
		"TV",
		"Games",
		"Music",
		"Apps",
		"Documentaries",
		"Anime",
		"Other",
		"XXX",
	}

	if int(x) < 0 || int(x) >= len(categories) {
		return ""
	}

	return categories[x]
}

func (x X1337xCategory) Parse() string {
	categories := []string{
		"",
		"Movies",
		"TV",
		"Games",
		"Music",
		"Apps",
		"Documentaries",
		"Anime",
		"Other",
		"XXX",
	}

	if int(x) < 0 || int(x) >= len(categories) {
		return ""
	}

	return categories[x]
}

const (
	CategoryAll X1337xCategory = iota
	CategoryMovie
	CategoryTV
	CategoryGames
	CategoryMusic
	CategoryApplications
	CategoryDocumentaries
	CategoryAnime
	CategoryOther
	CategoryXXX
)

type X1337xSort int

func (x X1337xSort) String() string {
	sorts := []string{
		"None",
		"time Desc",
		"time Asc",
		"Size Desc",
		"Size Asc",
		"Seeders Desc",
		"Seeders Asc",
		"Leechers Desc",
		"Leechers Asc",
	}

	if int(x) < 0 || int(x) >= len(sorts) {
		return ""
	}

	return sorts[x]
}

func (x X1337xSort) Parse() string {
	sorts := []string{
		"",
		"time/desc",
		"time/asc",
		"size/desc",
		"size/asc",
		"seeders/desc",
		"seeders/asc",
		"leechers/desc",
		"leechers/asc",
	}

	if int(x) < 0 || int(x) >= len(sorts) {
		return ""
	}

	return sorts[x]
}

const (
	SortNone X1337xSort = iota
	SortTimeDesc
	SortTimeAsc
	SortSizeDesc
	SortSizeAsc
	SortSeedersDesc
	SortSeedersAsc
	SortLeechersDesc
	SortLeechersAsc
)

type SearchResult struct {
	Name    string
	Url     string
	Size    string
	Seeds   string
	Leeches string
	Date    string
}
