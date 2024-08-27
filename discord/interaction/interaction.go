package interaction

import (
	"errors"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func RespondWithText(s *discordgo.Session, i *discordgo.InteractionCreate, text string, ephemeral bool) error {
	Flags := discordgo.MessageFlagsCrossPosted
	if ephemeral {
		Flags = discordgo.MessageFlagsEphemeral
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   Flags,
			Content: text,
		},
	})
}

func RespondWithThinking(s *discordgo.Session, i *discordgo.InteractionCreate, ephemeral bool) error {
	Flags := discordgo.MessageFlagsLoading
	if ephemeral {
		Flags += discordgo.MessageFlagsEphemeral
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: Flags},
	})
}

// Inform discord I got the interaction, I will edit the message later
func RespondWithNothing(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
}

// Edit the original message, Note: this will clear the components
func RespondEdit(s *discordgo.Session, i *discordgo.InteractionCreate, msg ...string) error {
	content := strings.Join(msg, " ")
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &content,
		Components: &[]discordgo.MessageComponent{}, // clear components
	})

	return err
}

func RespondEditWithComponents(s *discordgo.Session, i *discordgo.InteractionCreate, msg *string, components *[]discordgo.MessageComponent) error {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    msg,
		Components: components,
	})

	return err
}

func findUserVoiceState(session *discordgo.Session, userID string) (*discordgo.VoiceState, error) {
	for _, guild := range session.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == userID {
				return vs, nil
			}
		}
	}

	return nil, errors.New("could not find user's voice state")
}

// joinUserVoiceChannel joins a session to the same channel as another user.
func JoinUserVoiceChannel(session *discordgo.Session, userID string) (*discordgo.VoiceConnection, error) {
	// Find a user's current voice channel
	vs, err := findUserVoiceState(session, userID)
	if err != nil {
		return nil, err
	}

	// Join the user's channel and start unmuted and deafened.
	return session.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, true)
}

type interactionResponse struct {
	discordgo.InteractionResponse
}

func NewInteractionResponse() *interactionResponse {
	return &interactionResponse{}
}

func (i *interactionResponse) SetType(t discordgo.InteractionResponseType) *interactionResponse {
	i.Type = t
	return i
}

func (i *interactionResponse) SetData(data *discordgo.InteractionResponseData) *discordgo.InteractionResponse {
	i.Data = data

	return &discordgo.InteractionResponse{
		Type: i.Type,
		Data: i.Data,
	}
}

// * MARK: interaction Data

type interactionResponseData struct {
	discordgo.InteractionResponseData
}

func NewResponseData() *interactionResponseData {
	return &interactionResponseData{}
}

// NOTE: modal interaction only.
func (i *interactionResponseData) SetTitle(title string) *interactionResponseData {
	i.Title = title
	return i
}

// NOTE: modal interaction only.
func (i *interactionResponseData) SetCustomID(customID string) *interactionResponseData {
	i.CustomID = customID
	return i
}

// NOTE: only MessageFlagsSuppressEmbeds and MessageFlagsEphemeral can be set.
func (i *interactionResponseData) SetFlags(flags discordgo.MessageFlags) *interactionResponseData {
	i.Flags = flags
	return i
}

// NOTE: autocomplete interaction only. up to 25 choices
func (i *interactionResponseData) SetChoices(choices []*discordgo.ApplicationCommandOptionChoice) *interactionResponseData {
	i.Choices = choices
	return i
}

func (i *interactionResponseData) SetContent(content string) *interactionResponseData {
	i.Content = content
	return i
}

func (i *interactionResponseData) SetEmbeds(embeds ...*discordgo.MessageEmbed) *interactionResponseData {
	i.Embeds = embeds
	return i
}

func (i *interactionResponseData) SetComponents(components ...discordgo.MessageComponent) *interactionResponseData {
	i.Components = components
	return i
}

func (i *interactionResponseData) SetAllowedMentions(allowedMentions *discordgo.MessageAllowedMentions) *interactionResponseData {
	i.AllowedMentions = allowedMentions
	return i
}

func (i *interactionResponseData) SetFiles(files ...*discordgo.File) *interactionResponseData {
	i.Files = files
	return i
}

func (i *interactionResponseData) SetAttachments(attachments ...*discordgo.MessageAttachment) *interactionResponseData {
	i.Attachments = &attachments
	return i
}

func (i *interactionResponseData) SetTTS(tts bool) *interactionResponseData {
	i.TTS = tts
	return i
}

func (i *interactionResponseData) Into() *discordgo.InteractionResponseData {
	return &i.InteractionResponseData
}
