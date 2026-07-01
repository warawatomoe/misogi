package wm

import (
	"fmt"

	e "misogi/internal/err"
)

type Config struct {
	Size, MarginRight, MarginBottom int
	AlphaVariant                    string
}

var cfgTier = map[string]Config{
	"0.5k":          {Size: 48, MarginRight: 32, MarginBottom: 32},
	"1k":            {Size: 96, MarginRight: 64, MarginBottom: 64},
	"2k":            {Size: 96, MarginRight: 64, MarginBottom: 64},
	"4k":            {Size: 96, MarginRight: 64, MarginBottom: 64},
	"2k-new-margin": {Size: 96, MarginRight: 192, MarginBottom: 192, AlphaVariant: "20260520"},
}

var (
	cfg3x1k        = Config{Size: 48, MarginRight: 32, MarginBottom: 32}
	cfg3x1kLeg     = Config{Size: 96, MarginRight: 64, MarginBottom: 64}
	cfg3x1kLgMarg  = Config{Size: 48, MarginRight: 96, MarginBottom: 96}
)

func cfgStd(size int) Config {
	if size == 96 {
		return Config{Size: 96, MarginRight: 64, MarginBottom: 64}
	}
	return Config{Size: 48, MarginRight: 32, MarginBottom: 32}
}

func cfgDet(w, h int, sizeOpt string) (Config, *e.E) {
	if w <= 0 || h <= 0 {
		return Config{}, e.New("misogi", e.Call, "wm:cfg", "bad dimensions")
	}
	switch sizeOpt {
	case "", "auto":
		cs := catEnts(w, h)
		if len(cs) > 0 {
			return cs[0], nil
		}
		return cfgStd(48), nil
	case "48":
		return cfgStd(48), nil
	case "96":
		return cfgStd(96), nil
	default:
		return Config{}, e.New("misogi", e.Prov, "wm:cfg", "expected auto, 48, 96")
	}
}

func cfgKey(c Config) string {
	return fmt.Sprintf("%d:%d:%d:%s", c.Size, c.MarginRight, c.MarginBottom, c.AlphaVariant)
}

func fits(w, h, size, mr, mb int) bool {
	return w-mr-size >= 0 && h-mb-size >= 0
}