package wm

import (
	"image"
	"math"

	e "misogi/internal/err"
)

func Remove(rgba *image.RGBA, cfg Config, alpha []float32, gain float64, passes int) *e.E {
	if rgba == nil || len(alpha) == 0 || cfg.Size <= 0 || passes <= 0 || gain <= 0.0 {
		return e.New("misogi", e.Call, "wm:remove", "bad input")
	}
	if cfg.Size > 192 {
		return e.New("misogi", e.Call, "wm:remove", "unsupported watermark size")
	}

	b := rgba.Bounds()
	w, h := b.Dx(), b.Dy()
	x0 := w - cfg.MarginRight - cfg.Size
	y0 := h - cfg.MarginBottom - cfg.Size
	if x0 < 0 || y0 < 0 || x0+cfg.Size > w || y0+cfg.Size > h {
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

func rmNb(rgba *image.RGBA, cfg Config) float64 {
	if rgba == nil || cfg.Size <= 0 {
		return 0
	}
	b := rgba.Bounds()
	w, h := b.Dx(), b.Dy()
	x0 := w - cfg.MarginRight - cfg.Size
	y0 := h - cfg.MarginBottom - cfg.Size
	if x0 < 0 || y0 < 0 || x0+cfg.Size > w || y0+cfg.Size > h {
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

func clamp255(v float64) uint8 {
	if v <= 0.0 {
		return 0
	}
	if v >= 255.0 {
		return 255
	}
	return uint8(v + 0.5)
}