package data

import (
	"testing"

	"go.uber.org/zap"
)

func TestNewRates(t *testing.T) {
	l, _ := zap.NewProduction()
	tr, err := GetExchangeRatesHandler(l)
	if err != nil {
		t.Fatal(err)
	}
	tr.l.Info("[INFO]", zap.Any("tr", tr.rates))
}
