package service

import (
	"testing"
	"fmt"
)

func TestIntRelations(t *testing.T) {
	var res = struct{I int}{5}
	testHelper(&res, "$.I < 6", true, t)
	testHelper(&res, "$.I < 4", false, t)
	testHelper(&res, "$.I <= 5", true, t)
	testHelper(&res, "$.I <= 4", false, t)
	testHelper(&res, "$.I >= 5", true, t)
	testHelper(&res, "$.I >= 6", false, t)
	testHelper(&res, "$.J >= 4", false, t)
}

func TestStringRelations(t *testing.T) {
	var res = struct{Name string}{"Firefox"}
	testHelper(&res, "$.Name = 'Firefox'", true, t)
	testHelper(&res, "$.Name = 'firefox'", false, t)
	testHelper(&res, "$.Name =i 'firefox'", true, t)
	testHelper(&res, "$.Name <> 'Firefox'", false, t)
	testHelper(&res, "$.Name <> 'irefox'", true, t)
	testHelper(&res, "$.Name <>i 'firefox'", false, t)
	testHelper(&res, "$.Name <>i 'irefox'", true, t)
	testHelper(&res, "$.Name ~ 'irefox'", true, t)
	testHelper(&res, "$.Name ~ 'ireFox'", false, t)
	testHelper(&res, "$.Name ~i 'ireFox'", true, t)
}

func TestBoolRelations(t *testing.T) {
	var res = struct{IsSo bool}{true}
	testHelper(&res, "$.IsSo = true", true, t)
	testHelper(&res, "$.IsSo = false", false, t)
	testHelper(&res, "$.IsSo <> true", false, t)
	testHelper(&res, "$.IsSo <> false", true, t)

}

func TestAbsent(t *testing.T) {
	var res struct{}
	testHelper(&res, "$.Foo = 7", false, t)
	testHelper(&res, "$.Foo <> 7", false, t)
	testHelper(&res, "not $.Foo = 7", true, t)
	testHelper(&res, "not $.Foo <> 7", true, t)
}

func TestSlices(t *testing.T) {
	var res =[]int{4, 6, 8}
	testHelper(&res, "$[0] = 4", true, t)
	testHelper(&res, "$[3] = 4", false, t)
}

func TestWildcard(t *testing.T) {
	var res = []int{4, 7, 9}
	testHelper(&res, "$[*] = 4", true, t)
	testHelper(&res, "$[*] = 7", true, t)
	testHelper(&res, "$[*] = 9", true, t)
	testHelper(&res, "$[*] = 11", false, t)
}

func TestAnd(t *testing.T) {
	var res = struct{Name string; I int}{"Firefox", 7}
	testHelper(&res,"$.Name = 'Firefox' and $.I < 8", true, t);
	testHelper(&res,"$.Name = 'Firefox' and $.I = 8", false, t);
	testHelper(&res,"$.Name = 'Fyrefox' and $.I < 8", false, t);
	testHelper(&res,"$.Name = 'Fyrefox' and $.I = 8", false, t);
}

func TestAndNot(t *testing.T) {
	var res = struct{Name string; Length int}{"Firefox", 7}
	testHelper(&res, "not $.Name = 'Fyrefox' and $.Length < 8", true, t)
	testHelper(&res, "$.Name = 'Firefox' and not $.Length = 8", true, t)
}

func TestNotUndefined(t *testing.T) {
	var res struct {}
	testHelper(&res, "not $.Foo = true", true, t)
}

func TestBracketSyntax(t *testing.T) {
	var res = struct{Name string}{"Firefox"}
	testHelper(&res, "$['Name'] = 'Firefox'", true, t)
	testHelper(&res, "$['Name'] = 'Fyrefox'", false, t)
}

func TestMultiValue1(t *testing.T) {
	var res = struct{Name1 string; Name2 string; Name3 string}{"Firefox", "Chromium", "Opera"}
	testHelper(&res, "$['Name1', 'Name2'] = 'Firefox'", true, t)
	testHelper(&res, "$['Name2', 'Name3'] = 'Firefox'", false, t)
	testHelper(&res, "$['Name4', 'Name3'] = 'Firefox'", false, t)
}

func TestMultiValue2(t *testing.T) {
	var res = []string{"Firefox", "Chromium", "Opera"}
	testHelper(&res, "$[0, 1] = 'Firefox'", true, t)
	testHelper(&res, "$[2, 3] = 'Firefox'", false, t)
	testHelper(&res, "$[4, 3] = 'Firefox'", false, t)
}

func testHelper(resource interface{}, query string, expect bool, t *testing.T) {
	m, err := Parse(query);
	if err != nil {
		fmt.Println("Error:\n" + err.Error())
		t.Fail()
	} else {
		var actual = m(resource)
		fmt.Println(query, ", expect: ", expect, ", actual: ", actual)
		if actual != expect {
			t.Fail()
		}
	}
}


