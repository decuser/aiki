package substrate

import (
	"aiki/engine"
	"aiki/engine/runtime/hal"
)

func semanticSiteFromContext(ctx *hal.EvalContext) engine.SemanticSite {
	site := engine.SemanticSite{}
	if ctx == nil || ctx.Env == nil || ctx.Node == nil {
		return site
	}
	site.File = ctx.Env.GetFile()
	site.Line = ctx.Node.Pos.Line
	site.Col = ctx.Node.Pos.Col
	if frame, ok := ctx.Env.CurrentFrame(); ok {
		site.Function = frame.Name
	}
	if site.Line > 0 {
		site.Source = ctx.Env.GetSourceLine(site.Line)
	}
	return site
}

func semanticHit(ctx *hal.EvalContext, kind engine.SemanticKind) {
	semanticHitDetail(ctx, kind, "")
}

func semanticHitDetail(ctx *hal.EvalContext, kind engine.SemanticKind, detail string) {
	if ctx == nil || ctx.Probe == nil {
		return
	}
	site := engine.SemanticSite{}
	if attributed, ok := ctx.Probe.(engine.AttributionProbe); ok && attributed.WantsSites() {
		site = semanticSiteFromContext(ctx)
		site.Detail = detail
	}
	ctx.Probe.Hit(kind, site)
}
