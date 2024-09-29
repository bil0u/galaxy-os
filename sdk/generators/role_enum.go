package generators

import (
	"bytes"
	"go/format"
	"log"
	"os"
	"text/template"

	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
)

const RolesEnumTemplate = `// Code generated by go generate; DO NOT EDIT.
// Instead, run Makefile's "generate" target to update this file.
package enums

import (
	"github.com/disgoorg/snowflake/v2"
)

type RoleEnum string

const (
{{- range .Roles }}
    Role{{ .Name }} RoleEnum = "{{ .Name }}"
{{- end }}
)

var RoleMap = map[RoleEnum]snowflake.ID{
{{- range .Roles }}
    Role{{ .Name }}: {{ .ID }},
{{- end }}
}
`

func GenerateRoleEnum(client bot.Client, cfg sdk.BotConfig) error {

	roles, err := sdk.FetchRoles(client, cfg.Guilds)
	if err != nil {
		log.Fatalf("Failed to fetch roles: %v", err)
	}

	data := struct {
		Roles []discord.Role
	}{
		Roles: roles,
	}

	tmpl, err := template.New("enum").Parse(RolesEnumTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatalf("Failed to format source: %v", err)
	}

	if err := os.WriteFile("role.go", formatted, 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	return nil
}
