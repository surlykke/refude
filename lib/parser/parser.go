// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package parser

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

/**
 * We implement something like this grammar:
 *
 * EXPR -> EXPR LOGIC_OP EXPR | EXPR AND EXPR | '(' EXPR ') | RELATION
 * LOGIC_OP -> 'and' | 'or'
 * RELATION -> PATHSPEC OP value
 * REL_OP ->
 *
 * with identifier a normal identifier and value a string, number or boolean
 *
 * Rewritten, using normal techniques, to reflect precedence and be non-left-recursive, to:
 *
 * EXPR      -> TERM { 'or' TERM }*
 * TERM      -> FACTOR {'and' FACTOR}*
 * FACTOR    -> '!' FACTOR | '(' EXPR ')' | RELATION
 * RELATION  -> PATHSPEC OP value
 * PATHSPEC  ->
 * REL_OP    ->
 *
 */

type Matcher func(res interface{}) bool

type pathElementKind int

const (
	keys pathElementKind = iota
	indexes
	wildcard
)

type pathElement struct {
	kind    pathElementKind
	keys    []string
	indexes []int
}

//type ParseError struct {
//	parseError ParseError
//	query      string
//	position   int
//}
//
//func (pe *ParseError) Error() string {
//return "Error: " + string(parseError) +
//	       "query: " + strconv.Quote("") +
//		   "pos:   " + strconv.Itoa(l.pos);
//
//}

type ErrorMsg struct {
	query      string
	parseError ParseError
	position   int
}

func (em ErrorMsg) Error() string {
	return "query:    " + strconv.Quote(em.query) + "\n" +
		"error:    " + string(em.parseError) + "\n" +
		"position: " + strconv.Itoa(em.position)
}

func Parse(query string) (m Matcher, err error) {
	err = nil
	var l = MakeLexer(query)
	defer func() {
		r := recover()
		if r != nil {
			if parseError, ok := r.(ParseError); ok {
				err = errors.Errorf("query: %s\nerror:d %s\nat:%d", query, parseError, l.tokenStart)
			} else {
				panic(r)
			}
		}
	}()

	l.next()
	m = readExpr(l)
	l.Current.assertKind(End)

	return
}

func readExpr(ts *Lexer) Matcher {
	var termMatchers = []Matcher{readTerm(ts)}
	for ts.Current.Text == "or" {
		ts.next()
		termMatchers = append(termMatchers, readTerm(ts))
	}

	if len(termMatchers) == 1 {
		return termMatchers[0]
	} else {
		return func(res interface{}) bool {
			for _, t := range termMatchers {
				if t(res) {
					return true
				}
			}
			return false
		}
	}
}

func readTerm(ts *Lexer) Matcher {
	var factorMatchers = []Matcher{readFactor(ts)}
	for ts.Current.Text == "and" {
		ts.next()
		factorMatchers = append(factorMatchers, readFactor(ts))
	}

	if len(factorMatchers) == 1 {
		return factorMatchers[0]
	} else {
		return func(res interface{}) bool {
			for _, f := range factorMatchers {
				if !f(res) {
					return false
				}
			}
			return true
		}
	}
}

func readFactor(ts *Lexer) Matcher {
	var negate = false
	for ts.Current.Text == "not" {
		negate = !negate
		ts.next()
	}

	var m Matcher
	if ts.Current.Text == "(" {
		m = readExpr(ts)
		ts.Current.assertRaw(")")
		ts.next()
	} else {
		m = readRelation(ts)
	}

	if negate {
		return func(res interface{}) bool {
			return !m(res)
		}
	} else {
		return m
	}
}

var stringCheckers = map[string]func(lhs, rhs string) bool{
	"eq":   func(lhs, rhs string) bool { return lhs == rhs },
	"eqi":  func(lhs, rhs string) bool { return strings.ToUpper(lhs) == strings.ToUpper(rhs) },
	"neq":  func(lhs, rhs string) bool { return lhs != rhs },
	"neqi": func(lhs, rhs string) bool { return strings.ToUpper(lhs) != strings.ToUpper(rhs) },
	"~":    func(lhs, rhs string) bool { return strings.Contains(lhs, rhs) },
	"~i":   func(lhs, rhs string) bool { return strings.Contains(strings.ToUpper(lhs), strings.ToUpper(rhs)) },
}

var numberCheckers = map[string]func(lhs, rhs int) bool{
	"eq":  func(lhs, rhs int) bool { return lhs == rhs },
	"neq": func(lhs, rhs int) bool { return lhs != rhs },
	"lt":  func(lhs, rhs int) bool { return lhs < rhs },
	"lte": func(lhs, rhs int) bool { return lhs <= rhs },
	"gt":  func(lhs, rhs int) bool { return lhs > rhs },
	"gte": func(lhs, rhs int) bool { return lhs >= rhs },
}

var boolCheckers = map[string]func(lhs, rhs bool) bool{
	"eq":  func(lhs, rhs bool) bool { return lhs == rhs },
	"neq": func(lhs, rhs bool) bool { return lhs != rhs },
}

func readRelation(ts *Lexer) Matcher {
	var pathSpec = readPathSpec(ts)

	ts.Current.assertKind(Relation)
	var operator = ts.Current
	ts.next()

	ts.Current.assertKind(Identifier, String, Integer, Boolean)
	var value = ts.Current
	ts.next()

	if value.Kind == Identifier {
		if checker, ok := stringCheckers[operator.Text]; ok {
			return buildStringMatcher(pathSpec, value.Text, checker)
		} else {
			panic(ParseError("Operator '" + operator.Text + "' not applicable to string"))
		}
	} else if value.Kind == String {
		if checker, ok := stringCheckers[operator.Text]; ok {
			return buildStringMatcher(pathSpec, value.StrVal, checker)
		} else {
			panic(ParseError("Operator '" + operator.Text + "' not applicable to string"))
		}
	} else if value.Kind == Integer {
		if checker, ok := numberCheckers[operator.Text]; ok {
			return buildNumberMatcher(pathSpec, value.NumVal, checker)
		} else {
			panic(ParseError("Operator '" + operator.Text + "' not applicable to number"))
		}
	} else { // Boolean
		if checker, ok := boolCheckers[operator.Text]; ok {
			return buildBoolMatcher(pathSpec, value.BoolVal, checker)
		} else {
			panic(ParseError("Operator '" + operator.Text + "' not applicable to boolean"))
		}
	}
}

/**
 *  PATHSPEC -> 'r' ELEMENTLIST
 *  ELEMENTLIST -> '[' KEYLIST ']' ELEMENTLIST | <none>
 *  KEYLIST -> STRINGLIST | INTLIST
 *  STRINGLIST -> String ',' STRINGLIST | String
 *  INTLIST -> Integer , INTLIST | string
 */

func readPathSpec(ts *Lexer) []pathElement {
	var pe = make([]pathElement, 0)
	ts.Current.assertRaw("r")
	ts.next()
	for {
		if ts.Current.Text == "." {
			ts.next()
			ts.Current.assertKind(Identifier)
			pe = append(pe, pathElement{kind: keys, keys: []string{ts.Current.Text}})
			ts.next()
		} else if ts.Current.Text == "[" {
			ts.next()
			if ts.Current.Kind == String {
				names := []string{ts.Current.StrVal}
				ts.next()
				for ts.Current.Text == "," {
					ts.next()
					ts.Current.assertKind(String)
					names = append(names, ts.Current.StrVal)
					ts.next()
				}
				pe = append(pe, pathElement{kind: keys, keys: names})
			} else if ts.Current.Kind == Integer {
				numVals := []int{ts.Current.NumVal}
				ts.next()
				for ts.Current.Text == "," {
					ts.next()
					ts.Current.assertKind(Integer)
					numVals = append(numVals, ts.Current.NumVal)
					ts.next()
				}
				pe = append(pe, pathElement{kind: indexes, indexes: numVals})
			} else if ts.Current.Text == "%" {
				pe = append(pe, pathElement{kind: wildcard})
				ts.next()
			} else {
				panic(ParseError("Identifier, integer or wildcard expected"))
			}
			ts.Current.assertRaw("]")
			ts.next()
		} else {
			break
		}
	}
	return pe
}

func buildStringMatcher(pathSpec []pathElement, rhs string, checker func(string, string) bool) Matcher {
	return func(res interface{}) bool {
		for _, node := range extractFieldValues(pathSpec, res) {
			if lhs, ok := node.(string); ok && checker(lhs, rhs) {
				return true
			}
		}
		return false
	}
}

func buildNumberMatcher(pathSpec []pathElement, rhs int, checker func(int, int) bool) Matcher {
	return func(res interface{}) bool {
		for _, node := range extractFieldValues(pathSpec, res) {
			if lhs, ok := node.(int); ok && checker(lhs, rhs) {
				return true
			}
		}
		return false
	}
}

func buildBoolMatcher(pathSpec []pathElement, rhs bool, checker func(bool, bool) bool) Matcher {
	return func(res interface{}) bool {
		for _, node := range extractFieldValues(pathSpec, res) {
			if lhs, ok := node.(bool); ok && checker(lhs, rhs) {
				return true
			}
		}
		return false
	}
}

func extractFieldValues(pathSpec []pathElement, res interface{}) []interface{} {
	var leafs = []interface{}{}
	fieldCollector(pathSpec, res, &leafs)
	return leafs
}

func fieldCollector(pathSpec []pathElement, node interface{}, leafs *[]interface{}) {
	v := reflect.ValueOf(node)
	for v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}
	if len(pathSpec) == 0 {
		(*leafs) = append(*leafs, v.Interface())
	} else {
		if v.Kind() == reflect.Struct {
			if pathSpec[0].kind == keys {
				for _, fieldName := range pathSpec[0].keys {
					if f := v.FieldByName(fieldName); f.Kind() != reflect.Invalid && f.CanSet() {
						if f != reflect.Zero(f.Type()) {
							fieldCollector(pathSpec[1:], f.Addr().Interface(), leafs)
						}
					}
				}
			} else if pathSpec[0].kind == wildcard {
				for i := 0; i < v.NumField(); i++ {
					if f := v.Field(i); f.CanSet() && f != reflect.Zero(f.Type()) {
						fieldCollector(pathSpec[1:], f, leafs)
					}
				}
			}
		} else if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			if pathSpec[0].kind == indexes {
				for _, i := range pathSpec[0].indexes {
					if 0 <= i && i < v.Len() {
						fieldCollector(pathSpec[1:], v.Index(int(i)).Addr().Interface(), leafs)
					}
				}
			} else if pathSpec[0].kind == wildcard {
				for i := 0; i < v.Len(); i++ {
					fieldCollector(pathSpec[1:], v.Index(i).Addr().Interface(), leafs)
				}
			}
		}
	}
}
