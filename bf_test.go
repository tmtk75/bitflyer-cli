package main

import "testing"

func TestBalanceAsset(t *testing.T) {
	b := Balance{
		Assets: []Asset{
			Asset{
				CurrencyCode: "JPY",
				Amount:       1234,
			},
			Asset{
				CurrencyCode: "USD",
				Amount:       11,
			},
		},
	}

	_, err := b.Asset("EUR")
	if err == nil {
		t.Fatalf("EUR shuold be missing")
	}

	tbl := []struct {
		code   string
		amount float64
	}{
		{"JPY", 1234},
		{"USD", 11},
		{"usD", 11},
	}

	for _, e := range tbl {
		a, err := b.Asset(e.code)
		if err != nil {
			t.Fatalf("%v\n", err)
		}
		if a.Amount != e.amount {
			t.Fatalf("%v should be %v", e.code, e.amount)
		}
	}
}
