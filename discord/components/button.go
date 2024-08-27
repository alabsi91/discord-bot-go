package components

import "github.com/bwmarrin/discordgo"

type Button struct {
	discordgo.Button
}

func NewButton() *Button {
	return &Button{}
}

func (b *Button) SetLabel(label string) *Button {
	b.Label = label
	return b
}

func (b *Button) SetDisabled(disabled bool) *Button {
	b.Disabled = disabled
	return b
}

func (b *Button) SetEmoji(emoji *discordgo.ComponentEmoji) *Button {
	b.Emoji = emoji
	return b
}

func (b *Button) SetEmojiFromString(emoji string) *Button {
	b.Emoji = &discordgo.ComponentEmoji{Name: emoji}
	return b
}

func (b *Button) SetStyle(style discordgo.ButtonStyle) *Button {
	b.Style = style
	return b
}

func (b *Button) SetStylePrimary() *Button {
	b.Style = discordgo.PrimaryButton
	return b
}

func (b *Button) SetStyleSecondary() *Button {
	b.Style = discordgo.SecondaryButton
	return b
}

func (b *Button) SetStyleSuccess() *Button {
	b.Style = discordgo.SuccessButton
	return b
}

func (b *Button) SetStyleDanger() *Button {
	b.Style = discordgo.DangerButton
	return b
}

func (b *Button) SetStyleLink() *Button {
	b.Style = discordgo.LinkButton
	return b
}

func (b *Button) SetURL(url string) *Button {
	b.URL = url
	return b
}

func (b *Button) SetCustomID(customID string) *Button {
	b.CustomID = customID
	return b
}
