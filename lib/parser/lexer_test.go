// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package parser

import (
	"fmt"
	"testing"
)

func TestReadTokens(t *testing.T) {
	var ts = MakeLexer("\"firefox\" 'chrome' and or not Name true false ~i ~ neqi neq eqi eq lte lt gte gt % ( ) [ ]")
	ts.next()
	tokenAssert(ts, String, "\"firefox\"", t)
	tokenAssert(ts, String, "'chrome'", t)
	tokenAssert(ts, Identifier, "and", t)
	tokenAssert(ts, Identifier, "or", t)
	tokenAssert(ts, Identifier, "not", t)
	tokenAssert(ts, Identifier, "Name", t)
	tokenAssert(ts, Boolean, "true", t)
	tokenAssert(ts, Boolean, "false", t)
	tokenAssert(ts, Relation, "~i", t)
	tokenAssert(ts, Relation, "~", t)
	tokenAssert(ts, Relation, "neqi", t)
	tokenAssert(ts, Relation, "neq", t)
	tokenAssert(ts, Relation, "eqi", t)
	tokenAssert(ts, Relation, "eq", t)
	tokenAssert(ts, Relation, "lte", t)
	tokenAssert(ts, Relation, "lt", t)
	tokenAssert(ts, Relation, "gte", t)
	tokenAssert(ts, Relation, "gt", t)
	tokenAssert(ts, SpecialChar, "%", t)
	tokenAssert(ts, SpecialChar, "(", t)
	tokenAssert(ts, SpecialChar, ")", t)
	tokenAssert(ts, SpecialChar, "[", t)
	tokenAssert(ts, SpecialChar, "]", t)
	ts.Current.assertKind(End)
}

func tokenAssert(l *Lexer, kind TokenKind, text string, t *testing.T) {
	if kind != l.Current.Kind || text != l.currentText() {
		fmt.Println("Expected", kind, text, " - got:", l.Current.Kind, l.Current.Text)
		t.Fail()
	}
	l.next()
}
