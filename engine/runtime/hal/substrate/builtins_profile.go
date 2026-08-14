package substrate

import (
	"aiki/engine"
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func profileCountsValue(c engine.SemanticCounts) value.Value {
	pairs := []struct {
		name string
		n    int64
	}{
		{"arithmetic", c.Arithmetic},
		{"comparison", c.Comparison},
		{"call", c.Call},
		{"iteration", c.Iteration},
		{"index", c.Index},
		{"send", c.Send},
		{"receive", c.Receive},
		{"store_read", c.StoreRead},
		{"store_write", c.StoreWrite},
	}
	out := make([]value.Value, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, &value.List{Elements: []value.Value{
			&value.Symbol{Val: p.name},
			value.NewNumber(p.n, 1),
		}})
	}
	return &value.List{Elements: out}
}

func profileMeasurementValue(m engine.SemanticMeasurement) value.Value {
	sites := make([]value.Value, 0, len(m.Sites))
	for _, sc := range m.Sites {
		sites = append(sites, &value.List{Elements: []value.Value{
			&value.Symbol{Val: string(sc.Kind)},
			value.NewNumber(sc.Count, 1),
			&value.String{Val: sc.Site.File},
			value.NewNumber(int64(sc.Site.Line), 1),
			value.NewNumber(int64(sc.Site.Col), 1),
			&value.String{Val: sc.Site.Function},
			&value.String{Val: sc.Site.Detail},
			&value.String{Val: sc.Site.Source},
		}})
	}
	return &value.List{Elements: []value.Value{profileCountsValue(m.Counts), &value.List{Elements: sites}}}
}

func halProfileMeasure(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("profile.measure: want 1 argument, got %d", len(args))
	}
	if ctx == nil || ctx.Measure == nil {
		return value.NewFault("profile.measure: measurement context not available")
	}
	result, measurement := ctx.Measure(args[0], nil, true)
	if fault, ok := result.(*value.Fault); ok {
		return fault
	}
	return profileMeasurementValue(measurement)
}

func halProfileCounts(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("profile.counts: want 1 argument, got %d", len(args))
	}
	if ctx == nil || ctx.Measure == nil {
		return value.NewFault("profile.counts: measurement context not available")
	}
	result, measurement := ctx.Measure(args[0], nil, false)
	if fault, ok := result.(*value.Fault); ok {
		return fault
	}
	return profileCountsValue(measurement.Counts)
}

func profileOne(size value.Value, fn value.Value, ctx *hal.EvalContext) value.Value {
	result, measurement := ctx.Measure(fn, []value.Value{size}, false)
	if fault, ok := result.(*value.Fault); ok {
		return fault
	}
	return &value.List{Elements: []value.Value{size, profileCountsValue(measurement.Counts)}}
}

func halProfileExperiment(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("profile.experiment: want 2 arguments, got %d", len(args))
	}
	if ctx == nil || ctx.Measure == nil {
		return value.NewFault("profile.experiment: measurement context not available")
	}

	if sizes, ok := args[0].(*value.List); ok {
		results := make([]value.Value, 0, len(sizes.Elements))
		for _, size := range sizes.Elements {
			measured := profileOne(size, args[1], ctx)
			if fault, ok := measured.(*value.Fault); ok {
				return fault
			}
			results = append(results, measured)
		}
		return &value.List{Elements: results}
	}

	return profileOne(args[0], args[1], ctx)
}
