package filters

// This composite ILineFilter implementation is matching if any of sub filters is matching
type OrFilter struct {
	Filters []ILineFilter
}

func (filter OrFilter) EvaluateLine(line string) (isMatch bool) {
	for _, childFilter := range filter.Filters {
		if childFilter.EvaluateLine(line) {
			// we are done - this guy is matching
			isMatch = true
			return
		}
	}

	// looks none matched

	return
}
