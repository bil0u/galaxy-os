//go:generate go run $WORKDIR/$GENSCRIPT --generator --config-dir=$WORKDIR
package enums

import (
	"github.com/disgoorg/snowflake/v2"
)

type RoleEnum string

const (
// Les constantes seront générées par go generate
)

var RoleMap = map[RoleEnum]snowflake.ID{
	// Les valeurs seront générées par go generate
}
