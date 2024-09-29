package handlers

import (
	"log/slog"

	"github.com/disgoorg/disgo/events"
)

// OnRoleAssigned is triggered when a role is assigned to a user

func OnRoleAssigned(event *events.GuildMemberUpdate) {
	// Check if the role is the one we are interested in
	for _, roleID := range event.Member.RoleIDs {
		slog.Info("Role ID: ", slog.Any("roleID", roleID))
		// if roleID == "YourRoleName" {
		// 	log.Printf("Role %s assigned to user %s", role.Name, event.Member.User.ID)
		// 	// Add your custom logic here
		// 	break
		// }
	}
}
