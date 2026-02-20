package syntax

import(
	"aiki/engine"
	"regexp"
)

type GrammarContract interface {
	GetTokens() []TokenDef
	GetProduction(name string) (Production, bool)
	GetStart() string
	Observe() engine.Observer
	SetObserver(engine.Observer)
}

type TokenDef struct {
	Name    string
	Pattern *regexp.Regexp
	Skip    bool
}

type Production struct {
	Expressions [][]Term
}

type Term struct {
	Value    string
	IsSymbol bool
	IsOption bool
	IsRepeat bool
}
