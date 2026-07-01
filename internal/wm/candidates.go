package wm

func candList(w, h int, sizeOpt string, base Config) []Config {
	cs := catEnts(w, h)
	if len(cs) == 0 {
		cs = []Config{base}
	}

	cands := make([]Config, len(cs))
	copy(cands, cs)

	forced := sizeOpt != "" && sizeOpt != "auto"

	if !forced {
		alt := 48
		if base.Size == 48 {
			alt = 96
		}
		hasAlt := false
		for _, c := range cands {
			if c.Size == alt {
				hasAlt = true
				break
			}
		}
		if !hasAlt {
			cands = append(cands, cfgStd(alt))
		}
	}

	if !forced || sizeOpt == "48" {
		if fits(w, h, 48, 96, 96) {
			cands = append(cands, Config{Size: 48, MarginRight: 96, MarginBottom: 96})
		}
	}

	if !forced {
		if c := candLg48(w, h); c.Size > 0 {
			cands = append(cands, c)
		}
	}

	seen := map[Config]bool{}
	var out []Config
	for _, c := range cands {
		if !seen[c] {
			seen[c] = true
			out = append(out, c)
		}
	}
	return out
}

func candLg48(w, h int) Config {
	c := Config{Size: 48, MarginRight: 96, MarginBottom: 96}
	if !fits(w, h, c.Size, c.MarginRight, c.MarginBottom) {
		return Config{}
	}
	return c
}