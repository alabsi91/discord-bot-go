package events

import (
	tts "discord-bot/TTS"
	"discord-bot/common"
	"discord-bot/firebase"
	"discord-bot/utils"
	"time"

	"github.com/bwmarrin/discordgo"
)

func init() {
	RegisterOnVoiceStateUpdateEvent(ttsVoiceWelcomeMessage)
}

// ttsVoiceWelcomeMessage speaks out when a custom TTS voice message when a user joins the voice channel
func ttsVoiceWelcomeMessage(s *discordgo.Session, vs *discordgo.VoiceStateUpdate, state common.VoiceState) {
	if state != common.VoiceJoined {
		return
	}

	utils.Log.Debug(utils.Log.Level.Info, "Event: OnVoiceStateUpdate, Handler: ttsVoiceWelcomeMessage, GuildID:", vs.GuildID, "ChannelID:", vs.ChannelID, "UserID:", vs.UserID, "UserName:", vs.Member.User.Username)

	user, err := s.User(vs.UserID)
	if err != nil {
		utils.Log.Error("\nTTSVoiceWelcome:", err.Error())
		utils.Log.Debug(utils.Log.Level.Error, err.Error())
		return
	}

	guildID := vs.GuildID
	if guildID == "" {
		utils.Log.Error("\nTTSVoiceWelcome:", "GuildID is empty")
		utils.Log.Debug(utils.Log.Level.Error, "TTSVoiceWelcome: GuildID is empty")
		return
	}

	guildData, err := firebase.GetGuildData(guildID)
	if err != nil {
		utils.Log.Error("\nTTSVoiceWelcome:", err.Error())
		utils.Log.Debug(utils.Log.Level.Error, err.Error())
		return
	}

	// get user's voice message
	message, notFoundErr := guildData.VoiceMessagesGetItem(user.ID)
	if notFoundErr != nil {
		return
	}

	//  join the user's voice channel
	vc, err := s.ChannelVoiceJoin(guildID, vs.ChannelID, false, true)
	if err != nil {
		utils.Log.Error("\nTTSVoiceWelcome:", err.Error())
		utils.Log.Debug(utils.Log.Level.Error, err.Error())
		return
	}

	time.Sleep(250 * time.Millisecond)
	vc.Speaking(true)

	err = tts.GenerateAndSendToVoiceChannel(message.Message, vc, tts.TTSOptions{Lang: message.Lang, Slow: false})
	if err != nil {
		utils.Log.Error("\nTTSVoiceWelcome:", err.Error())
		utils.Log.Debug(utils.Log.Level.Error, err.Error())
		return
	}

	vc.Speaking(false)
	time.Sleep(250 * time.Millisecond)
	vc.Disconnect()
}
