package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"unicode"
	"strings"
)

type ParseError string

type TokenKind int

const (
	Start	      TokenKind = iota
	String
	Identifier
	Integer     //: TODO: Floats
	Boolean     // 'true' or 'false'
	Relation    // See relationOperators below
	SpecialChar // Any char which does not start one of the above
	End
)

// Ordered so that no string is a prefix of a later
var relationOperators = []string{"~i", "~", "<>i", "<>", "=i", "=", "<=", "<", ">=", ">"}

func (tk TokenKind) String() string {
	switch tk {
	case Start:       return "Start"
	case String:      return "String"
	case Identifier:  return "Identifier"
	case Integer:     return "Integer"
	case Boolean:     return "Boolean"
	case Relation:    return "Relation"
	case SpecialChar: return "SpecialChar"
	case End:         return "End"
	default:          return "<undefined>"  
	}
}

type Token struct {
	Kind TokenKind
	Text string

	// Union
	StrVal  string
	NumVal  int
	BoolVal bool
}

func (t Token) String() string {
	return fmt.Sprint(t.Kind, ":", t.Text)
}

func (t Token) assertKind(acceptableKinds ...TokenKind) {
	for _, kind := range acceptableKinds {
		if t.Kind == kind {
			return
		}
	}
	var errorMsg string
	if len(acceptableKinds) == 1 {
		errorMsg = "Expected " + acceptableKinds[0].String()
	} else {
		errorMsg = "Expected one of " + acceptableKinds[0].String()
		for _,acceptableKind := range acceptableKinds[1:] {
			errorMsg = errorMsg + ", " + acceptableKind.String()
		}
	}
	panic(ParseError(errorMsg))
}

func (t Token) assertRaw(acceptableRawValues ...string) {
	for _, rawVal := range acceptableRawValues {
		if t.Text == rawVal {
			return
		}
	}
	var errorMsg string
	if len(acceptableRawValues) == 1 {
		errorMsg = "Expected " + acceptableRawValues[0]
	} else {
		errorMsg = "Expected one of " + acceptableRawValues[0]
		for _,acceptableRawValue := range acceptableRawValues[1:] {
			errorMsg = errorMsg + ", " + acceptableRawValue
		}
	}
	panic(ParseError(errorMsg))
}


type Lexer struct {
	Current      Token
	s            string
	pos          int
	tokenStart   int
}

func MakeLexer(s string) *Lexer {
	var ts = &Lexer{Current: Token{Kind: Start}, s: s}
	return ts
}


func (l *Lexer) currentText() string {
	return l.s[l.tokenStart:l.pos]
}

func (l *Lexer) ch() byte {
	return l.s[l.pos]
}

func (l *Lexer) rune() rune {
	return rune(l.ch())
}

func (l *Lexer) next() {
	for ; l.pos < len(l.s) && unicode.IsSpace(l.rune()); l.pos++ {
	}
	l.tokenStart = l.pos

	if l.pos >= len(l.s) {
		l.Current = Token{Kind: End}
	} else if l.ch() == '"' || l.ch() == '\'' {
		bytes := l.readString()
		var strVal string
		if err := json.Unmarshal(bytes, &strVal); err == nil {
			l.Current = Token{Kind: String, StrVal: strVal}
		} else {
			panic(ParseError("Invalid string syntax: " + err.Error()))
		}
	} else if unicode.IsDigit(l.rune()) || l.ch() == '-' && l.pos < len(l.s)-1 && unicode.IsDigit(l.rune()) {
		for l.pos++; l.pos < len(l.s) && unicode.IsDigit(l.rune()); l.pos++ {
		}
		if numVal, err := strconv.Atoi(l.currentText()); err != nil {
			panic(ParseError("Invalid number: " + err.Error()))
		} else {
			l.Current = Token{Kind: Integer, NumVal: numVal}
		}
	} else if l.s[l.pos] == '_' || unicode.IsLetter(l.rune()) {
		for l.pos++; l.pos < len(l.s) && (l.ch() == '_' || unicode.IsLetter(l.rune()) || unicode.IsDigit(l.rune())); l.pos++ {
		}
		switch l.currentText() {
		case "true":
			l.Current = Token{Kind: Boolean, BoolVal: true}
		case "false":
			l.Current = Token{Kind: Boolean, BoolVal: false}
		default:
			l.Current = Token{Kind: Identifier}
		}
	} else {
		var relOpStr = ""
		for _, s := range relationOperators {
			if strings.HasPrefix(l.s[l.pos:], s) {
				relOpStr = s
				break
			}
		}
		if relOpStr != "" {
			l.pos += len(relOpStr)
			l.Current = Token{Kind:Relation}
		} else {
			l.pos++
			l.Current = Token{Kind:SpecialChar}
		}
	}
	l.Current.Text = l.currentText()
}

func (l *Lexer) readString() []byte {
	if l.ch() == '"' {
		var escaping = false
		for l.pos++ ; l.pos < len(l.s); l.pos++ {
			if l.ch() == '"' && !escaping {
				l.pos++
				return []byte(l.currentText())
			} else {
				escaping = l.ch() == '\\' && !escaping
			}
		}
	} else {

		// We transform the stringliteral from a single quoted to a double quoted
		// (as we want to feed it into json.Unmarshall which only accepts double quoted)
		// So escaped single quotes must be unescaped and unescaped doublequotes must be escaped
		var transformed = make([]byte, 0, 30)
		transformed = append(transformed, '"')

		var escaping = false
		for l.pos++; l.pos < len(l.s); l.pos++ {
			if escaping {
				if l.ch() == '\'' {
					transformed = append(transformed, '\'')
				} else {
					transformed = append(transformed, '\\', l.ch())
				}
				escaping = false
			} else {
				if l.ch() == '\\' {
					escaping = true
				} else if l.ch() == '"' {
					transformed = append(transformed, '\\', '"')
				} else if l.ch() == '\'' {
					transformed = append(transformed, '"')
					l.pos++
					return transformed
				} else {
					transformed = append(transformed, l.ch())
				}
			}
		}
	}
	panic(ParseError("Runaway string"))
}


