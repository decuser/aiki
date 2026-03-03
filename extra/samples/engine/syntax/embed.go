package syntax

import _ "embed"

//go:embed grammar.ebnfx
var EbnfxSource string

//go:embed grammar.help
var HelpSource string
