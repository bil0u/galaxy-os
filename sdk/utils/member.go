package utils

import (
	"fmt"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

// FetchMembers retrieves all members from the bot's dev guilds and returns them as a slice of discord.Member
func FetchMembers(client bot.Client, guildID snowflake.ID) ([]discord.Member, error) {
	guilds := client.Rest()
	// Fetch the guild users
	guildMembers := []discord.Member{}
	chunkSize := 100
	for i := 0; i < len(guildMembers); i += chunkSize {
		end := i + chunkSize
		if end > len(guildMembers) {
			end = len(guildMembers)
		}
		members, err := guilds.GetMembers(guildID, chunkSize, guildMembers[end].User.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch guildMembers for guild %s: %v", guildID.String(), err)
		}
		guildMembers = append(guildMembers, members...)
	}
	return guildMembers, nil
}
