package service

import (
	"fmt"
	"unicode"
	"strconv"
	"reflect"
	"encoding/json"
	"strings"
	"errors"
)

/**
 * We implement something like this grammar:
 *
 * EXPR -> EXPR LOGIC_OP EXPR | EXPR AND EXPR | '(' EXPR ') | RELATION
 * LOGIC_OP -> 'and' | 'or'
 * RELATION -> identifier OP value
 * REL_OP ->
 *
 * with identifier a normal identifier and value a string, number or boolean
 *
 * Rewritten, using normal techniques, to reflect precedence and be non-left-recursive, to:
 *
 * EXPR      -> TERM { 'or' TERM }*
 * TERM      -> FACTOR {'and' FACTOR}*
 * FACTOR    -> '!' FACTOR | '(' EXPR ')' | RELATION
 * RELATION  -> identifier OP value
 * REL_OP    -> '=', '<', '<=', '>' '>='
 *
 */

type Matcher func(res interface{}) bool

type ParseError error

func str2Err(msg string) ParseError {
	return ParseError(errors.New(msg))
}

func err2Err(err error) ParseError {
	return ParseError(err)
}

func parseQuery(query string) (m Matcher, err error) {
	err = nil
	defer func() {
		if r := recover(); r != nil {
			if parseError, ok := r.(ParseError); ok {
				err = parseError
			}
		}
	}()

	m = readAllAsExpr(readTokens(query))
	return
}

func readAllAsExpr(ts tokenStream) Matcher {
	var m Matcher
	m, ts = readExpr(ts)
	if ts[0].kind != end {
		panic(str2Err("Traling characters"))
	}

	return m
}


type tokenKind int

const (
	identifier tokenKind = iota
	str
	boolean
	number
	// Operators (arranged so that none is a prefix of a later)
	or         // or
	and        // and
	not        // !
	lp         // (
	rp         // )
	eqi        // =i
	eq         // =
	has        // has
	hastilde   // has~
	hastildei  // has~i
	tildei     // ~i
	tilde      // ~
	lte        // <=
	lt         // <
	gte        // <=
	gt         // <
	end
)

func (tk tokenKind) String() string {
	switch tk {
	case identifier:
		return "identifier"
	case str:
		return "str"
	case boolean:
		return "boolean"
	case number:
		return "number"
	case or:
		return "or"
	case and:
		return "and"
	case not:
		return "!"
	case lp:
		return "("
	case rp:
		return ")"
	case eqi:
		return "=i"
	case eq:
		return "="
	case hastildei:
		return "has~i"
	case hastilde:
		return "has~"
	case has:
		return "has"
	case tildei:
		return "~i"
	case tilde:
		return "~"
	case lte:
		return "<="
	case lt:
		return "<"
	case gte:
		return ">="
	case gt:
		return ">"
	case end:
		return "end"
	default:
		return strconv.Itoa(int(tk))
	}
}

func oneOf(t tokenKind, kinds ...tokenKind) bool {
	for _, tk := range kinds {
		if t == tk {
			return true
		}
	}

	return false
}

type token struct {
	kind      tokenKind
	val       string
	strVal    string
	numberVal int64
	boolVal   bool
}

type tokenStream []token

func next(ts tokenStream, acceptableKinds ...tokenKind) (token, tokenStream, bool) {
	for _, kind := range acceptableKinds {
		if (ts[0].kind == kind) {
			return ts[0], ts[1:], true
		}
	}
	return ts[0], ts, false
}

func readTokens(s string) tokenStream {

	var strPos = 0

	var skipSpace = func() int {
		var i = strPos
		for i < len(s) && unicode.IsSpace(rune(s[i])) {
			i++
		}
		return i
	}

	var readEnd = func() (token, int, bool) {
		if strPos >= len(s) {
			return token{end, "", "", 0, false}, len(s), true
		} else {
			return token{}, strPos, false
		}
	}

	var readOperator = func() (token, int, bool) {
		for kind := or; kind < end; kind++ {
			str := kind.String()
			if strings.HasPrefix(s[strPos:], str) {
				return token{kind, str, "", 0, false}, strPos + len(str), true
			}
		}
		return token{}, strPos, false
	}

	var readIdentifier = func() (token, int, bool) {
		if s[strPos] == '_' || unicode.IsLetter(rune(s[strPos])) {
			var i = strPos + 1
			for ; i < len(s) && (s[i] == '_' || unicode.IsLetter(rune(s[i])) || unicode.IsDigit(rune(s[i]))); i++ {
			}
			return token{identifier, s[strPos:i], "", 0, false}, i, true
		} else {
			return token{}, strPos, false
		}
	}

	var readNumber = func() (token, int, bool) { // TODO Handle non-ints
		if unicode.IsDigit(rune(s[strPos])) {
			var i = strPos
			for i++; i < len(s) && unicode.IsDigit(rune(s[i])); i++ {
			}
			if numVal, err := strconv.ParseInt(s[strPos:i], 10, 64); err == nil {
				return token{number, s[strPos:i], "", numVal, false}, i, true
			} else {
				panic(err2Err(err))
			}
		} else {
			return token{}, strPos, false
		}
	}

	var readDoublequotedString = func() (token, int, bool) {
		if s[strPos] == '"' {
			var escaping = false
			var i = strPos + 1
			for i := i + 1; i < len(s); i++ {
				if s[i] == '"' && !escaping {
					i++
					var val string
					if err := json.Unmarshal([]byte(s[strPos:i]), &val); err != nil {
						panic(str2Err(fmt.Sprintf("Problem with string at %d, %s", strPos, err.Error())))
					}
					return token{str, s[strPos:i], val, 0, false}, i, true
				} else {
					escaping = s[i] == '\\' && !escaping
				}
			}

			panic(str2Err(fmt.Sprintf("Runaway string at %d", strPos)))
		} else {
			return token{}, strPos, false
		}
	}

	var readSinglequotedString = func() (token, int, bool) {
		if s[strPos] == '\'' {
			var i = strPos
			var transformed = make([]byte, 0, 30)
			transformed = append(transformed, '"')

			// We transform the string from a single quoted to a double quoted
			// So escaped single quotes must be unescaped and unescaped doublequotes must be escaped
			var escaping = false
			for i++; i < len(s); i++ {
				if escaping {
					if s[i] == '\'' {
						transformed = append(transformed, '\'')
					} else {
						transformed = append(transformed, '\\', s[i])
					}
					escaping = false
				} else {
					if s[i] == '\\' {
						escaping = true
					} else if s[i] == '"' {
						transformed = append(transformed, '\\', '"')
					} else if s[i] == '\'' {
						transformed = append(transformed, '"')
						i++
						break
					} else {
						transformed = append(transformed, s[i])
					}
				}
			}
			if transformed[len(transformed)-1] != '"' {
				panic(str2Err(fmt.Sprintf("Runaway string at %d", strPos)))
			}

			var val string
			if err := json.Unmarshal(transformed, &val); err != nil {
				panic(str2Err(fmt.Sprintf("Problem with string at %d, %s", strPos, err.Error())))
			}

			return token{str, s[strPos:i], val, 0, false}, i, true
		} else {
			return token{}, strPos, false
		}
	}

	var readLeftPar = func() (token, int, bool) {
		if s[strPos] == '(' {
			return token{lp, "(", "", 0, false}, strPos + 1, true
		} else {
			return token{}, strPos, false
		}
	}

	var readRightPar = func() (token, int, bool) {
		if s[strPos] == ')' {
			return token{lp, ")", "", 0, false}, strPos + 1, true
		} else {
			return token{}, strPos, false
		}
	}

	var tRs = []func() (token, int, bool){
		readEnd, readOperator, readIdentifier, readNumber, readDoublequotedString, readSinglequotedString, readLeftPar, readRightPar,
	}

	var ts = make(tokenStream, 0, 20)
	for {
		strPos = skipSpace()

		var tok = token{}
		var ok bool

		for _, tokenReader := range tRs {
			if tok, strPos, ok = tokenReader(); ok {
				ts = append(ts, tok)
				break;
			}
		}

		if !ok {
			panic(str2Err(fmt.Sprintf("Unexpected character %c at %d", s[strPos], strPos)))
		} else if tok.kind == end {
			break
		}
	}
	return ts
}

func readExpr(ts tokenStream) (Matcher, tokenStream) {
	var termMatchers = []Matcher{}
	var termMatcher Matcher
	for {
		termMatcher, ts = readTerm(ts)
		termMatchers = append(termMatchers, termMatcher)
		var ok bool
		if _, ts, ok = next(ts, or); !ok {
			break
		}
	}

	fmt.Println("readExpr", len(termMatchers), "termMatchers")

	if len(termMatchers) == 1 {
		return termMatchers[0], ts
	} else {
		return func(res interface{}) bool {
			for _, t := range termMatchers {
				if  t(res) {
					return true
				}
			}
			return false
		}, ts
	}
}

func readTerm(ts tokenStream) (Matcher, tokenStream) {
	var factorMatchers = []Matcher{}
	var factorMatcher Matcher
	for {
		factorMatcher, ts = readFactor(ts)
		factorMatchers = append(factorMatchers, factorMatcher)
		var ok bool
		if _, ts, ok = next(ts, and); !ok {
			break
		}
	}

	fmt.Println("readExpr", len(factorMatchers), "factorMatchers")

	if len(factorMatchers) == 1 {
		return factorMatchers[0], ts
	} else {
		return func(res interface{}) bool {
			for _, f := range factorMatchers {
				if ! f(res) {
					return false
				}
			}
			return true
		}, ts
	}
}

func readFactor(ts tokenStream) (Matcher, tokenStream) {
	fmt.Println("In readFactor, first: ", ts[0].kind.String())
	var negate, ok bool = false, false
	var count = 0;
	for {
		if _, ts, ok = next(ts, not); ok {
			negate = ! negate
		} else {
			break
		}
		count++
		if count > 5 {
			break
		}

	}

	fmt.Println("now first: ", ts[0].kind.String())

	var m Matcher
	if _, ts, ok = next(ts, lp); ok {
		m, ts = readExpr(ts)
		if _, ts, ok = next(ts, rp); !ok {
			panic(str2Err("Right paranthesis expected"))
		}
	} else {
		m, ts = readRelation(ts)
	}

	if negate {
		return func(res interface{}) bool {
			return ! m(res)
		}, ts
	} else {
		return m, ts
	}
}

func readRelation(ts tokenStream) (Matcher, tokenStream) {
	var field, operator, value token
	var ok bool

	if field, ts, ok = next(ts, identifier); !ok {
		panic(str2Err(fmt.Sprintf("Identifier expected got %s", field.val)))
	}

	if operator, ts, ok = next(ts, lp, rp, eqi, eq, has, hastilde, hastildei, tildei, tilde, lte, lt, gte, gt); !ok {
		panic(str2Err(fmt.Sprintf("Relational operator expected, got %s", ts[0].val)))
	}

	if value, ts, ok = next(ts, str, boolean, number); !ok {
		panic(str2Err(fmt.Sprintf("String, number or boolean expected, got %s", value.val)))
	}

	switch value.kind {
	case str:
		switch operator.kind {
		case eq:
			return func(res interface{}) bool {
				if fieldVal, ok := extractString(res, field.val); ok {
					return fieldVal == value.strVal
				} else {
					return false
				}
			}, ts
		case eqi:
			return func(res interface{}) bool {
				if fieldVal, ok := extractString(res, field.val); ok {
					return strings.ToUpper(fieldVal) == strings.ToUpper(value.strVal)
				} else {
					return false
				}
			}, ts
		case tilde:
			return func(res interface{}) bool {
				if fieldVal, ok := extractString(res, field.val); ok {
					return strings.Contains(fieldVal, value.strVal)
				} else {
					return false
				}
			}, ts
		case tildei:
			return func(res interface{}) bool {
				if fieldVal, ok := extractString(res, field.val); ok {
					return strings.Contains(strings.ToUpper(fieldVal), strings.ToUpper(value.strVal))
				} else {
					return false
				}
			}, ts
		case has:
			return func(res interface{}) bool {
				if fieldVal, ok := extractStringSlice(res, field.val); ok {
					if find(fieldVal, value.strVal, false, true) {
						return true
					}
				}
				return false
			}, ts
		case hastilde:
			return func(res interface{}) bool {
				if fieldVal, ok := extractStringSlice(res, field.val); ok {
					if find(fieldVal, value.strVal, true, true) {
						return true
					}
				}
				return false
			}, ts
		case hastildei:
			return func(res interface{}) bool {
				if fieldVal, ok := extractStringSlice(res, field.val); ok {
					if find(fieldVal, value.strVal, true, false) {
						return true
					}
				}
				return false
			}, ts
		default:
			panic(str2Err(fmt.Sprintf("Operator '%s' may not be applied to string at %d", operator.val, 0)))
		}

	case boolean:
		switch operator.kind {
		case eq:
			return func(res interface{}) bool {
				if fieldVal, ok := extractBool(res, field.val); ok {
					return fieldVal == value.boolVal
				} else {
					return false
				}
			}, ts
		default:
			panic(str2Err(fmt.Sprintf("Operator '%s' may not be applied to boolean at %d", operator.val, 0)))
		}
	case number:
		switch operator.kind {
		case eq:
			return func(res interface{}) bool {
				if fv, ok := extractNumber(res, field.val); ok {
					return fv == value.numberVal
				} else {
					return false
				}
			}, ts
		case lt:
			return func(res interface{}) bool {
				if fv, ok := extractNumber(res, field.val); ok {
					return fv < value.numberVal
				} else {
					return false
				}
			}, ts
		case lte:
			return func(res interface{}) bool {
				if fv, ok := extractNumber(res, field.val); ok {
					return fv <= value.numberVal
				} else {
					return false
				}
			}, ts
		case gt:
			return func(res interface{}) bool {
				if fv, ok := extractNumber(res, field.val); ok {
					return fv > value.numberVal
				} else {
					return false
				}
			}, ts
		case gte:
			return func(res interface{}) bool {
				if fv, ok := extractNumber(res, field.val); ok {
					return fv >= value.numberVal
				} else {
					return false
				}
			}, ts

		default:
			panic(str2Err(fmt.Sprintf("Only eq, lt, lte, gt, gte can be applied to number, got %s", operator.val)))
		}
	default:
		panic(str2Err(fmt.Sprintf("One of string, boolean, number expected at %d, but got %s", 0, ts[0].val)))
	}
}

func extractField(res interface{}, fieldName string) (reflect.Value, bool) {
	v := reflect.ValueOf(res)

	for v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}

	if v.Kind() == reflect.Struct {
		f := v.FieldByName(fieldName)
		return f, f != reflect.Value{}
	} else {
		return reflect.Value{}, false
	}
}

func find(stringSlice []string, val string, contain bool, casesensitive bool) bool {
	for _, s := range stringSlice {
		if contain && casesensitive && strings.Contains(s, val) {
			return true
		} else if contain && !casesensitive && strings.Contains(strings.ToUpper(s), strings.ToUpper(val)) {
			return true
		} else if !contain && casesensitive && s == val {
			return true
		} else if !contain && !casesensitive && strings.ToUpper(s) == strings.ToUpper(val) {
			return false
		}
	}

	return false
}

func extractString(res interface{}, fieldName string) (string, bool) {
	if f, ok := extractField(res, fieldName); ok && f.Kind() == reflect.String {
		return f.String(), true
	} else {
		return "", false
	}
}

func extractStringSlice(res interface{}, fieldName string) ([]string, bool) {
	if f, ok := extractField(res, fieldName); ok {
		if slice, ok2 := f.Interface().([]string); ok2 {
			return slice, ok
		}
	}
	return []string{}, false
}

func extractBool(res interface{}, fieldName string) (bool, bool) {
	if f, ok := extractField(res, fieldName); ok && f.Kind() == reflect.Bool {
		return f.Bool(), true
	} else {
		return false, false
	}
}

func extractNumber(res interface{}, fieldName string) (int64, bool) {
	if f, ok := extractField(res, fieldName); ok {
		switch f.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return f.Int(), true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return int64(f.Uint()), true
		}
	}
	return 0, false
}
