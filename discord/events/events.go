package events

import (
	"discord-bot/common"
	"discord-bot/utils"

	"github.com/bwmarrin/discordgo"
)

type (
	OnComponentReactionEvent = func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.MessageComponentInteractionData)
	OnModalReactionEvent     = func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ModalSubmitInteractionData)
	OnReadyEvent             = func(s *discordgo.Session, e *discordgo.Ready)
	OnDmMessageEvent         = func(s *discordgo.Session, m *discordgo.MessageCreate)
	OnSlashCommandEvent      = func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData)
	OnVoiceStateUpdateEvent  = func(s *discordgo.Session, v *discordgo.VoiceStateUpdate, state common.VoiceState)
)

var (
	OnComponentReactionEvents []OnComponentReactionEvent
	OnModalReactionEvents     []OnModalReactionEvent
	OnReadyEvents             []OnReadyEvent
	OnDmMessageEvents         []OnDmMessageEvent
	OnSlashCommandEvents      []OnSlashCommandEvent
	OnVoiceStateUpdateEvents  []OnVoiceStateUpdateEvent
)

// * Register events
func RegisterComponentReactionEvent(event OnComponentReactionEvent) {
	OnComponentReactionEvents = append(OnComponentReactionEvents, event)
}
func RegisterModalReactionEvent(event OnModalReactionEvent) {
	OnModalReactionEvents = append(OnModalReactionEvents, event)
}
func RegisterOnReadyEvent(event OnReadyEvent) {
	OnReadyEvents = append(OnReadyEvents, event)
}
func RegisterOnDmMessageEvent(event OnDmMessageEvent) {
	OnDmMessageEvents = append(OnDmMessageEvents, event)
}
func RegisterOnSlashCommandEvent(event OnSlashCommandEvent) {
	OnSlashCommandEvents = append(OnSlashCommandEvents, event)
}
func RegisterOnVoiceStateUpdateEvent(event OnVoiceStateUpdateEvent) {
	OnVoiceStateUpdateEvents = append(OnVoiceStateUpdateEvents, event)
}

// * Handlers

// OnReady is called when the bot is up and running
func OnReady(s *discordgo.Session, e *discordgo.Ready) {
	utils.Log.Success("\nBot is up and running!")

	for _, event := range OnReadyEvents {
		event(s, e)
	}
}

// OnDM is called when the bot receives a DM
func OnDM(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if m.GuildID == "" {
		return
	}

	for _, event := range OnDmMessageEvents {
		event(s, m)
	}
}

// OnInteraction is called when the bot receives an interaction
func OnInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		return
	}

	// Commands
	if i.Type == discordgo.InteractionApplicationCommand {
		data := i.ApplicationCommandData()
		for _, event := range OnSlashCommandEvents {
			event(s, i, &data)
		}
		return
	}

	// Component Interactions
	if i.Type == discordgo.InteractionMessageComponent {
		data := i.MessageComponentData()
		for _, event := range OnComponentReactionEvents {
			event(s, i, &data)
		}
		return
	}

	// Modal Submissions
	if i.Type == discordgo.InteractionModalSubmit {
		data := i.ModalSubmitData()
		for _, event := range OnModalReactionEvents {
			event(s, i, &data)
		}
	}
}

// Event handler for voice state updates
func VoiceStateUpdate(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	user, err := s.User(vs.UserID)
	if err != nil {
		return
	}

	if user.Bot {
		return
	}

	status := common.VoiceStateNone
	if vs.BeforeUpdate == nil {
		status = common.VoiceJoined
	}
	if vs.BeforeUpdate != nil && vs.ChannelID == "" {
		status = common.VoiceLeft
	}
	if vs.BeforeUpdate != nil && vs.ChannelID != "" {
		status = common.VoiceSwitched
	}

	if status != common.VoiceStateNone {
		for _, event := range OnVoiceStateUpdateEvents {
			event(s, vs, status)
		}
	}
}
