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

func GetExchangeRatesHandler(log *zap.Logger) (*ExchangeRatesHandler, error) {
	e := &ExchangeRatesHandler{
		l:     log,
		rates: map[string]float64{},
	}
	err := e.getRates()

	return e, err
}

func (e ExchangeRatesHandler) GetRates(base, dest string) (float64, error) {
	br, ok := e.rates[base]
	if !ok {
		return 0, fmt.Errorf("rate not found for currency %s", base)
	}

	dr, ok := e.rates[dest]
	if !ok {
		return 0, fmt.Errorf("rate not found for currency %s", dest)
	}

	return br / dr, nil
}

func (e ExchangeRatesHandler) getRates() error {
	resp, err := http.DefaultClient.Get("https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml")
	if err != nil {
		e.l.Error("error attempting GET to URL", zap.Error(err))
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
			e.l.Error("error parsing float", zap.Error(err))
			return err
		}
		e.rates[v.Currency] = r
	}

	e.rates["EUR"] = 1

	return nil
}

type Cubes struct {
	CubeData []Cube `xml:"Cube>Cube>Cube"`
}

type Cube struct {
	Currency string `xml:"currency,attr"`
	Rate     string `xml:"rate,attr"`
}
