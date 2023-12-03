package rules

import (
	"fmt"

	"github.com/keytiles/prometheus-metrics-filter/pkg/conf"
	"github.com/keytiles/prometheus-metrics-filter/pkg/filters"
)

// Takes a metric line and decides if this should be included in the response or not
//
// It has 'remove' and 'include' voters. And 'include' is stronger so even if removers voted YES includers can over rule this decision
type IncludeRemoveRule struct {
	remove  []filters.ILineFilter // If any of the filters is TRUE then it will vote for "let's remove!"
	include []filters.ILineFilter // If any of the filters is TRUE then it will vote for "let's add!" and this will outrule a possible 'remove' vote
}

func NewIncludeRemoveRule(fromConfigRule conf.ProxyRule) (rule IncludeRemoveRule, failure error) {

	rule.remove = make([]filters.ILineFilter, 0, len(fromConfigRule.Remove))
	for idx, removeItem := range fromConfigRule.Remove {
		filter, err := createFilterFromLineMatchRule(removeItem)
		if err != nil {
			failure = fmt.Errorf("problem at #%v entry in 'remove' block: %v", idx, failure)
			return
		}
		rule.remove = append(rule.remove, filter)
	}

	rule.include = make([]filters.ILineFilter, 0, len(fromConfigRule.Include))
	for idx, includeItem := range fromConfigRule.Include {
		filter, err := createFilterFromLineMatchRule(includeItem)
		if err != nil {
			failure = fmt.Errorf("problem at #%v entry in 'include' block: %v", idx, failure)
			return
		}
		rule.include = append(rule.include, filter)
	}

	return
}

func createFilterFromLineMatchRule(lineMatchRule conf.LineMatchRule) (filter filters.ILineFilter, failure error) {

	if len(lineMatchRule.And) > 0 {
		if len(lineMatchRule.Or) > 0 {
			failure = fmt.Errorf("you can not use 'and' and 'or' at the same time in a rule entry")
			return
		}
		if lineMatchRule.Negate {
			failure = fmt.Errorf("you can not negate an 'and' expression")
			return
		}

		childFilters := make([]filters.ILineFilter, 0, len(lineMatchRule.And))
		for idx, component := range lineMatchRule.And {
			childFilter, err := createFilterFromLineMatchRule(component)
			if err != nil {
				failure = fmt.Errorf("problem at #%v AND entry: %v", idx, err)
				return
			}
			childFilters = append(childFilters, childFilter)
		}

		filter = filters.AndFilter{
			Filters: childFilters,
		}

	} else if len(lineMatchRule.Or) > 0 {
		if len(lineMatchRule.And) > 0 {
			failure = fmt.Errorf("you can not use 'and' and 'or' at the same time in a rule entry")
			return
		}
		if lineMatchRule.Negate {
			failure = fmt.Errorf("you can not negate an 'or' expression")
			return
		}

		childFilters := make([]filters.ILineFilter, 0, len(lineMatchRule.Or))
		for idx, component := range lineMatchRule.Or {
			childFilter, err := createFilterFromLineMatchRule(component)
			if err != nil {
				failure = fmt.Errorf("problem at #%v OR entry: %v", idx, err)
				return
			}
			childFilters = append(childFilters, childFilter)
		}

		filter = filters.OrFilter{
			Filters: childFilters,
		}

	} else if lineMatchRule.Regexp != "" {
		filter, failure = filters.NewRegexpFilter(lineMatchRule.Regexp, lineMatchRule.Negate)
		if failure != nil {
			failure = fmt.Errorf("failed to compile RegularExpression filter: %v", failure)
			return
		}

	} else {
		failure = fmt.Errorf("you must specify one of 'and', 'or' or 'regexp' entry")
		return
	}

	return
}

func (rule IncludeRemoveRule) EvaluateLine(line string) (doInclude bool) {

	// by default we will include this line
	doInclude = true

	// as 'include' is stronger start with them
	for _, includer := range rule.include {
		if includer.EvaluateLine(line) {
			// we are done - this line is in the game
			return
		}
	}

	// now let's check the 'removers'
	for _, remover := range rule.remove {
		if remover.EvaluateLine(line) {
			// we are done - this line is out!
			doInclude = false
			return
		}
	}

	return
}
