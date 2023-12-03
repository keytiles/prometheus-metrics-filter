package filters

// This composite ILineFilter implementation is matching if all sub filters are matching
type AndFilter struct {
	Filters []ILineFilter
}

func (filter AndFilter) EvaluateLine(line string) (isMatch bool) {
	for _, childFilter := range filter.Filters {
		if !childFilter.EvaluateLine(line) {
			// we are done - this guy does not match
			return
		}
	}

	// looks all matched
	isMatch = true

	return
}
