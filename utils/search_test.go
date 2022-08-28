package utils

import "testing"

func TestContainsString(t *testing.T) {

	if ContainsString("foo", []string{"foo", "bar", "baz"}) == false {
		t.Errorf("string not found")
	}

	if ContainsString("foo", []string{"bar", "baz"}) == true {
		t.Errorf("string found, expected not found")
	}
}
