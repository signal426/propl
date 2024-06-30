package propl

type pathSet map[string]bool

// NewpathSet creates a string set.
// Duplicate keys are de-duped.
func newPathSet(paths ...string) pathSet {
	ps := make(pathSet)
	for _, p := range paths {
		ps[p] = false
	}
	return ps
}

func (ps pathSet) empty() bool {
	return ps == nil || len(ps) == 0
}

func (ps pathSet) has(e string) bool {
	_, ok := ps[e]
	return ok
}

func (ps pathSet) claim(e string) {
	ps[e] = true
}

func (ps pathSet) unclaimed() []string {
	unclaimed := []string{}
	for k, v := range ps {
		if !v {
			unclaimed = append(unclaimed, k)
		}
	}
	return unclaimed
}

func (ps pathSet) claimed(e string) bool {
	return ps[e]
}
