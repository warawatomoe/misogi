package wm

import (
	"image"
	"math"

	e "misogi/internal/err"
)

const minDetectScore = 0.0002

var gainCandidates = []float64{0.6, 0.85, 1.0}

type Config struct {
	Size, MarginRight, MarginBottom int
}

func standardCfg(size int) Config {
	if size == 96 {
		return Config{96, 64, 64}
	}
	return Config{48, 32, 32}
}

func detectCfg(width, height int, sizeOpt string) (Config, *e.E) {
	if width <= 0 || height <= 0 {
		return Config{}, e.New("misogi", e.Call, "wm:cfg", "bad dimensions")
	}
	switch sizeOpt {
	case "", "auto":
		if width > 1024 && height > 1024 {
			return standardCfg(96), nil
		}
		return standardCfg(48), nil
	case "48":
		return standardCfg(48), nil
	case "96":
		return standardCfg(96), nil
	default:
		return Config{}, e.New("misogi", e.Prov, "wm:cfg", "expected auto, 48, 96")
	}
}

func ResolveCfg(rgba *image.RGBA, sizeOpt string) (Config, float64, *e.E) {
	b := rgba.Bounds()
	width := b.Dx()
	height := b.Dy()

	cfg, err := detectCfg(width, height, sizeOpt)
	if err != nil {
		return Config{}, 0, err
	}
	if sizeOpt != "" && sizeOpt != "auto" {
		return cfg, scoreCfg(rgba, cfg), nil
	}

	candidates := []Config{cfg}
	alt := standardCfg(48)
	if cfg.Size == 48 {
		alt = standardCfg(96)
	}
	candidates = append(candidates, alt)
	candidates = append(candidates, autoExtraCandidates(width, height)...)

	seen := make(map[Config]bool)
	uniq := make([]Config, 0, len(candidates))
	for _, c := range candidates {
		if seen[c] {
			continue
		}
		seen[c] = true
		uniq = append(uniq, c)
	}

	best := cfg
	bestScore := scoreCfg(rgba, cfg)
	for _, c := range uniq[1:] {
		s := scoreCfg(rgba, c)
		if s > bestScore {
			best = c
			bestScore = s
		}
	}
	return best, bestScore, nil
}

func Detected(score float64) bool {
	return score >= minDetectScore
}

func PickGain(src *image.RGBA, cfg Config, alpha []float32, passes int, userGain float64) float64 {
	if userGain != 1.0 {
		return userGain
	}

	baseNearBlack := nearBlackRatio(src, cfg)
	maxNearBlack := math.Min(1, baseNearBlack+0.15)

	bestGain := 1.0
	bestResidual := math.Inf(1)
	for _, g := range gainCandidates {
		trial := cloneRGBA(src)
		if trial == nil {
			continue
		}
		if Remove(trial, cfg, alpha, g, passes) != nil {
			continue
		}
		if nearBlackRatio(trial, cfg) > maxNearBlack {
			continue
		}
		residual := scoreCfg(trial, cfg)
		if residual < bestResidual {
			bestResidual = residual
			bestGain = g
		}
	}
	return bestGain
}

func Alpha(size int) []float32 {
	switch size {
	case 48:
		return alpha48[:]
	case 96:
		return alpha96[:]
	default:
		return nil
	}
}

func Remove(rgba *image.RGBA, cfg Config, alpha []float32, gain float64, passes int) *e.E {
	if rgba == nil || len(alpha) == 0 || cfg.Size <= 0 || passes <= 0 || gain <= 0.0 {
		return e.New("misogi", e.Call, "wm:remove", "bad input")
	}
	if cfg.Size != 48 && cfg.Size != 96 {
		return e.New("misogi", e.Call, "wm:remove", "unsupported watermark size")
	}

	b := rgba.Bounds()
	width := b.Dx()
	height := b.Dy()
	x0 := width - cfg.MarginRight - cfg.Size
	y0 := height - cfg.MarginBottom - cfg.Size

	if x0 < 0 || y0 < 0 || x0+cfg.Size > width || y0+cfg.Size > height {
		return e.New("misogi", e.Prov, "wm:remove", "watermark outside image")
	}

	for pass := 0; pass < passes; pass++ {
		for row := 0; row < cfg.Size; row++ {
			for col := 0; col < cfg.Size; col++ {
				aidx := row*cfg.Size + col
				raw := float64(alpha[aidx])
				sig := math.Max(0, raw-(3.0/255.0)) * gain
				if sig < 0.002 {
					continue
				}
				a := raw * gain
				if a > 0.99 {
					a = 0.99
				}
				one := 1.0 - a
				x := b.Min.X + x0 + col
				y := b.Min.Y + y0 + row
				idx := rgba.PixOffset(x, y)
				for c := 0; c < 3; c++ {
					ch := float64(rgba.Pix[idx+c])
					orig := (ch - a*255.0) / one
					rgba.Pix[idx+c] = clamp255(orig)
				}
			}
		}
	}
	return nil
}

func nearBlackRatio(rgba *image.RGBA, cfg Config) float64 {
	if rgba == nil || cfg.Size <= 0 {
		return 0
	}
	b := rgba.Bounds()
	width := b.Dx()
	height := b.Dy()
	x0 := width - cfg.MarginRight - cfg.Size
	y0 := height - cfg.MarginBottom - cfg.Size
	if x0 < 0 || y0 < 0 || x0+cfg.Size > width || y0+cfg.Size > height {
		return 0
	}

	black := 0
	total := cfg.Size * cfg.Size
	for row := 0; row < cfg.Size; row++ {
		for col := 0; col < cfg.Size; col++ {
			x := b.Min.X + x0 + col
			y := b.Min.Y + y0 + row
			idx := rgba.PixOffset(x, y)
			if rgba.Pix[idx] == 0 && rgba.Pix[idx+1] == 0 && rgba.Pix[idx+2] == 0 {
				black++
			}
		}
	}
	return float64(black) / float64(total)
}

func autoExtraCandidates(width, height int) []Config {
	out := []Config{}
	if large := largeMargin48Cfg(width, height); large.Size > 0 {
		out = append(out, large)
	}
	return out
}

func largeMargin48Cfg(width, height int) Config {
	cfg := Config{48, 96, 96}
	if width-cfg.MarginRight-cfg.Size < 0 || height-cfg.MarginBottom-cfg.Size < 0 {
		return Config{}
	}
	return cfg
}

func scoreCfg(rgba *image.RGBA, cfg Config) float64 {
	if cfg.Size != 48 && cfg.Size != 96 {
		return math.Inf(-1)
	}

	b := rgba.Bounds()
	width := b.Dx()
	height := b.Dy()
	x0 := width - cfg.MarginRight - cfg.Size
	y0 := height - cfg.MarginBottom - cfg.Size
	if x0 < 0 || y0 < 0 || x0+cfg.Size > width || y0+cfg.Size > height {
		return math.Inf(-1)
	}

	alpha := Alpha(cfg.Size)
	if alpha == nil {
		return math.Inf(-1)
	}

	patch := regionGray(rgba, cfg)
	alpha64 := make([]float64, len(alpha))
	for i, v := range alpha {
		alpha64[i] = float64(v)
	}
	return normalizedCrossCorrelation(patch, alpha64)
}

func regionGray(rgba *image.RGBA, cfg Config) []float64 {
	b := rgba.Bounds()
	width := b.Dx()
	height := b.Dy()
	x0 := width - cfg.MarginRight - cfg.Size
	y0 := height - cfg.MarginBottom - cfg.Size

	patch := make([]float64, cfg.Size*cfg.Size)
	for row := 0; row < cfg.Size; row++ {
		for col := 0; col < cfg.Size; col++ {
			x := b.Min.X + x0 + col
			y := b.Min.Y + y0 + row
			idx := rgba.PixOffset(x, y)
			r := float64(rgba.Pix[idx])
			g := float64(rgba.Pix[idx+1])
			bl := float64(rgba.Pix[idx+2])
			patch[row*cfg.Size+col] = (0.2126*r + 0.7152*g + 0.0722*bl) / 255.0
		}
	}
	return patch
}

func normalizedCrossCorrelation(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	meanA, varA := meanVariance(a)
	meanB, varB := meanVariance(b)
	den := math.Sqrt(varA*varB) * float64(len(a))
	if den < 1e-12 {
		return 0
	}
	var num float64
	for i := range a {
		num += (a[i] - meanA) * (b[i] - meanB)
	}
	return num / den
}

func meanVariance(v []float64) (mean, variance float64) {
	if len(v) == 0 {
		return 0, 0
	}
	for _, x := range v {
		mean += x
	}
	mean /= float64(len(v))
	for _, x := range v {
		d := x - mean
		variance += d * d
	}
	return mean, variance
}

func clamp255(v float64) uint8 {
	if v <= 0.0 {
		return 0
	}
	if v >= 255.0 {
		return 255
	}
	return uint8(v + 0.5)
}
