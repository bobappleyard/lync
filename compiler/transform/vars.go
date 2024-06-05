package transform

import "slices"

type vars []string

func (vs *vars) clone() vars {
	res := slices.Clone(*vs)
	return res
}

func (vs *vars) add(v string) {
	idx, ok := slices.BinarySearch(*vs, v)
	if ok {
		return
	}
	if idx >= len(*vs) {
		*vs = append(*vs, v)
		return
	}
	vss := *vs
	vss = append(vss, "")
	copy(vss[idx+1:], vss[idx:])
	vss[idx] = v
	*vs = vss
}

func (vs *vars) addAll(vv vars) {
	for _, v := range vv {
		vs.add(v)
	}
}

func (vs *vars) contains(v string) bool {
	_, ok := slices.BinarySearch(*vs, v)
	return ok
}
