package components

import "github.com/bwmarrin/discordgo"

type textInput struct {
	*discordgo.TextInput
}

// NOTE: modal interaction only.
func NewTextInput() *textInput {
	return &textInput{}
}

func (t *textInput) SetLabel(label string) *textInput {
	t.Label = label
	return t
}

func (t *textInput) SetPlaceholder(placeholder string) *textInput {
	t.Placeholder = placeholder
	return t
}

func (t *textInput) SetCustomID(customID string) *textInput {
	t.CustomID = customID
	return t
}

func (t *textInput) SetStyleShort() *textInput {
	t.Style = discordgo.TextInputShort
	return t
}

func (t *textInput) SetStyleParagraph() *textInput {
	t.Style = discordgo.TextInputParagraph
	return t
}

func (t *textInput) SetStyle(style discordgo.TextInputStyle) *textInput {
	t.Style = style
	return t
}

func (t *textInput) SetValue(value string) *textInput {
	t.Value = value
	return t
}

func (t *textInput) SetRequired(required bool) *textInput {
	t.Required = required
	return t
}

func (t *textInput) SetMinLength(minLength int) *textInput {
	t.MinLength = minLength
	return t
}

func (t *textInput) SetMaxLength(maxLength int) *textInput {
	t.MaxLength = maxLength
	return t
}
