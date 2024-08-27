package firebase

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type (
	VoiceWelcomeMessage struct {
		Id      string
		Message string
		Lang    string
	}
	CustomCommand struct {
		When string
		Say  string
	}
	BotActivity struct {
		Activity     string
		ActivityType discordgo.ActivityType
	}
)

type FirebaseData struct {
	VoiceMessages  []VoiceWelcomeMessage
	CustomCommands []CustomCommand
	SavedList      []string
	Prefix         string
}

// * MARK: Voice Messages

func (data *FirebaseData) VoiceMessagesIsInList(userId string) bool {
	for _, v := range data.VoiceMessages {
		if v.Id == userId {
			return true
		}
	}
	return false
}

func (data *FirebaseData) VoiceMessagesAddItem(item VoiceWelcomeMessage) {
	data.VoiceMessages = append(data.VoiceMessages, item)
}

func (data *FirebaseData) VoiceMessagesRemoveItem(userId string) {
	for i, v := range data.VoiceMessages {
		if v.Id == userId {
			data.VoiceMessages = append(data.VoiceMessages[:i], data.VoiceMessages[i+1:]...)
		}
	}
}

func (data *FirebaseData) VoiceMessagesUpdateItem(item VoiceWelcomeMessage) {
	for i, v := range data.VoiceMessages {
		if v.Id == item.Id {
			data.VoiceMessages[i].Message = item.Message
			data.VoiceMessages[i].Lang = item.Lang
		}
	}
}

func (data *FirebaseData) VoiceMessagesGetItem(userId string) (*VoiceWelcomeMessage, error) {
	for _, voiceMessage := range data.VoiceMessages {
		if voiceMessage.Id == userId {
			return &voiceMessage, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

func (data *FirebaseData) VoiceMessagesToMap() []map[string]interface{} {
	voiceMessagesMap := make([]map[string]interface{}, len(data.VoiceMessages))

	for i, v := range data.VoiceMessages {
		voiceMessagesMap[i] = map[string]interface{}{
			"id":      v.Id,
			"message": v.Message,
			"lang":    v.Lang,
		}
	}

	return voiceMessagesMap
}

func (data *FirebaseData) VoiceMessagesFromMap(mapArr []interface{}) []VoiceWelcomeMessage {
	voiceMessagesArr := make([]VoiceWelcomeMessage, len(mapArr))

	for i, v := range mapArr {
		voiceMessagesArr[i] = VoiceWelcomeMessage{
			Id:      v.(map[string]interface{})["id"].(string),
			Message: v.(map[string]interface{})["message"].(string),
			Lang:    v.(map[string]interface{})["lang"].(string),
		}
	}

	return voiceMessagesArr
}

// * MARK: Custom Commands

func (data *FirebaseData) CustomCommandsIsInList(trigger string) bool {
	for _, v := range data.CustomCommands {
		if v.When == trigger {
			return true
		}
	}
	return false
}

func (data *FirebaseData) CustomCommandsGetItem(trigger string) (*CustomCommand, error) {
	for _, customCommand := range data.CustomCommands {
		if customCommand.When == trigger {
			return &customCommand, nil
		}
	}

	return nil, fmt.Errorf("trigger not found")
}

func (data *FirebaseData) CustomCommandsAddItem(item CustomCommand) {
	data.CustomCommands = append(data.CustomCommands, item)
}

func (data *FirebaseData) CustomCommandsRemoveItem(trigger string) {
	for i, v := range data.CustomCommands {
		if v.When == trigger {
			data.CustomCommands = append(data.CustomCommands[:i], data.CustomCommands[i+1:]...)
		}
	}
}

func (data *FirebaseData) CustomCommandsUpdateItem(item CustomCommand) {
	for i, v := range data.CustomCommands {
		if v.When == item.When {
			data.CustomCommands[i].Say = item.Say
		}
	}
}

func (data *FirebaseData) CustomCommandsToMap() []map[string]interface{} {
	customCommandsMap := make([]map[string]interface{}, len(data.CustomCommands))

	for i, v := range data.CustomCommands {
		customCommandsMap[i] = map[string]interface{}{
			"when": v.When,
			"say":  v.Say,
		}
	}

	return customCommandsMap
}

func (data *FirebaseData) CustomCommandsFromMap(mapArr []interface{}) []CustomCommand {
	customCommandsArr := make([]CustomCommand, len(mapArr))

	for i, v := range mapArr {
		customCommandsArr[i] = CustomCommand{
			When: v.(map[string]interface{})["when"].(string),
			Say:  v.(map[string]interface{})["say"].(string),
		}
	}

	return customCommandsArr
}

// * MARK: Saved List

func (data *FirebaseData) SavedListIsInList(item string) bool {
	for _, v := range data.SavedList {
		if v == item {
			return true
		}
	}
	return false
}

func (data *FirebaseData) SavedListAddItem(item string) {
	data.SavedList = append(data.SavedList, item)
}

func (data *FirebaseData) SavedListRemoveItem(item string) {
	for i, v := range data.SavedList {
		if v == item {
			data.SavedList = append(data.SavedList[:i], data.SavedList[i+1:]...)
		}
	}
}

func (data *FirebaseData) SavedListFromMap(mapArr []interface{}) []string {
	savedListArr := make([]string, len(mapArr))

	for i, v := range mapArr {
		savedListArr[i] = v.(string)
	}

	return savedListArr
}

// * MARK: Data

func (data *FirebaseData) SetDefaults() {
	data.VoiceMessages = []VoiceWelcomeMessage{}
	data.CustomCommands = []CustomCommand{}
	data.SavedList = []string{}
	data.Prefix = "!"
}

func (data *FirebaseData) CreateFromMap(mapData map[string]interface{}) {
	if voiceMessages, ok := mapData["voiceMessages"]; ok {
		data.VoiceMessages = data.VoiceMessagesFromMap(voiceMessages.([]interface{}))
	}

	if customCommands, ok := mapData["customCommands"]; ok {
		data.CustomCommands = data.CustomCommandsFromMap(customCommands.([]interface{}))
	}

	if savedList, ok := mapData["savedList"]; ok {
		data.SavedList = data.SavedListFromMap(savedList.([]interface{}))
	}

	if prefix, ok := mapData["prefix"]; ok {
		data.Prefix = prefix.(string)
	}
}
