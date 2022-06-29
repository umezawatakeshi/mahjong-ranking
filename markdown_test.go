package main

import (
	"testing"
)

func TestEscapeString(t *testing.T) {
	if escapeString(`hoge_fuga~foo*bar\baz`) != `hoge\_fuga\~foo\*bar\\baz` {
		t.Error()
	}
}
