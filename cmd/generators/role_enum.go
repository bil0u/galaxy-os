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

var RoleEnumGenerator = &utils.SourceFileGenerator{
	OutputFile: "sdk/enums/role.go",
	Header:     utils.GeneratedFileHeader,
	TemplateFuncs: template.FuncMap{
		"FormatRoleName": formatRoleName,
	},
	GetData:  getDiscordRoles,
	Template: roleEnumTemplate,
}

type roleEnumGeneratorData struct {
	Roles []discord.Role
}

func formatRoleName(name string) string {
	// Remove diacritics
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	name, _, _ = transform.String(t, name)
	// Remove special characters
	return regexp.MustCompile(`[\P{L}+]`).ReplaceAllString(strings.Title(name), "")
}

func fetchRoles(client bot.Client, Guilds []snowflake.ID) ([]discord.Role, error) {
	guilds := client.Rest()
	var roles []discord.Role
	for _, guildID := range Guilds {
		guildRoles, err := guilds.GetRoles(guildID)
		if err != nil {
			return nil, err
		}
		roles = append(roles, guildRoles...)
	}
	return roles, nil
}

func getDiscordRoles(g *utils.SourceFileGenerator) any {
	// Fetch roles from Discord
	roles, err := fetchRoles(g.Client, g.Cfg.Guilds)
	if err != nil {
		log.Fatalf("Failed to fetch roles: %v", err)
	}

	// Filtering roles to remove bot roles
	roles = utils.Filter(roles, func(role discord.Role) bool {
		return role.Tags == nil || role.Tags.BotID == nil
	})

	// Sorting them by position, higher to lower
	slices.SortStableFunc(roles, func(r1 discord.Role, r2 discord.Role) int {
		return r2.Position - r1.Position
	})

	return roleEnumGeneratorData{Roles: roles}
}

const roleEnumTemplate = `
package enums

import (
	"github.com/disgoorg/snowflake/v2"
)

type RoleEnum string

const (
{{- range .Roles }}
    Role{{ FormatRoleName .Name }} RoleEnum = "{{ .Name }}"
{{- end }}
)

var RoleMap = map[RoleEnum]snowflake.ID{
{{- range .Roles }}
    Role{{ FormatRoleName .Name }}: {{ .ID }},
{{- end }}
}

func (e RoleEnum) String() string {
	return string(e)
}

func (e RoleEnum) ID() snowflake.ID {
	return RoleMap[e]
}

func (e RoleEnum) IsValid() bool {
	_, ok := RoleMap[e]
	return ok
}
`
