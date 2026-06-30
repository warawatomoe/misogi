package main

import (
	"os"

	"misogi/internal/arg"
	e "misogi/internal/err"
	"misogi/internal/pipe"
)

func main() {
	help := false
	quiet := false
	opts := []arg.Opt{
		{Key: 'h', Typ: arg.TFlg, Dst: &help, Doc: "show help"},
		{Key: 'q', Typ: arg.TFlg, Dst: &quiet, Doc: "suppress non-error output"},
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
	if r != nil {
		defer r.Close()
	}

	w, err := pipe.Writer(output)
	if err != nil {
		e.Die(e.Wrap("misogi", e.Trans, "create", err))
	}
	defer w.Close()

	_ = quiet
}
