package enginestats

import "sort"

// EngineTotal is one engine's counters over a whole period, for one kind of
// search. Share is its part of the answered searches of that kind, 0 to 1.
type EngineTotal struct {
	Engine   string  `json:"engine"`
	Free     bool    `json:"free"`
	Tried    int64   `json:"tried"`
	Answered int64   `json:"answered"`
	Short    int64   `json:"short"`
	Failed   int64   `json:"failed"`
	TimedOut int64   `json:"timed_out"`
	Share    float64 `json:"share"`
}

// KindTotal is one kind of search over a whole period: how many searches got
// an answer, how many of those a free engine gave and how many a paid one.
type KindTotal struct {
	Kind         string        `json:"kind"`
	Answered     int64         `json:"answered"`
	FreeAnswered int64         `json:"free_answered"`
	PaidAnswered int64         `json:"paid_answered"`
	FreeShare    float64       `json:"free_share"`
	PaidShare    float64       `json:"paid_share"`
	Engines      []EngineTotal `json:"engines"`
}

// Summarize adds rows up per kind and engine. Rows must carry Free.
func Summarize(rows []Row) []KindTotal {
	kinds := map[string]map[string]*EngineTotal{}
	for _, r := range rows {
		if kinds[r.Kind] == nil {
			kinds[r.Kind] = map[string]*EngineTotal{}
		}
		t := kinds[r.Kind][r.Engine]
		if t == nil {
			t = &EngineTotal{Engine: r.Engine, Free: r.Free}
			kinds[r.Kind][r.Engine] = t
		}
		t.Tried, t.Answered, t.Short = t.Tried+r.Tried, t.Answered+r.Answered, t.Short+r.Short
		t.Failed, t.TimedOut = t.Failed+r.Failed, t.TimedOut+r.TimedOut
	}
	out := make([]KindTotal, 0, len(kinds))
	for kind, engines := range kinds {
		out = append(out, kindTotal(kind, engines))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Kind < out[j].Kind })
	return out
}

func kindTotal(kind string, engines map[string]*EngineTotal) KindTotal {
	total := KindTotal{Kind: kind, Engines: make([]EngineTotal, 0, len(engines))}
	for _, e := range engines {
		total.Answered += e.Answered
		if e.Free {
			total.FreeAnswered += e.Answered
		} else {
			total.PaidAnswered += e.Answered
		}
	}
	for _, e := range engines {
		e.Share = share(e.Answered, total.Answered)
		total.Engines = append(total.Engines, *e)
	}
	sort.Slice(total.Engines, func(i, j int) bool { return total.Engines[i].Engine < total.Engines[j].Engine })
	total.FreeShare, total.PaidShare = share(total.FreeAnswered, total.Answered), share(total.PaidAnswered, total.Answered)
	return total
}

func share(part, whole int64) float64 {
	if whole == 0 {
		return 0
	}
	return float64(part) / float64(whole)
}
