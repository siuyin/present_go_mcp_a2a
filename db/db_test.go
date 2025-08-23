package db

import "testing"

func TestGet(t *testing.T) {
	dat := []struct {
		i string
		o string
	}{
		{"iPhone 14", "ID=1, product name=iPhone 14, price in USD=899, quantity in stock=0"},
		{"iPhone 16", "Sorry we do not have any iPhone 16"},
		{"simpleX", "ID=3, product name=simpleX, price in USD=199, quantity in stock=32"},
		{"iPhone 15", "ID=2, product name=iPhone 15, price in USD=1,099, quantity in stock=56"},
	}
	for i, d := range dat {
		s := Get(d.i)
		if s != d.o {
			t.Errorf("case %d: expected: %q, got: %q\n", i, s, d.o)
		}
	}
}
