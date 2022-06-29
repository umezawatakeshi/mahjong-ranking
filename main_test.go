package main

import (
	"testing"
)

func TestParseScore(t *testing.T) {
	if parseScore("0") != 0 {
		t.Error()
	}
	if parseScore("+0") != 0 {
		t.Error()
	}
	if parseScore("-0") != 0 {
		t.Error()
	}
	if parseScore("1") != 10 {
		t.Error()
	}
	if parseScore("+1") != +10 {
		t.Error()
	}
	if parseScore("-1") != -10 {
		t.Error()
	}
	if parseScore("1.3") != 13 {
		t.Error()
	}
	if parseScore("+1.3") != 13 {
		t.Error()
	}
	if parseScore("-1.3") != -13 {
		t.Error()
	}
	if parseScore("0.3") != 3 {
		t.Error()
	}
	if parseScore("+0.3") != 3 {
		t.Error()
	}
	if parseScore("-0.3") != -3 {
		t.Error()
	}
}

func TestScoreToDaten(t *testing.T) {
	Uma = []int{20 + 20, 10, -10, -20}
	Kaeshiten = 30000

	if scoreToDaten(559, 0) != 45900 {
		t.Error()
	}
	if scoreToDaten(47, 1) != 24700 {
		t.Error()
	}
	if scoreToDaten(-190, 2) != 21000 {
		t.Error()
	}
	if scoreToDaten(-416, 3) != 8400 {
		t.Error()
	}
}
