// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package parser

import (
	"testing"
)

func TestIntRelations(t *testing.T) {
	var res = struct{ I int }{5}
	testHelper(&res, "r.I lt 6", true, t)
	testHelper(&res, "r.I lt 4", false, t)
	testHelper(&res, "r.I lte 5", true, t)
	testHelper(&res, "r.I lte 4", false, t)
	testHelper(&res, "r.I gte 5", true, t)
	testHelper(&res, "r.I gte 6", false, t)
	testHelper(&res, "r.J gte 4", false, t)
}

func TestStringRelations(t *testing.T) {
	var res = struct{ Name string }{"Firefox"}
	testHelper(&res, "r.Name eq 'Firefox'", true, t)
	testHelper(&res, "r.Name eq 'firefox'", false, t)
	testHelper(&res, "r.Name eqi 'firefox'", true, t)
	testHelper(&res, "r.Name neq 'Firefox'", false, t)
	testHelper(&res, "r.Name neq 'irefox'", true, t)
	testHelper(&res, "r.Name neqi 'firefox'", false, t)
	testHelper(&res, "r.Name neqi 'irefox'", true, t)
	testHelper(&res, "r.Name ~ 'irefox'", true, t)
	testHelper(&res, "r.Name ~ 'ireFox'", false, t)
	testHelper(&res, "r.Name ~i 'ireFox'", true, t)
}

func TestIdentifierRelations(t *testing.T) {
	var res = struct{ Name string }{"Firefox"}
	testHelper(&res, "r.Name eq Firefox", true, t)
	testHelper(&res, "r.Name eq firefox", false, t)
	testHelper(&res, "r.Name eqi firefox", true, t)
	testHelper(&res, "r.Name neq Firefox", false, t)
	testHelper(&res, "r.Name neq irefox", true, t)
	testHelper(&res, "r.Name neqi firefox", false, t)
	testHelper(&res, "r.Name neqi irefox", true, t)
	testHelper(&res, "r.Name ~ irefox", true, t)
	testHelper(&res, "r.Name ~ ireFox", false, t)
	testHelper(&res, "r.Name ~i ireFox", true, t)
}

func TestBoolRelations(t *testing.T) {
	var res = struct{ IsSo bool }{true}
	testHelper(&res, "r.IsSo eq true", true, t)
	testHelper(&res, "r.IsSo eq false", false, t)
	testHelper(&res, "r.IsSo neq true", false, t)
	testHelper(&res, "r.IsSo neq false", true, t)

}

func TestAbsent(t *testing.T) {
	var res struct{}
	testHelper(&res, "r.Foo eq 7", false, t)
	testHelper(&res, "r.Foo neq 7", false, t)
	testHelper(&res, "not r.Foo eq 7", true, t)
	testHelper(&res, "not r.Foo neq 7", true, t)
}

func TestSlices(t *testing.T) {
	var res = []int{4, 6, 8}
	testHelper(&res, "r[0] eq 4", true, t)
	testHelper(&res, "r[3] eq 4", false, t)
}

func TestWildcard(t *testing.T) {
	var res = []int{4, 7, 9}
	testHelper(&res, "r[%] eq 4", true, t)
	testHelper(&res, "r[%] eq 7", true, t)
	testHelper(&res, "r[%] eq 9", true, t)
	testHelper(&res, "r[%] eq 11", false, t)
}

func TestAnd(t *testing.T) {
	var res = struct {
		Name string
		I    int
	}{"Firefox", 7}
	testHelper(&res, "r.Name eq 'Firefox' and r.I lt 8", true, t)
	testHelper(&res, "r.Name eq 'Firefox' and r.I eq 8", false, t)
	testHelper(&res, "r.Name eq 'Fyrefox' and r.I lt 8", false, t)
	testHelper(&res, "r.Name eq 'Fyrefox' and r.I eq 8", false, t)
}

func TestAndNot(t *testing.T) {
	var res = struct {
		Name   string
		Length int
	}{"Firefox", 7}
	testHelper(&res, "not r.Name eq 'Fyrefox' and r.Length lt 8", true, t)
	testHelper(&res, "r.Name eq 'Firefox' and not r.Length eq 8", true, t)
}

func TestNotUndefined(t *testing.T) {
	var res struct{}
	testHelper(&res, "not r.Foo eq true", true, t)
}

func TestBracketSyntax(t *testing.T) {
	var res = struct{ Name string }{"Firefox"}
	testHelper(&res, "r['Name'] eq 'Firefox'", true, t)
	testHelper(&res, "r['Name'] eq 'Fyrefox'", false, t)
}

func TestMultiValue1(t *testing.T) {
	var res = struct {
		Name1 string
		Name2 string
		Name3 string
	}{"Firefox", "Chromium", "Opera"}
	testHelper(&res, "r['Name1', 'Name2'] eq 'Firefox'", true, t)
	testHelper(&res, "r['Name2', 'Name3'] eq 'Firefox'", false, t)
	testHelper(&res, "r['Name4', 'Name3'] eq 'Firefox'", false, t)
}

func TestMultiValue2(t *testing.T) {
	var res = []string{"Firefox", "Chromium", "Opera"}
	testHelper(&res, "r[0, 1] eq 'Firefox'", true, t)
	testHelper(&res, "r[2, 3] eq 'Firefox'", false, t)
	testHelper(&res, "r[4, 3] eq 'Firefox'", false, t)
}

func testHelper(resource interface{}, query string, expect bool, t *testing.T) {
	m, err := Parse(query)
	if err != nil {
		t.Fail()
	} else {
		var actual = m(resource)
		if actual != expect {
			t.Fail()
		}
	}
}
