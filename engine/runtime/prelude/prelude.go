// Package prelude embeds the Aiki prelude source.
package prelude

import (
	_ "embed"
)

//go:embed prelude.ai
var Source string
