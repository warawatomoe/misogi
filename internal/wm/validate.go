package wm

import (
	"image"
	"math"

	e "misogi/internal/err"
)

var selGains = []float64{0.6, 0.85, 1.0}

const (
	selMaxNb     = 0.05
	selMaxNbInt  = 0.03
	selMinOrigSp = 0.00008
	selMinOrigGr = 0.00004
	selMinImp    = 0.00002
	selMinImpWk  = 0.000008
	selMaxResSp  = 0.00025
)

type Selection struct {
	Config      Config
	Gain        float64
	Spatial     float64
	Accepted    bool
	Improvement float64
}

type selEnt struct {
	cfg Config
	pri int
}

type selEv struct {
	cfg      Config
	gain     float64
	pri      int
	accepted bool
	orig     scrVal
	proc     scrVal
	imp      float64
	nbInc    float64
	rank     float64
}

func ResolveSelection(rgba *image.RGBA, sizeOpt string, force bool) (Selection, *e.E) {
	if rgba == nil {
		return Selection{}, e.New("misogi", e.Call, "wm:select", "nil image")
	}
	b := rgba.Bounds()
	w, h := b.Dx(), b.Dy()

	base, err := cfgDet(w, h, sizeOpt)
	if err != nil {
		return Selection{}, err
	}

	ordered := selOrd(w, h, sizeOpt, base)
	if len(ordered) == 0 {
		return Selection{Config: base}, nil
	}

	baseNB := rmNb(rgba, ordered[0].cfg)

	var best *selEv
	for _, ent := range ordered {
		ev := selEval(rgba, ent.cfg, baseNB, ent.pri, 1)
		if ev == nil || !ev.accepted {
			continue
		}
		if selBetter(best, ev) {
			best = ev
		}
	}

	if best != nil {
		return Selection{
			Config:      best.cfg,
			Gain:        best.gain,
			Spatial:     best.orig.spatial,
			Accepted:    true,
			Improvement: best.imp,
		}, nil
	}

	if force {
		if fb := selForce(rgba, ordered, baseNB, 1); fb != nil {
			return Selection{
				Config:      fb.cfg,
				Gain:        fb.gain,
				Spatial:     fb.orig.spatial,
				Accepted:    true,
				Improvement: fb.imp,
			}, nil
		}
	}

	primary := ordered[0]
	return Selection{
		Config:   primary.cfg,
		Spatial:  scrCfg(rgba, primary.cfg),
		Accepted: false,
	}, nil
}

func selOrd(w, h int, sizeOpt string, base Config) []selEnt {
	raw := candList(w, h, sizeOpt, base)
	forced := sizeOpt != "" && sizeOpt != "auto"

	if forced {
		var filtered []Config
		for _, c := range raw {
			switch sizeOpt {
			case "48":
				if c.Size == 48 {
					filtered = append(filtered, c)
				}
			case "96":
				if c.Size == 96 {
					filtered = append(filtered, c)
				}
			default:
				filtered = append(filtered, c)
			}
		}
		if len(filtered) > 0 {
			raw = filtered
		}
	}

	official := catMatchOff(w, h) != nil
	out := make([]selEnt, 0, len(raw))
	for i, c := range raw {
		out = append(out, selEnt{cfg: c, pri: selPri(c, official, i)})
	}
	return out
}

func selPri(c Config, official bool, idx int) int {
	if official && idx == 0 {
		return 0
	}
	if c.AlphaVariant == "20260520" {
		return 1
	}
	if c.Size == 48 && c.MarginRight == 32 && c.MarginBottom == 32 {
		return 2
	}
	if c.MarginRight >= 96 && c.MarginBottom >= 96 && c.Size <= 72 {
		return 3
	}
	if c.Size != 48 && c.Size != 96 {
		return 4
	}
	return 5 + idx
}

func selEval(rgba *image.RGBA, cfg Config, baseNB float64, pri, passes int) *selEv {
	alpha := AlphaForConfig(cfg)
	if alpha == nil {
		return nil
	}

	orig := scrReg(rgba, cfg, alpha)
	maxNB := selMaxNb
	if cfg.Size != 48 && cfg.Size != 96 {
		maxNB = selMaxNbInt
	}

	var best *selEv
	for _, gain := range selGains {
		trial := imgClone(rgba)
		if trial == nil || Remove(trial, cfg, alpha, gain, passes) != nil {
			continue
		}
		proc := scrReg(trial, cfg, alpha)
		nb := rmNb(trial, cfg)
		ev := &selEv{
			cfg:   cfg,
			gain:  gain,
			pri:   pri,
			orig:  orig,
			proc:  proc,
			imp:   orig.spatial - proc.spatial,
			nbInc: nb - baseNB,
		}
		ev.accepted = selAcc(ev, maxNB)
		ev.rank = selRank(ev)
		if !ev.accepted {
			continue
		}
		if selBetter(best, ev) {
			best = ev
		}
	}
	return best
}

func selForce(rgba *image.RGBA, ordered []selEnt, baseNB float64, passes int) *selEv {
	var best *selEv
	for _, ent := range ordered {
		alpha := AlphaForConfig(ent.cfg)
		if alpha == nil {
			continue
		}
		orig := scrReg(rgba, ent.cfg, alpha)
		for _, gain := range selGains {
			trial := imgClone(rgba)
			if trial == nil || Remove(trial, ent.cfg, alpha, gain, passes) != nil {
				continue
			}
			proc := scrReg(trial, ent.cfg, alpha)
			nb := rmNb(trial, ent.cfg)
			ev := &selEv{
				cfg:      ent.cfg,
				gain:     gain,
				pri:      ent.pri,
				orig:     orig,
				proc:     proc,
				imp:      orig.spatial - proc.spatial,
				nbInc:    nb - baseNB,
				accepted: true,
			}
			ev.rank = selRank(ev)
			if best == nil || ev.rank > best.rank {
				best = ev
			}
		}
	}
	return best
}

func selAcc(ev *selEv, maxNB float64) bool {
	if ev.nbInc > maxNB {
		return false
	}

	weakNM := ev.cfg.AlphaVariant == "20260520"
	hasSig := math.Abs(ev.orig.spatial) >= selMinOrigSp || ev.orig.gradient >= selMinOrigGr
	if weakNM {
		hasSig = hasSig || math.Abs(ev.orig.spatial) >= 0.00002 || ev.orig.gradient >= 0.00001
	}
	if !hasSig {
		return false
	}

	minI := selMinImp
	if weakNM || ev.orig.spatial < selMinOrigSp {
		minI = selMinImpWk
	}
	if ev.imp < minI {
		return false
	}
	if math.Abs(ev.proc.spatial) > selMaxResSp && ev.imp < selMinImp {
		return false
	}
	if ev.proc.gradient > ev.orig.gradient+0.00004 {
		return false
	}
	return true
}

func selBetter(best, ev *selEv) bool {
	if best == nil {
		return true
	}
	if ev.pri < best.pri {
		return true
	}
	if ev.pri > best.pri {
		return false
	}
	if ev.rank > best.rank {
		return true
	}
	return ev.rank == best.rank && ev.gain < best.gain
}

func selRank(ev *selEv) float64 {
	pen := ev.nbInc * 2.0
	if ev.cfg.Size != 48 && ev.cfg.Size != 96 {
		pen *= 2.0
	}
	gradDrop := ev.orig.gradient - ev.proc.gradient
	return ev.imp + gradDrop*0.5 - math.Abs(ev.proc.spatial)*0.25 - pen
}