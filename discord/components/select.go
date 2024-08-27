package components

import "github.com/bwmarrin/discordgo"

type SelectMenu struct {
	discordgo.SelectMenu
}

// max is 25 options
func NewSelectMenu() *SelectMenu {
	return &SelectMenu{}
}

func (s *SelectMenu) SetStringType() *SelectMenu {
	s.MenuType = discordgo.StringSelectMenu
	return s
}

func (s *SelectMenu) SetUserType() *SelectMenu {
	s.MenuType = discordgo.UserSelectMenu
	return s
}

func (s *SelectMenu) SetRoleType() *SelectMenu {
	s.MenuType = discordgo.RoleSelectMenu
	return s
}

func (s *SelectMenu) SetMentionableType() *SelectMenu {
	s.MenuType = discordgo.MentionableSelectMenu
	return s
}

func (s *SelectMenu) SetChannelType() *SelectMenu {
	s.MenuType = discordgo.ChannelSelectMenu
	return s
}

func (s *SelectMenu) SetType(t discordgo.SelectMenuType) *SelectMenu {
	s.MenuType = t
	return s
}

func (s *SelectMenu) SetCustomID(id string) *SelectMenu {
	s.CustomID = id
	return s
}

func (s *SelectMenu) SetPlaceholder(placeholder string) *SelectMenu {
	s.Placeholder = placeholder
	return s
}

func (s *SelectMenu) SetMinValues(min *int) *SelectMenu {
	s.MinValues = min
	return s
}

func (s *SelectMenu) SetMaxValues(max *int) *SelectMenu {
	s.MaxValues = *max
	return s
}

func (s *SelectMenu) SetDisabled(disabled bool) *SelectMenu {
	s.Disabled = disabled
	return s
}

func (s *SelectMenu) SetChannelTypes(types ...discordgo.ChannelType) *SelectMenu {
	s.ChannelTypes = types
	return s
}

func (s *SelectMenu) SetOptions(options ...*SelectMenuOption) *SelectMenu {
	discordOptions := make([]discordgo.SelectMenuOption, len(options))

	for i, option := range options {
		discordOptions[i] = option.SelectMenuOption
	}

	s.Options = discordOptions
	return s
}

// * MARK: Select Menu Options

type SelectMenuOption struct {
	discordgo.SelectMenuOption
}

func NewMenuOption() *SelectMenuOption {
	return &SelectMenuOption{}
}

func (s *SelectMenuOption) SetLabel(label string) *SelectMenuOption {
	s.Label = label
	return s
}

func (s *SelectMenuOption) SetValue(value string) *SelectMenuOption {
	s.Value = value
	return s
}

func (s *SelectMenuOption) SetDescription(description string) *SelectMenuOption {
	s.Description = description
	return s
}

func (s *SelectMenuOption) SetEmoji(emoji *discordgo.ComponentEmoji) *SelectMenuOption {
	s.Emoji = emoji
	return s
}

func (s *SelectMenuOption) SetEmojiFromString(emoji string) *SelectMenuOption {
	s.Emoji = &discordgo.ComponentEmoji{Name: emoji}
	return s
}

func (s *SelectMenuOption) SetDefault(def bool) *SelectMenuOption {
	s.Default = def
	return s
}
