package sets

import "sort"

type Empty struct{}

type StringSet map[string]Empty

func NewStringSet(items ...string) StringSet {
	s := make(StringSet)
	s.Insert(items...)
	return s
}

func (s StringSet) Insert(items ...string) {
	for _, item := range items {
		s[item] = Empty{}
	}
}

func (s StringSet) Has(item string) bool {
	_, contained := s[item]
	return contained
}

func (s StringSet) List() []string {
	if len(s) == 0 {
		return nil
	}

	res := make([]string, 0, len(s))
	for key := range s {
		res = append(res, key)
	}
	sort.Strings(res)
	return res
}
