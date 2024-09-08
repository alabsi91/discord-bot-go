package torrentCommand

import (
	"discord-bot/discord/components"
	"discord-bot/discord/interaction"
	"discord-bot/torrentClient"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cenkalti/rain/torrent"
)

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
				state.Name, "_\n",
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
