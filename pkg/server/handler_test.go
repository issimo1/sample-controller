package server

import "testing"

func TestGetResource(t *testing.T) {
	i, err := repeat(14)
	if err != nil {
		t.Fatal(err)
	}
	if b, ok := i.([]byte); ok {
		t.Log(string(b))
	}
}
