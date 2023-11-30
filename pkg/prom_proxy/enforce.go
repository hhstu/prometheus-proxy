package prom_proxy

import (
	"fmt"
	"github.com/prometheus/prometheus/pkg/labels"
	"net/url"

	"github.com/prometheus/prometheus/promql/parser"
)

type Enforcer struct {
	labelMatchers  map[string]*labels.Matcher
	errorOnReplace bool
}

func NewEnforcer(errorOnReplace bool, ms ...*labels.Matcher) *Enforcer {
	entries := make(map[string]*labels.Matcher)

	for _, matcher := range ms {
		entries[matcher.Name] = matcher
	}

	return &Enforcer{
		labelMatchers:  entries,
		errorOnReplace: errorOnReplace,
	}
}

func newIllegalLabelMatcherError(existing string, replacement string) IllegalLabelMatcherError {
	return IllegalLabelMatcherError{
		msg: fmt.Sprintf("label matcher value (%s) conflicts with injected value (%s)", existing, replacement),
	}
}

func (ms Enforcer) EnforceNode(node parser.Node) error {
	switch n := node.(type) {
	case *parser.EvalStmt:
		if err := ms.EnforceNode(n.Expr); err != nil {
			return err
		}

	case parser.Expressions:
		for _, e := range n {
			if err := ms.EnforceNode(e); err != nil {
				return err
			}
		}

	case *parser.AggregateExpr:
		if err := ms.EnforceNode(n.Expr); err != nil {
			return err
		}

	case *parser.BinaryExpr:
		if err := ms.EnforceNode(n.LHS); err != nil {
			return err
		}

		if err := ms.EnforceNode(n.RHS); err != nil {
			return err
		}

	case *parser.Call:
		if err := ms.EnforceNode(n.Args); err != nil {
			return err
		}

	case *parser.SubqueryExpr:
		if err := ms.EnforceNode(n.Expr); err != nil {
			return err
		}

	case *parser.ParenExpr:
		if err := ms.EnforceNode(n.Expr); err != nil {
			return err
		}

	case *parser.UnaryExpr:
		if err := ms.EnforceNode(n.Expr); err != nil {
			return err
		}

	case *parser.NumberLiteral, *parser.StringLiteral:
	// nothing to do

	case *parser.MatrixSelector:
		// inject labelselector
		if vs, ok := n.VectorSelector.(*parser.VectorSelector); ok {
			var err error
			vs.LabelMatchers, err = ms.EnforceMatchers(vs.LabelMatchers)
			if err != nil {
				return err
			}
		}

	case *parser.VectorSelector:
		// inject labelselector
		var err error
		n.LabelMatchers, err = ms.EnforceMatchers(n.LabelMatchers)
		if err != nil {
			return err
		}

	default:
		panic(fmt.Errorf("parser.Walk: unhandled node type %T", n))
	}

	return nil
}

func (ms Enforcer) EnforceMatchers(targets []*labels.Matcher) ([]*labels.Matcher, error) {
	var res []*labels.Matcher

	for _, target := range targets {
		if matcher, ok := ms.labelMatchers[target.Name]; ok {
			// matcher.String() returns something like "labelfoo=value"
			if ms.errorOnReplace && matcher.String() != target.String() {
				return res, newIllegalLabelMatcherError(matcher.String(), target.String())
			}
			continue
		}

		res = append(res, target)
	}

	for _, enforcedMatcher := range ms.labelMatchers {
		res = append(res, enforcedMatcher)
	}

	return res, nil
}

const (
	queryParam    = "query"
	matchersParam = "match[]"
)

func EnforceQueryValues(e *Enforcer, v url.Values) (values string, noQuery bool, err error) {
	if v.Get(queryParam) == "" {
		return v.Encode(), false, nil
	}
	expr, err := GetExpr(e, v.Get(queryParam))

	if err != nil {
		return "", true, err
	}
	v.Set(queryParam, expr)
	return v.Encode(), true, nil
}

func GetExpr(e *Enforcer, query string) (string, error) {
	expr, err := parser.ParseExpr(query)
	if err != nil {
		queryParseError := newQueryParseError(err)
		return "", queryParseError
	}

	if err := e.EnforceNode(expr); err != nil {
		if _, ok := err.(IllegalLabelMatcherError); ok {
			return "", err
		}
		enforceLabelError := newEnforceLabelError(err)
		return "", enforceLabelError
	}
	return expr.String(), err
}

type QueryParseError struct {
	msg string
}

func (e QueryParseError) Error() string {
	return e.msg
}

func newQueryParseError(err error) QueryParseError {
	return QueryParseError{msg: fmt.Sprintf("error parsing query string %q", err.Error())}
}

type EnforceLabelError struct {
	msg string
}

func (e EnforceLabelError) Error() string {
	return e.msg
}

func newEnforceLabelError(err error) EnforceLabelError {
	return EnforceLabelError{msg: fmt.Sprintf("error enforcing label %q", err.Error())}
}

type IllegalLabelMatcherError struct {
	msg string
}

func (e IllegalLabelMatcherError) Error() string { return e.msg }

func NewIllegalLabelMatcherError(existing string, replacement string) IllegalLabelMatcherError {
	return IllegalLabelMatcherError{
		msg: fmt.Sprintf("label matcher value (%s) conflicts with injected value (%s)", existing, replacement),
	}
}
