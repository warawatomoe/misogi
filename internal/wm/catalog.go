package wm

import (
	"math"
)

type catOffEnt struct {
	family string
	tier   string
	w, h   int
}

var catOffSz = []catOffEnt{
	{"gemini-3.x-image", "0.5k", 512, 512},
	{"gemini-3.x-image", "0.5k", 256, 1024},
	{"gemini-3.x-image", "0.5k", 192, 1536},
	{"gemini-3.x-image", "0.5k", 424, 632},
	{"gemini-3.x-image", "0.5k", 632, 424},
	{"gemini-3.x-image", "0.5k", 448, 600},
	{"gemini-3.x-image", "0.5k", 1024, 256},
	{"gemini-3.x-image", "0.5k", 600, 448},
	{"gemini-3.x-image", "0.5k", 464, 576},
	{"gemini-3.x-image", "0.5k", 576, 464},
	{"gemini-3.x-image", "0.5k", 1536, 192},
	{"gemini-3.x-image", "0.5k", 384, 688},
	{"gemini-3.x-image", "0.5k", 688, 384},
	{"gemini-3.x-image", "0.5k", 792, 168},
	{"gemini-3.x-image", "1k", 1024, 1024},
	{"gemini-3.x-image", "1k", 512, 2048},
	{"gemini-3.x-image", "1k", 384, 3072},
	{"gemini-3.x-image", "1k", 848, 1264},
	{"gemini-3.x-image", "1k", 1264, 848},
	{"gemini-3.x-image", "1k", 896, 1200},
	{"gemini-3.x-image", "1k", 2048, 512},
	{"gemini-3.x-image", "1k", 1200, 896},
	{"gemini-3.x-image", "1k", 928, 1152},
	{"gemini-3.x-image", "1k", 1152, 928},
	{"gemini-3.x-image", "1k", 3072, 384},
	{"gemini-3.x-image", "1k", 768, 1376},
	{"gemini-3.x-image", "1k", 1376, 768},
	{"gemini-3.x-image", "1k", 1408, 768},
	{"gemini-3.x-image", "1k", 1584, 672},
	{"gemini-3.x-image", "2k", 2048, 2048},
	{"gemini-3.x-image", "2k", 1024, 4096},
	{"gemini-3.x-image", "2k", 768, 6144},
	{"gemini-3.x-image", "2k", 1696, 2528},
	{"gemini-3.x-image", "2k", 2528, 1696},
	{"gemini-3.x-image", "2k", 1792, 2400},
	{"gemini-3.x-image", "2k", 4096, 1024},
	{"gemini-3.x-image", "2k", 2400, 1792},
	{"gemini-3.x-image", "2k", 1856, 2304},
	{"gemini-3.x-image", "2k", 2304, 1856},
	{"gemini-3.x-image", "2k", 6144, 768},
	{"gemini-3.x-image", "2k", 1536, 2752},
	{"gemini-3.x-image", "2k", 2752, 1536},
	{"gemini-3.x-image", "2k", 3168, 1344},
	{"gemini-3.x-image", "2k-new-margin", 2816, 1536},
	{"gemini-3.x-image", "4k", 4096, 4096},
	{"gemini-3.x-image", "4k", 2048, 8192},
	{"gemini-3.x-image", "4k", 1536, 12288},
	{"gemini-3.x-image", "4k", 3392, 5056},
	{"gemini-3.x-image", "4k", 5056, 3392},
	{"gemini-3.x-image", "4k", 3584, 4800},
	{"gemini-3.x-image", "4k", 8192, 2048},
	{"gemini-3.x-image", "4k", 4800, 3584},
	{"gemini-3.x-image", "4k", 3712, 4608},
	{"gemini-3.x-image", "4k", 4608, 3712},
	{"gemini-3.x-image", "4k", 12288, 1536},
	{"gemini-3.x-image", "4k", 3072, 5504},
	{"gemini-3.x-image", "4k", 5504, 3072},
	{"gemini-3.x-image", "4k", 6336, 2688},
	{"gemini-2.5-flash-image", "1k", 1024, 1024},
	{"gemini-2.5-flash-image", "1k", 832, 1248},
	{"gemini-2.5-flash-image", "1k", 1248, 832},
	{"gemini-2.5-flash-image", "1k", 864, 1184},
	{"gemini-2.5-flash-image", "1k", 1184, 864},
	{"gemini-2.5-flash-image", "1k", 896, 1152},
	{"gemini-2.5-flash-image", "1k", 1152, 896},
	{"gemini-2.5-flash-image", "1k", 768, 1344},
	{"gemini-2.5-flash-image", "1k", 1344, 768},
	{"gemini-2.5-flash-image", "1k", 1536, 672},
}

func catMatchOff(w, h int) *catOffEnt {
	for i := range catOffSz {
		e := &catOffSz[i]
		if e.w == w && e.h == h {
			return e
		}
	}
	return nil
}

func catEntCfg(e *catOffEnt) Config {
	if e != nil && e.family == "gemini-3.x-image" && e.tier == "1k" {
		return cfg3x1k
	}
	if e != nil {
		if c, ok := cfgTier[e.tier]; ok {
			return c
		}
	}
	return Config{}
}

func catEntLeg(e *catOffEnt) []Config {
	if e != nil && e.family == "gemini-3.x-image" && e.tier == "1k" {
		return []Config{cfg3x1kLeg}
	}
	return nil
}

func catNewMarg(base Config, w, h int) (Config, bool) {
	if base.Size != 96 || (base.MarginRight == 192 && base.MarginBottom == 192) {
		return Config{}, false
	}
	c := Config{Size: 96, MarginRight: 192, MarginBottom: 192, AlphaVariant: "20260520"}
	if !fits(w, h, c.Size, c.MarginRight, c.MarginBottom) {
		return Config{}, false
	}
	return c, true
}

func catLgMarg(base Config, w, h int, anyBase bool) (Config, bool) {
	if !anyBase && (base.Size != 48 || (base.MarginRight == 96 && base.MarginBottom == 96)) {
		return Config{}, false
	}
	if base.MarginRight == 96 && base.MarginBottom == 96 {
		return Config{}, false
	}
	c := cfg3x1kLgMarg
	if !fits(w, h, c.Size, c.MarginRight, c.MarginBottom) {
		return Config{}, false
	}
	return c, true
}

func catV2Sm(w, h int) (Config, bool) {
	if math.Max(float64(w), float64(h)) > 2048 {
		return Config{}, false
	}
	long := math.Max(float64(w), float64(h))
	short := math.Min(float64(w), float64(h))
	srcLong := 2848.0
	if short >= 566 {
		srcLong = 2752
	} else if short >= 550 {
		srcLong = 2816
	}
	margin := int(math.Round(192 * (long / srcLong)))
	c := Config{Size: 36, MarginRight: margin, MarginBottom: margin, AlphaVariant: "v2"}
	if !fits(w, h, c.Size, c.MarginRight, c.MarginBottom) {
		return Config{}, false
	}
	return c, true
}

type catPrjSd struct {
	cfg   Config
	round func(float64) float64
}

func catPrjSds(e *catOffEnt, base Config) []catPrjSd {
	seeds := []catPrjSd{{cfg: base, round: math.Round}}
	if e != nil && e.family == "gemini-3.x-image" && e.tier == "1k" {
		seeds = append(seeds, catPrjSd{cfg: cfg3x1kLgMarg, round: math.Ceil})
	}
	return seeds
}

func catPrjCfg(base Config, sx, sy float64, minLogo, maxLogo int, round func(float64) float64) Config {
	if round == nil {
		round = math.Round
	}
	avg := (sx + sy) / 2
	size := int(round(float64(base.Size) * avg))
	if size < minLogo {
		size = minLogo
	}
	if size > maxLogo {
		size = maxLogo
	}
	return Config{
		Size:         size,
		MarginRight:  int(math.Max(8, math.Round(float64(base.MarginRight)*sx))),
		MarginBottom: int(math.Max(8, math.Round(float64(base.MarginBottom)*sy))),
		AlphaVariant: base.AlphaVariant,
	}
}

func catOffEnts(w, h int) []Config {
	if w <= 0 || h <= 0 {
		return nil
	}

	if m := catMatchOff(w, h); m != nil {
		base := catEntCfg(m)
		var out []Config
		out = append(out, base)

		if m.family == "gemini-3.x-image" && m.tier == "1k" {
			if c, ok := catLgMarg(base, w, h, false); ok {
				out = append(out, c)
			}
			if c, ok := catV2Sm(w, h); ok {
				out = append(out, c)
			}
		}
		out = append(out, catEntLeg(m)...)
		if !(m.family == "gemini-3.x-image" && m.tier == "1k") {
			if c, ok := catNewMarg(base, w, h); ok {
				out = append(out, c)
			}
		}
		return out
	}

	const (
		maxARDelta   = 0.02
		maxScaleMiss = 0.12
		minLogo      = 24
		maxLogo      = 192
		limit        = 3
	)

	targetAR := float64(w) / float64(h)
	type scored struct {
		cfg   Config
		score float64
	}
	var cands []scored

	for i := range catOffSz {
		e := &catOffSz[i]
		base := catEntCfg(e)
		if base.Size == 0 {
			continue
		}

		sx := float64(w) / float64(e.w)
		sy := float64(h) / float64(e.h)
		scale := (sx + sy) / 2
		entryAR := float64(e.w) / float64(e.h)
		arDelta := math.Abs(targetAR-entryAR) / entryAR
		scaleMiss := math.Abs(sx-sy) / math.Max(sx, sy)

		if arDelta > maxARDelta || scaleMiss > maxScaleMiss {
			continue
		}

		for _, seed := range catPrjSds(e, base) {
			c := catPrjCfg(seed.cfg, sx, sy, minLogo, maxLogo, seed.round)
			if !fits(w, h, c.Size, c.MarginRight, c.MarginBottom) {
				continue
			}
			cands = append(cands, scored{
				cfg:   c,
				score: arDelta*100 + scaleMiss*20 + math.Abs(math.Log2(math.Max(scale, 1e-6))),
			})
		}
	}

	for i := 0; i < len(cands); i++ {
		for j := i + 1; j < len(cands); j++ {
			if cands[j].score < cands[i].score {
				cands[i], cands[j] = cands[j], cands[i]
			}
		}
	}

	seen := map[string]bool{}
	var out []Config
	for _, s := range cands {
		k := cfgKey(s.cfg)
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, s.cfg)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func catDef(w, h int) Config {
	if m := catMatchOff(w, h); m != nil {
		return catEntCfg(m)
	}
	if w > 1024 && h > 1024 {
		return cfgStd(96)
	}
	return cfgStd(48)
}

func catEnts(w, h int) []Config {
	if w <= 0 || h <= 0 {
		return nil
	}

	def := catDef(w, h)
	official := catMatchOff(w, h) != nil

	var cs []Config
	cs = append(cs, def)
	cs = append(cs, catOffEnts(w, h)...)

	if c, ok := catLgMarg(def, w, h, false); ok {
		cs = append(cs, c)
	}
	if !official {
		if c, ok := catNewMarg(def, w, h); ok {
			cs = append(cs, c)
		}
		if c, ok := catLgMarg(def, w, h, true); ok {
			cs = append(cs, c)
		}
	}

	seen := map[string]bool{}
	var out []Config
	for _, c := range cs {
		k := cfgKey(c)
		if !seen[k] {
			seen[k] = true
			out = append(out, c)
		}
	}
	return out
}