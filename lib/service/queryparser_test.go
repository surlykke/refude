package service

import (
	"testing"
	"fmt"
)

func TestReadTokens(t *testing.T) {
	var s = "Name =i \"firefox\" and Length < 7"
	var ts = readTokens(s)
	for _,t := range ts {
		fmt.Println(t.kind.String(), ":", t.val)
	}
}


var res = struct {
	Name string
	Length int
}{"Firefox", 6}


func TestReadRelation(t *testing.T) {
	m,_ := readRelation(readTokens("Name = \"Firefox\""))

	if !m(res) {
		t.Fail()
	}

	m,_ = readRelation(readTokens("Length < 8"))
	if !m(res) {
		t.Fail()
	}
}

func TestNot(t *testing.T) {
	ts := readTokens("!Name = \"Firefox\"")
	fmt.Println("Got:")
	for _, t := range ts {
		fmt.Println(t.kind.String(), "->", t.val)
	}
	m, _ := readExpr(ts)

	if m(res) {
		t.Fail()
	}
}


func TestAnd(t *testing.T) {
	ts := readTokens("Name = 'Firefox' and Length < 8")
	printTokens(ts)
	m := readAllAsExpr(ts)
	if !m(res) {
		t.Fail()
	}
}

func TestAndNot(t *testing.T) {
	ts := readTokens("! Name = \"Firefox\" and Length < 8")
	printTokens(ts)
	m := readAllAsExpr(ts)
	if m(res) {
		t.Fail()
	}

}

func printTokens(ts tokenStream) {
	for i,t := range ts {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(t.kind.String())
	}
	fmt.Print("\n")
}
