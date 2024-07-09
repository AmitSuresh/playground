package server

import (
	"context"

	"github.com/AmitSuresh/playground/playservices/v14/currency/data"
	protos "github.com/AmitSuresh/playground/playservices/v14/currency/protos/currency"
	"go.uber.org/zap"
)

// CurrencyServerHandler implements protos.CurrencyServer.
type CurrencyServerHandler struct {
	l *zap.Logger
	e *data.ExchangeRatesHandler

	protos.UnimplementedCurrencyServer
}

// GetCurrencyServerHandler creates a new instance of CurrencyServerHandler.
func GetCurrencyServerHandler(e *data.ExchangeRatesHandler, log *zap.Logger) protos.CurrencyServer {
	return &CurrencyServerHandler{
		l: log,
		e: e,
	}
}

// GetRate implements the GetRate RPC method.
func (c *CurrencyServerHandler) GetRate(ctx context.Context, req *protos.RateRequest) (*protos.RateResponse, error) {
	c.l.Debug("Handling GetRate", zap.Any("base", req.Base), zap.Any("destination", req.Destination))

	rate, err := c.e.GetRates(req.GetBase().String(), req.GetDestination().String())
	if err != nil {
		return nil, err
	}
	response := &protos.RateResponse{
		Rate: rate,
	}
	return response, nil
}
