package generators

import (
	"html/template"
	"log"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/bil0u/galaxy-os/sdk/utils"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var ChannelEnumGenerator = &utils.SourceFileGenerator{
	OutputFile: "sdk/enums/channel.go",
	Header:     utils.GeneratedFileHeader,
	TemplateFuncs: template.FuncMap{
		"FormatChannelName": formatChannelName,
	},
	GetData:  getDiscordChannels,
	Template: channelEnumTemplate,
}

type channelEnumGeneratorData struct {
	CategoryChannels []discord.GuildChannel
	Channels         []discord.GuildChannel
}

func formatChannelName(name string) string {
	// Remove diacritics
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	name, _, _ = transform.String(t, name)
	// Remove special characters
	return regexp.MustCompile(`[\P{L}+]`).ReplaceAllString(strings.Title(name), "")
}

func fetchGuildChannels(client bot.Client, Guilds []snowflake.ID) ([]discord.GuildChannel, error) {
	guilds := client.Rest()
	var channels []discord.GuildChannel
	for _, guildID := range Guilds {
		guildChannels, err := guilds.GetGuildChannels(guildID)
		if err != nil {
			return nil, err
		}
		channels = append(channels, guildChannels...)
	}
	return channels, nil
}

func getParentChannel(channel discord.GuildChannel, candidates []discord.GuildChannel) *discord.GuildChannel {
	if channel.ParentID() == nil {
		return nil
	}
	return utils.Find(candidates, func(c discord.GuildChannel) bool {
		return c.ID().String() == channel.ParentID().String()
	})
}

func getDiscordChannels(g *utils.SourceFileGenerator) any {
	// Fetch roles from Discord
	channels, err := fetchGuildChannels(g.Client, g.Cfg.Guilds)
	if err != nil {
		log.Fatalf("Failed to fetch channels: %v", err)
	}

	categoryChannels := utils.Filter(channels, func(channel discord.GuildChannel) bool {
		return channel.Type() == discord.ChannelTypeGuildCategory
	})

	channels = utils.Filter(channels, func(channel discord.GuildChannel) bool {
		return channel.Type() == discord.ChannelTypeGuildText ||
			channel.Type() == discord.ChannelTypeGuildVoice ||
			channel.Type() == discord.ChannelTypeGuildNews ||
			channel.Type() == discord.ChannelTypeGuildForum
	})

	// Sorting Category Channels by position, lower to higher
	slices.SortStableFunc(categoryChannels, func(c1, c2 discord.GuildChannel) int {
		return c1.Position() - c2.Position()
	})

	var typePriority = []discord.ChannelType{
		discord.ChannelTypeGuildText,
		discord.ChannelTypeGuildNews,
		discord.ChannelTypeGuildForum,
		discord.ChannelTypeGuildVoice,
	}

	// Sorting Channels according this rules, higher to lower:

	// - Channels with no parents first
	// - Channels with parents, sorted by parent position, then type, then position
	slices.SortStableFunc(channels, func(c1, c2 discord.GuildChannel) int {

		// Channels with parents, sorted by parent position, then type, then position
		parent1 := getParentChannel(c1, categoryChannels)
		parent2 := getParentChannel(c2, categoryChannels)

		// Channels with no parents first
		if parent1 == nil && parent2 == nil {
			return c1.Position() - c2.Position()
		} else if parent1 == nil && parent2 != nil {
			return -1
		} else if parent1 != nil && parent2 == nil {
			return 1
		}

		if (*parent1).Position() < (*parent2).Position() {
			return -1
		} else if (*parent1).Position() > (*parent2).Position() {
			return 1
		}

		if c1.Type() == c2.Type() {
			return c1.Position() - c2.Position()
		}

		return utils.IndexOf(typePriority, c1.Type()) - utils.IndexOf(typePriority, c2.Type())

	})
	return channelEnumGeneratorData{
		CategoryChannels: categoryChannels,
		Channels:         channels,
	}
}

const channelEnumTemplate = `
package enums

import (
	"github.com/disgoorg/snowflake/v2"
)

type GuildCategoryChannelEnum string
type GuildChannelEnum string

const (
{{- range .CategoryChannels }}
	GuildCategoryChannel{{ FormatChannelName .Name }} GuildCategoryChannelEnum = "{{ .Name }}"
{{- end }}
)

const (
{{- range .Channels }}
    GuildChannel{{ FormatChannelName .Name }} GuildChannelEnum = "{{ .Name }}"
{{- end }}
)

var GuildCategoryChannelMap = map[GuildCategoryChannelEnum]snowflake.ID{
{{- range .CategoryChannels }}
	GuildCategoryChannel{{ FormatChannelName .Name }}: {{ .ID }},
{{- end }}
}

var GuildChannelMap = map[GuildChannelEnum]snowflake.ID{
{{- range .Channels }}
    GuildChannel{{ FormatChannelName .Name }}: {{ .ID }},
{{- end }}
}

var GuildChannelCategoryMap = map[GuildChannelEnum]snowflake.ID{
{{- range .Channels }}
	GuildChannel{{ FormatChannelName .Name }}: {{ if .ParentID }}{{ .ParentID }}{{ else }}0{{ end }},
{{- end }}
}

// Guild Category Channels functions

func (e GuildCategoryChannelEnum) String() string {
	return string(e)
}

func (e GuildCategoryChannelEnum) ID() snowflake.ID {
	return GuildCategoryChannelMap[e]
}

func (e GuildCategoryChannelEnum) IsValid() bool {
	_, ok := GuildCategoryChannelMap[e]
	return ok
}

func (e GuildCategoryChannelEnum) GetChannels() []GuildChannelEnum {
	var channels []GuildChannelEnum
	for channel, parentID := range GuildChannelCategoryMap {
		if parentID == e.ID() {
			channels = append(channels, channel)
		}
	}
	return channels
}

// Guild Channels functions

func (e GuildChannelEnum) String() string {
	return string(e)
}

func (e GuildChannelEnum) ID() snowflake.ID {
	return GuildChannelMap[e]
}

func (e GuildChannelEnum) IsValid() bool {
	_, ok := GuildChannelMap[e]
	return ok
}

func (e GuildChannelEnum) ParentID() snowflake.ID {
	return GuildChannelCategoryMap[e]
}

`
