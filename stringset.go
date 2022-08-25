package logrlint

type stringSet map[string]struct{}

func newStringSet(ss []string) stringSet {
	set := make(stringSet)
	for _, s := range ss {
		set[s] = struct{}{}
	}
	return set
}

func (set stringSet) has(s string) bool {
	_, ok := set[s]
	return ok
}
