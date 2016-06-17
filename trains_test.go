package main

import "testing"

func TestAngle(t *testing.T) {
	var a, b Angle
	b = a.Add(1)
	if b != 1 {
		t.Error("add 1")
	}
	b = a.Add(11)
	if b != 11 {
		t.Error("add 11")
	}
	b = a.Add(13)
	if b != 1 {
		t.Error("add 13")
	}
	b = a.Add(-1)
	if b != 11 {
		t.Error("add -1 shoud be: 11 is", b)
	}

	b = a.Add(-14)
	if b != 10 {
		t.Error("add -14 shoud be: 10 is", b)
	}
}

func updateChain(pieces []Kind) []Part {
	n := len(pieces)
	parts := make([]Part, n)
	for i, p := range pieces {
		parts[i].Kind = p
		if (i + 1) < n {
			parts[i].Update(&parts[i+1])
		}
	}
	return parts
}

func TestUpdate(t *testing.T) {
	initUpdateTab()

	tests := []struct {
		pieces []Kind
		expEnd Part
	}{
		{[]Kind{Straight, End}, Part{1, 0, 0, End}},
		{[]Kind{Bridge, End}, Part{8, 0, 0, End}},
		{[]Kind{Left, Right, Right, Left, End}, Part{4, 0, 0, End}},
		{[]Kind{Right, Left, Left, Right, End}, Part{4, 0, 0, End}},
		{[]Kind{Left, Left, Left, End}, Part{2, 2, 3, End}},
		{[]Kind{Left, Left, Left, Left, Left, Left, End}, Part{0, 4, 6, End}},
		{[]Kind{Right, Right, Right, End}, Part{2, -2, 9, End}},
		{[]Kind{Right, Right, Right, Right, Right, Right, End}, Part{0, -4, 6, End}},
		{[]Kind{Left, Left, Left, Left, Left, Left, Left, Left, Left, Left, Left, Left, End}, Part{0, 0, 0, End}},
		{[]Kind{Right, Right, Right, Right, Right, Right, Right, Right, Right, Right, Right, Right, End}, Part{0, 0, 0, End}},
		{[]Kind{Left, Left, Left, Left, Left, Left, Right, Right, Right, Right, Right, Right, End}, Part{0, 8, 0, End}},
		{[]Kind{Right, Right, Right, Right, Right, Right, Left, Left, Left, Left, Left, Left, End}, Part{0, -8, 0, End}},
	}

	for _, test := range tests {
		p := updateChain(test.pieces)
		e := p[len(p)-1]
		if !test.expEnd.AtSamePlaceAs(&e) {
			t.Error(test, p)
		}
	}
}

func TestUpdateIkea(t *testing.T) {
	setIkeaParams()
	initUpdateTab()

	tests := []struct {
		pieces []Kind
		expEnd Part
	}{
		{[]Kind{Straight, End}, Part{1, 0, 0, End}},
		{[]Kind{Bridge, End}, Part{2, 0, 0, End}},
		{[]Kind{Left, Left, Right, Right, Right, Right, Left, Left, End}, Part{4, 0, 0, End}},
		{[]Kind{Right, Right, Left, Left, Left, Left, Right, Right, End}, Part{4, 0, 0, End}},
		// {[]Kind{Left, Left, Left, End}, Part{2, 2, 3, End}},
		// {[]Kind{Left, Left, Left, Left, Left, Left, End}, Part{0, 4, 6, End}},
		// {[]Kind{Right, Right, Right, End}, Part{2, -2, 9, End}},
		// {[]Kind{Right, Right, Right, Right, Right, Right, End}, Part{0, -4, 6, End}},
		// {[]Kind{Left, Left, Left, Left, Left, Left, Left, Left, Left, Left, Left, Left, End}, Part{0, 0, 0, End}},
		// {[]Kind{Right, Right, Right, Right, Right, Right, Right, Right, Right, Right, Right, Right, End}, Part{0, 0, 0, End}},
		// {[]Kind{Left, Left, Left, Left, Left, Left, Right, Right, Right, Right, Right, Right, End}, Part{0, 8, 0, End}},
		// {[]Kind{Right, Right, Right, Right, Right, Right, Left, Left, Left, Left, Left, Left, End}, Part{0, -8, 0, End}},
	}

	for _, test := range tests {
		p := updateChain(test.pieces)
		e := p[len(p)-1]
		if !test.expEnd.AtSamePlaceAs(&e) {
			t.Error(test, p)
		}
	}
}
