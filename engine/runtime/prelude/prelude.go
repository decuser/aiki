// Package prelude embeds the Aiki prelude source and documentation.
package prelude

import (
	_ "embed"
)

//go:embed prelude.ai
var Source string

//go:embed prelude.help
var HelpSource string

//go:embed prelude.doc
var DocSource string
