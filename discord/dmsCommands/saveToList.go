package dmsCommands

import (
	"discord-bot/common"
	"discord-bot/discord/components"
	"discord-bot/discord/events"
	"discord-bot/firebase"
	"discord-bot/utils"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var saveToListCommand = common.DmsCommand{
	Name:        "list",
	Subcommands: []string{saveToListSubcommands.Save, saveToListSubcommands.Get, saveToListSubcommands.Remove},
	Handler:     saveToListHandler,
}

func init() {
	registerDmsCommands(&saveToListCommand)
	events.RegisterComponentReactionEvent(saveListOnSelect)
}

var saveToListSubcommands = struct {
	Save   string
	Get    string
	Remove string
}{"save", "get", "remove"}

func saveToListHandler(s *discordgo.Session, m *discordgo.MessageCreate, subcommand string, args *[]string) {
	utils.Log.Debug(Log.Level.Info, "Command: list, subcommand:", subcommand, "GuildID:", m.GuildID, "ChannelID:", m.ChannelID, "UserID:", m.Author.ID, "UserName:", m.Author.Username, "Content:", m.Content)

	// save
	if subcommand == saveToListSubcommands.Save {
		saveSubCommand(s, m)
		return
	}

	// get
	if subcommand == saveToListSubcommands.Get {
		getSubCommand(s, m)
		return
	}

	// remove
	if subcommand == saveToListSubcommands.Remove {
		removeSubCommand(s, m)
		return
	}

	// not a valid subcommand
	_, sendError := s.ChannelMessageSendReply(
		m.ChannelID,
		"You need to specify a subcommand: "+saveToListSubcommands.Save+", "+saveToListSubcommands.Get+", "+saveToListSubcommands.Remove,
		m.Reference(),
	)
	if sendError != nil {
		Log.Error("\nSaveToList:", sendError.Error())
		Log.Debug(Log.Level.Error, sendError.Error())
	}
}

func saveListOnSelect(s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.MessageComponentInteractionData) {
	if data.CustomID != "saveToList_remove" {
		return
	}

	// response to the interaction, (shut up)
	interactionErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if interactionErr != nil {
		Log.Error("\nSaveToList:", interactionErr.Error())
		Log.Debug(Log.Level.Error, interactionErr.Error())
		return
	}

	// remove the select menu component from channel
	content := "-"
	_, sendError := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    i.ChannelID,
		ID:         i.Message.ID,
		Content:    &content,
		Components: &[]discordgo.MessageComponent{},
	})
	if sendError != nil {
		Log.Error("\nSaveToList:", sendError.Error())
		Log.Debug(Log.Level.Error, sendError.Error())
		return
	}

	selectedValue := data.Values[0]

	// string to number
	messageIndex, err := strconv.Atoi(selectedValue)
	if err != nil {
		Log.Error("\nSaveToList:", err.Error())
		Log.Debug(Log.Level.Error, err.Error())
		return
	}

	guildData, err := firebase.GetGuildData(i.GuildID)
	if err != nil { // shouldn't happen at this point
		Log.Error("\nSaveToList:", err.Error())
		Log.Debug(Log.Level.Error, err.Error())
		return
	}

	// check if index is valid
	if messageIndex < 0 || messageIndex >= len(guildData.SavedList) {
		_, sendError := s.ChannelMessageSend(i.ChannelID, "❗ Invalid message index.")
		if sendError != nil {
			Log.Error("\nSaveToList:", sendError.Error())
			Log.Debug(Log.Level.Error, sendError.Error())
		}
		return
	}

	// get the message
	message := guildData.SavedList[messageIndex]

	// remove it from guild data
	guildData.SavedListRemoveItem(message)

	// save it to firebase
	err = firebase.SetSavedList(i.GuildID, guildData.SavedList)
	if err != nil {
		// re-add the message on error
		guildData.SavedListAddItem(message)

		_, sendError := s.ChannelMessageSend(i.ChannelID, "❗ Something went wrong while removing the message.")
		if sendError != nil {
			Log.Error("\nSaveToList:", sendError.Error())
			Log.Debug(Log.Level.Error, sendError.Error())
		}
		Log.Error("\nSaveToList:", err.Error())
		Log.Debug(Log.Level.Error, sendError.Error())
		return
	}
	
	_, sendError = s.ChannelMessageSend(i.ChannelID, "✅ Message removed.")
	if sendError != nil {
		Log.Error("\nSaveToList:", sendError.Error())
		Log.Debug(Log.Level.Error, sendError.Error())
	}
}

// * Subcommands

func saveSubCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// check if the message is a reply
	if m.Message.Type != discordgo.MessageTypeReply {
		_, sendError := s.ChannelMessageSendReply(m.ChannelID, "Please reference a message to save by replying to it.", m.Reference())
		if sendError != nil {
			Log.Error("\nSaveToList:", sendError.Error())
			Log.Debug(Log.Level.Error, sendError.Error())
		}
		return
	}

	// get the replied message
	repliedMessage, err := s.ChannelMessage(m.ChannelID, m.MessageReference.MessageID)
	if err != nil {
		_, sendError := s.ChannelMessageSendReply(m.ChannelID, "❗ Something went wrong while getting the message to save.", m.Reference())
		if sendError != nil {
			Log.Error("\nSaveToList:", sendError.Error())
			Log.Debug(Log.Level.Error, sendError.Error())
		}
		Log.Error("\nSaveToList:", err.Error())
		Log.Debug(Log.Level.Error, err.Error())
		return
	}

	// save the message
	guildData, err := firebase.GetGuildData(m.GuildID)
	if err != nil { // shouldn't happen at this point
		Log.Error("\nSaveToList:", err.Error())
		Log.Debug(Log.Level.Error, err.Error())
		return
	}

	// already saved
	isExist := guildData.SavedListIsInList(repliedMessage.Content)
	if isExist {
		_, sendError := s.ChannelMessageSendReply(m.ChannelID, "❗ Message already saved.", m.Reference())
		if sendError != nil {
			Log.Error("\nSaveToList:", sendError.Error())
			Log.Debug(Log.Level.Error, sendError.Error())
		}
		return
	}

	// add to guild data
	guildData.SavedListAddItem(repliedMessage.Content)

	// upload new saved list data to firebase
	err = firebase.SetSavedList(m.GuildID, guildData.SavedList)
	if err != nil {
		// remove the message from guild data on error
		guildData.SavedListRemoveItem(repliedMessage.Content)

		_, sendError := s.ChannelMessageSendReply(m.ChannelID, "❗ Something went wrong while saving the message.", m.Reference())
		if sendError != nil {
			Log.Error("\nSaveToList:", sendError.Error())
			Log.Debug(Log.Level.Error, sendError.Error())
		}
		Log.Error("\nSaveToList:", err.Error())
		Log.Debug(Log.Level.Error, err.Error())
		return
	}

	_, sendError := s.ChannelMessageSendReply(m.ChannelID, "✅ Message saved.", m.Reference())
	if sendError != nil {
		Log.Error("\nSaveToList:", sendError.Error())
		Log.Debug(Log.Level.Error, sendError.Error())
	}
}

func getSubCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	guildData, err := firebase.GetGuildData(m.GuildID)
	if err != nil { // shouldn't happen at this point
		Log.Error("\nSaveToList:", err.Error())
		Log.Debug(Log.Level.Error, err.Error())
		return
	}

	if len(guildData.SavedList) == 0 {
		_, sendError := s.ChannelMessageSendReply(m.ChannelID, "❗ No saved messages found.", m.Reference())
		if sendError != nil {
			Log.Error("\nSaveToList:", sendError.Error())
			Log.Debug(Log.Level.Error, sendError.Error())
		}
		return
	}

	// print the saved messages
	_, sendError := s.ChannelMessageSendReply(m.ChannelID,
		fmt.Sprint("Saved Messages:\n🔹  ", strings.Join(guildData.SavedList, "\n🔹  ")),
		m.Reference(),
	)
	if sendError != nil {
		Log.Error("\nSaveToList:", sendError.Error())
		Log.Debug(Log.Level.Error, sendError.Error())
	}
}

func removeSubCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	guildData, err := firebase.GetGuildData(m.GuildID)
	if err != nil { // shouldn't happen at this point
		Log.Error("\nSaveToList:", err.Error())
		Log.Debug(Log.Level.Error, err.Error())
		return
	}

	if len(guildData.SavedList) == 0 {
		_, sendError := s.ChannelMessageSendReply(m.ChannelID, "❗ No saved messages found.", m.Reference())
		if sendError != nil {
			Log.Error("\nSaveToList:", sendError.Error())
			Log.Debug(Log.Level.Error, sendError.Error())
		}
		return
	}

	menuOptions := []*components.SelectMenuOption{}
	for i, message := range guildData.SavedList {
		menuOptions = append(menuOptions, components.NewMenuOption().SetLabel(message).SetValue(fmt.Sprintf("%d", i)))

		// ! max 25 options
		if i >= 24 {
			break
		}
	}

	_, sendError := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content: "Select a message to remove",
		Components: *components.AddMessageComponents(
			components.NewRow(
				components.NewSelectMenu().SetCustomID("saveToList_remove").SetPlaceholder("Select a message to remove").SetOptions(menuOptions...),
			),
		),
	})
	if sendError != nil {
		Log.Error("\nSaveToList:", sendError.Error())
		Log.Debug(Log.Level.Error, sendError.Error())
	}
}
