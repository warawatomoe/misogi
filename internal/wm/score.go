package wm

import (
	"image"
	"math"
)

type scrVal struct {
	spatial  float64
	gradient float64
}

func scrCfg(rgba *image.RGBA, cfg Config) float64 {
	if cfg.Size <= 0 || cfg.Size > 192 {
		return math.Inf(-1)
	}
	b := rgba.Bounds()
	w, h := b.Dx(), b.Dy()
	x0 := w - cfg.MarginRight - cfg.Size
	y0 := h - cfg.MarginBottom - cfg.Size
	if x0 < 0 || y0 < 0 || x0+cfg.Size > w || y0+cfg.Size > h {
		return math.Inf(-1)
	}
	alpha := AlphaForConfig(cfg)
	if alpha == nil {
		return math.Inf(-1)
	}
	patch := scrGray(rgba, cfg)
	a64 := make([]float64, len(alpha))
	for i, v := range alpha {
		a64[i] = float64(v)
	}
	return ncc(patch, a64)
}

func scrReg(rgba *image.RGBA, cfg Config, alpha []float32) scrVal {
	patch := scrGray(rgba, cfg)
	a64 := make([]float64, len(alpha))
	for i, v := range alpha {
		a64[i] = float64(v)
	}
	gP := sobel(patch, cfg.Size)
	gA := sobel(a64, cfg.Size)
	return scrVal{
		spatial:  ncc(patch, a64),
		gradient: ncc(gP, gA),
	}
}

func scrGray(rgba *image.RGBA, cfg Config) []float64 {
	b := rgba.Bounds()
	w, h := b.Dx(), b.Dy()
	x0 := w - cfg.MarginRight - cfg.Size
	y0 := h - cfg.MarginBottom - cfg.Size

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

func sobel(v []float64, size int) []float64 {
	if len(v) != size*size || size < 3 {
		out := make([]float64, len(v))
		copy(out, v)
		return out
	}
	out := make([]float64, len(v))
	for y := 1; y < size-1; y++ {
		for x := 1; x < size-1; x++ {
			i := y*size + x
			gx := -v[i-size-1] - 2*v[i-1] - v[i+size-1] +
				v[i-size+1] + 2*v[i+1] + v[i+size+1]
			gy := -v[i-size-1] - 2*v[i-size] - v[i-size+1] +
				v[i+size-1] + 2*v[i+size] + v[i+size+1]
			out[i] = math.Hypot(gx, gy)
		}
	}
	return out
}

func ncc(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	meanA, varA := meanVar(a)
	meanB, varB := meanVar(b)
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

func meanVar(v []float64) (mean, variance float64) {
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