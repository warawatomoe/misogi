package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"misogi/internal/arg"
	e "misogi/internal/err"
	"misogi/internal/pipe"
	"misogi/internal/wm"
)

func init() {
	image.RegisterFormat("jpeg", "\xff\xd8", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "\x89PNG\r\n\x1a\n", png.Decode, png.DecodeConfig)
}

var version = "0.0.75"

func main() {
	help := false
	quiet := false
	showVersion := false
	passes := 1
	gain := 1.0
	sizeOpt := "auto"
	force := false

	opts := []arg.Opt{
		{Key: 'h', Typ: arg.TFlg, Dst: &help, Doc: "show help"},
		{Key: 'q', Typ: arg.TFlg, Dst: &quiet, Doc: "suppress non-error output"},
		{Key: 'v', Typ: arg.TFlg, Dst: &showVersion, Doc: "print version"},
		{Key: 'p', Typ: arg.TInt, Dst: &passes, Met: "num", Doc: "removal passes (1..8)"},
		{Key: 'g', Typ: arg.TNum, Dst: &gain, Met: "num", Doc: "alpha gain (>0 and <=4)"},
		{Key: 's', Typ: arg.TStr, Dst: &sizeOpt, Met: "size", Doc: "watermark size: auto, 48, 96"},
		{Key: 'f', Typ: arg.TFlg, Dst: &force, Doc: "force removal even if no strong detection"},
	}
	usage := arg.Usage(opts, []arg.Pos{
		{Name: "input"},
		{Name: "output"},
	})
	res := arg.Parse("misogi", usage, opts)

	if help {
		arg.Help(os.Stdout, "misogi", usage, opts)
		return
	}
	if showVersion {
		fmt.Printf("misogi %s\n", version)
		return
	}
	if passes < 1 || passes > 8 {
		e.Die(e.New("misogi", e.Call, "args", "passes must be 1..8"))
	}
	if gain <= 0.0 || gain > 4.0 {
		e.Die(e.New("misogi", e.Call, "args", "gain must be >0 and <=4"))
	}

	input := ""
	if len(res.Pos) > 0 {
		input = res.Pos[0]
	}
	output := ""
	if len(res.Pos) > 1 {
		output = res.Pos[1]
	}

	r, err := pipe.Reader(input)
	if err != nil {
		e.Die(e.Wrap("misogi", e.Trans, "open", err))
	}
	if r == nil {
		arg.Help(os.Stderr, "misogi", usage, opts)
		e.Die(e.New("misogi", e.Call, "args", "missing input"))
	}
	defer r.Close()

	img, _, err := image.Decode(r)
	if err != nil {
		e.Die(e.Wrap("misogi", e.Trans, "decode", err))
	}

	rgba := wm.ToRGBA(img)

	sel, werr := wm.ResolveSelection(rgba, sizeOpt, force)
	if werr != nil {
		e.Die(werr)
	}

	if sel.Accepted {
		if gain == 1.0 {
			gain = sel.Gain
		}
		alpha := wm.AlphaForConfig(sel.Config)
		if alpha == nil {
			e.Die(e.New("misogi", e.Bug, "wm:alpha", "missing mask"))
		}
		if werr = wm.Remove(rgba, sel.Config, alpha, gain, passes); werr != nil {
			e.Die(werr)
		}
	}

	if output == "" {
		if input != "" {
			ext := filepath.Ext(input)
			base := strings.TrimSuffix(input, ext)
			output = base + "_clean.png"
		}
	}

	w, err := pipe.Writer(output)
	if err != nil {
		e.Die(e.Wrap("misogi", e.Trans, "create", err))
	}
	defer w.Close()

	if err := png.Encode(w, rgba); err != nil {
		e.Die(e.Wrap("misogi", e.Trans, "encode", err))
	}

	_ = quiet
}
