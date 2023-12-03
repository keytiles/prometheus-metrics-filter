package rules

// Implementors are capable of filtering lines
type ILineEvaluationRule interface {

	// Checks the given line and tells if this line should be included or not in the response
	EvaluateLine(line string) (doInclude bool)
}
