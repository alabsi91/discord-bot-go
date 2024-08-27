package events

import (
	"discord-bot/firebase"
	"discord-bot/utils"

	"github.com/bwmarrin/discordgo"
)

func init() {
	RegisterOnReadyEvent(botActivity)
}

// botActivity get bot activity from firebase and update status upon ready
func botActivity(s *discordgo.Session, e *discordgo.Ready) {
	utils.Log.Debug(utils.Log.Level.Info, "Event: OnReady, Handler: botActivity")

	// get bot activity from firebase
	botActivity, err := firebase.GetBotActivity()
	if err != nil {
		utils.Log.Error("\nOnReady:", err.Error())
		utils.Log.Debug(utils.Log.Level.Error, err.Error())
	}

	// update status
	err = s.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{{Name: botActivity.Activity, Type: botActivity.ActivityType}},
	})
	if err != nil {
		utils.Log.Error("\nOnReady:", err.Error())
		utils.Log.Debug(utils.Log.Level.Error, err.Error())
	}
}
