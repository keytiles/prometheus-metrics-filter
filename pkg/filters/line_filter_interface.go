package filters

// Implementors are capable of filtering lines
type ILineFilter interface {

	// Checks the given line and tells if this is a match or not
	EvaluateLine(line string) (isMatch bool)
}
