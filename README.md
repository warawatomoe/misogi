# misogi

The mark washes away.

In Shinto practice, misogi is purification — water, ritual, the clearing of
what clings where it should not. Gemini leaves a small spark in the corner.
This tool unblends it and moves on.

A small CLI: read an image, find the watermark box, reverse the alpha blend,
write PNG.

## To use

```sh
./misogi [-hqv] [-p passes] [-g gain] [-s size] input output
```

`input` and `output` may be paths or `-` for stdin/stdout where the pipe
helpers allow it. With no output path, `input_clean.png` is used.

`-h` shows help. `-v` prints the version.

### Size

`-s auto` is the default. It picks among the usual Gemini placements by
matching the corner patch. Use `-s 48` or `-s 96` when you already know which
one you have.

If auto sees no spark signal, the image is passed through unchanged.

### Gain and passes

`-g` sets alpha gain (default `1.0`). At default gain, misogi tries a short
sweep and keeps the least destructive result. `-p` sets removal passes (1–8).

## License

Copyright (c) ともえ (warawatomoe@proton.me)

SPDX-License-Identifier: BSD-2-Clause

See [LICENSE](LICENSE).
