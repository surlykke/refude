package service

import (
	"testing"
	"fmt"
)

func TestReadTokens(t *testing.T) {
	var ts = MakeLexer("\"firefox\" 'chrome' and or not Name true false ~i ~ <>i <> =i = <= < >= > * $ ( ) [ ]")
	var expected = []struct{Kind TokenKind; RawVal string}{
		{String, "\"firefox\""},
		{String, "'chrome'"},
		{Identifier, "and"},
		{Identifier, "or"},
		{Identifier, "not"},
		{Identifier, "Name"},
		{Boolean, "true"},
		{Boolean, "false"},
		{Relation, "~i"},
		{Relation, "~"},
		{Relation, "<>i"},
		{Relation, "<>"},
		{Relation, "=i"},
		{Relation, "="},
		{Relation, "<="},
		{Relation, "<"},
		{Relation, ">="},
		{Relation, ">"},
		{SpecialChar, "*"},
		{SpecialChar, "$"},
		{SpecialChar, "("},
		{SpecialChar, ")"},
		{SpecialChar, "["},
		{SpecialChar, "]"}}
	for _, e := range expected {
		fmt.Println("Now: ", ts.Current, ", expecting: ", e)
		ts.Current.assertKind(e.Kind)
		ts.Current.assertRaw(e.RawVal)
		ts.next()
	}
	ts.Current.assertKind(End)
}
