package data

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"

	"go.uber.org/zap"
)

type ExchangeRatesHandler struct {
	l     *zap.Logger
	rates map[string]float64
}

func NewExchangeRatesHandler(log *zap.Logger) (*ExchangeRatesHandler, error) {
	erh := &ExchangeRatesHandler{
		l:     log,
		rates: map[string]float64{},
	}
	err := erh.getRates()

	return erh, err
}

func (erh ExchangeRatesHandler) getRates() error {
	resp, err := http.DefaultClient.Get("https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml")
	if err != nil {
		erh.l.Error("[ERROR]", zap.Any("error attempting GET to URL", err))
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status code 200 got %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	md := &Cubes{}
	xml.NewDecoder(resp.Body).Decode(&md)

	for _, v := range md.CubeData {
		r, err := strconv.ParseFloat(v.Rate, 64)
		if err != nil {
			erh.l.Error("[ERROR]", zap.Any("error parsing float", err))
			return err
		}
		erh.rates[v.Currency] = r
	}

	return nil
}

type Cubes struct {
	CubeData []Cube `xml:"Cube>Cube>Cube"`
}

type Cube struct {
	Currency string `xml:"currency,attr"`
	Rate     string `xml:"rate,attr"`
}
