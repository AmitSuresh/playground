package data

import (
	"encoding/xml"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type Cubes struct {
	CubeData []Cube `xml:"Cube>Cube>Cube"`
}

type Cube struct {
	Currency string `xml:"currency,attr"`
	Rate     string `xml:"rate,attr"`
}

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

// MonitorRates checks the rates in the ECB API every interval and sends a message to the
// returned channel when there are changes
//
// Note: the ECB API only returns data once a day, this function only simulates the changes
// in rates for demonstration purposes
func (e *ExchangeRatesHandler) MonitorRates(interval time.Duration) chan struct{} {
	ret := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			// just add a random difference to the rate and return it
			// this simulates the fluctuations in currency rates
			for k, v := range e.rates {
				// change can be 10% of original value
				change := (rand.Float64() / 10)

				// random direction: 0 or 1
				direction := rand.Intn(2)

				if direction == 0 {
					// new value will be at least 90% of the old
					change = 1 - change
				} else {
					// new value will be at most 110% of the old
					change = 1 + change
				}

				// modify the rate
				e.rates[k] = v * change
			}

			// notify updates, this will block unless there is a listener on the other end
			ret <- struct{}{}
		}
	}()

	return ret
}
