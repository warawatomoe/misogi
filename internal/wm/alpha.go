package wm

import (
	"embed"
	"encoding/binary"
	"fmt"
	"io/fs"
	"math"
)

//go:embed data/*.bin
var alpFS embed.FS

var (
	alp48    []float32
	alp96    []float32
	alp96Nm  []float32
	alpCache = map[string][]float32{}
)

func init() {
	alp48 = alpLoad("data/alpha48.bin")
	alp96 = alpLoad("data/alpha96.bin")
	alp96Nm = alpLoad("data/alpha96_20260520.bin")
}

func alpLoad(name string) []float32 {
	data, err := fs.ReadFile(alpFS, name)
	if err != nil {
		panic("wm: failed to load alpha mask " + name + ": " + err.Error())
	}
	if len(data)%4 != 0 {
		panic("wm: bad alpha mask size " + name)
	}
	n := len(data) / 4
	out := make([]float32, n)
	for i := 0; i < n; i++ {
		bits := binary.LittleEndian.Uint32(data[i*4:])
		out[i] = math.Float32frombits(bits)
	}
	return out
}

func alpInterp(src []float32, srcSize, dstSize int) []float32 {
	if dstSize <= 0 {
		return nil
	}
	if srcSize == dstSize {
		out := make([]float32, len(src))
		copy(out, src)
		return out
	}

	out := make([]float32, dstSize*dstSize)
	denom := dstSize - 1
	if denom < 1 {
		denom = 1
	}
	scale := float64(srcSize-1) / float64(denom)
	for y := 0; y < dstSize; y++ {
		sy := float64(y) * scale
		y0 := int(sy)
		y1 := y0 + 1
		if y1 >= srcSize {
			y1 = srcSize - 1
		}
		fy := sy - float64(y0)

		for x := 0; x < dstSize; x++ {
			sx := float64(x) * scale
			x0 := int(sx)
			x1 := x0 + 1
			if x1 >= srcSize {
				x1 = srcSize - 1
			}
			fx := sx - float64(x0)

			p00 := src[y0*srcSize+x0]
			p10 := src[y0*srcSize+x1]
			p01 := src[y1*srcSize+x0]
			p11 := src[y1*srcSize+x1]
			top := p00 + (p10-p00)*float32(fx)
			bottom := p01 + (p11-p01)*float32(fx)
			out[y*dstSize+x] = top + (bottom-top)*float32(fy)
		}
	}
	return out
}

func alpSrc(cfg Config) (src []float32, srcSize int) {
	if cfg.MarginRight >= 96 || cfg.MarginBottom >= 96 {
		return alp48, 48
	}
	if cfg.MarginRight == 64 && cfg.MarginBottom == 64 && cfg.Size >= 64 {
		return alp96, 96
	}
	return alp48, 48
}

func alpKey(cfg Config) string {
	_, sz := alpSrc(cfg)
	return fmt.Sprintf("%d:%s:%d", cfg.Size, cfg.AlphaVariant, sz)
}

func Alpha(size int, variant string) []float32 {
	return AlphaForConfig(Config{Size: size, AlphaVariant: variant})
}

func AlphaForConfig(cfg Config) []float32 {
	size := cfg.Size
	variant := cfg.AlphaVariant
	if size <= 0 || size > 192 {
		return nil
	}
	switch variant {
	case "20260520":
		if size == 96 {
			return alp96Nm
		}
		return nil
	case "v2":
		return nil
	case "":
	default:
		return nil
	}

	switch size {
	case 48:
		return alp48
	case 96:
		if variant == "" {
			return alp96
		}
	}

	key := alpKey(cfg)
	if cached, ok := alpCache[key]; ok {
		return cached
	}
	src, srcSize := alpSrc(cfg)
	out := alpInterp(src, srcSize, size)
	alpCache[key] = out
	return out
}