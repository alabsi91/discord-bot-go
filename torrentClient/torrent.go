package torrentClient

import (
	"discord-bot/utils"
	"fmt"
	"math"
	"path/filepath"
	"time"

	"github.com/cenkalti/rain/torrent"
)

var Log = utils.Log

type TorrentInfo struct {
	Name          string
	Status        string
	Downloaded    string
	Uploaded      string
	TotalSize     string
	Progress      string
	DownloadSpeed string
	UploadSpeed   string
	Peers         string
	ETA           string
	Completed     bool
	Stopped       bool
	Error         error
}

var session *torrent.Session
var currentTorID string = ""

func Initialize() {
	var err error

	config := utils.GetAppConfig()

	torrent.DisableLogging()

	sessionConfig := torrent.DefaultConfig
	sessionConfig.DataDir = config.Torrent.DownloadDir
	sessionConfig.Database = filepath.Join(config.Torrent.DownloadDir, "torrents.db")
	sessionConfig.DataDirIncludesTorrentID = false
	sessionConfig.ResumeOnStartup = false

	session, err = torrent.NewSession(sessionConfig)

	if err != nil {
		Log.Debug(Log.Level.Fatal, err.Error())
		Log.Fatal("\nInitialize Torrent Session:", err.Error())
	}
}

func CloseSession() {
	session.Close()
}

func Download(uri string) (func() TorrentInfo, error) {
	currentTor := session.GetTorrent(currentTorID)
	if currentTor != nil {
		currentTor.Stop()
	}

	// Add magnet link
	tor, err := session.AddURI(uri, &torrent.AddTorrentOptions{StopAfterDownload: true})
	if err != nil {
		return nil, err
	}

	currentTorID = tor.ID()

	return generateStatsHandler(tor), nil
}

func Stop(remove bool) {
	currentTor := session.GetTorrent(currentTorID)
	if currentTor != nil {
		currentTor.Stop()
		if remove {
			session.RemoveTorrent(currentTorID)
		}
		currentTorID = ""
	}
}

func GetAllTorrents() []*torrent.Torrent {
	return session.ListTorrents()
}

func Remove(id string) error {
	return session.RemoveTorrent(id)
}

func Exists(id string) bool {
	return session.GetTorrent(id) != nil
}

func Resume(tor *torrent.Torrent) (func() TorrentInfo, error) {
	currentTor := session.GetTorrent(currentTorID)
	if currentTor != nil {
		currentTor.Stop()
	}

	err := tor.Start()
	if err != nil {
		return nil, err
	}

	currentTorID = tor.ID()

	return generateStatsHandler(tor), nil
}

func GetMagnetLink(id string) (string, error) {
	t := session.GetTorrent(id)
	if t == nil {
		return "", fmt.Errorf("torrent %s not found", id)
	}
	return t.Magnet()
}

func GetTorrentName(id string) string {
	t := session.GetTorrent(id)
	if t == nil {
		return ""
	}
	return t.Name()
}

func GetTorrentByID(id string) (*torrent.Torrent, error) {
	t := session.GetTorrent(id)
	if t == nil {
		return nil, fmt.Errorf("torrent %s not found", id)
	}
	return t, nil
}

func generateStatsHandler(tor *torrent.Torrent) func() TorrentInfo {
	return func() TorrentInfo {
		s := tor.Stats()

		eta := "0"
		if s.ETA != nil {
			eta = formatTime(*s.ETA)
		}

		progress := float64(s.Bytes.Completed) / float64(s.Bytes.Total) * 100
		if math.IsNaN(progress) {
			progress = 0
		}

		isSeeding := s.Status == torrent.Seeding

		return TorrentInfo{
			Name:          s.Name,
			Status:        s.Status.String(),
			Downloaded:    formatBytes(s.Bytes.Completed),
			Uploaded:      formatBytes(s.Bytes.Uploaded),
			TotalSize:     formatBytes(s.Bytes.Total),
			Progress:      fmt.Sprintf("%.2f%%", progress),
			DownloadSpeed: formatBytes(int64(s.Speed.Download)),
			UploadSpeed:   formatBytes(int64(s.Speed.Upload)),
			Peers:         fmt.Sprintf("%d", s.Peers.Total),
			ETA:           eta,
			Completed:     isSeeding || (s.Bytes.Total != 0 && s.Bytes.Completed == s.Bytes.Total),
			Stopped:       s.Status == torrent.Stopped,
			Error:         s.Error,
		}
	}
}

func formatBytes(bytes int64) string {
	const unit = 1000
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatTime(d time.Duration) string {
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour

	hours := d / time.Hour
	d -= hours * time.Hour

	minutes := d / time.Minute
	d -= minutes * time.Minute

	seconds := d / time.Second

	// Build the formatted string
	var result string
	if days > 0 {
		result += fmt.Sprintf("%dd ", days)
	}
	if hours > 0 || days > 0 {
		result += fmt.Sprintf("%dh ", hours)
	}
	if minutes > 0 || hours > 0 || days > 0 {
		result += fmt.Sprintf("%dm ", minutes)
	}
	result += fmt.Sprintf("%ds", seconds)

	return result
}
