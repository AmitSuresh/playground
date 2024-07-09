package server

import (
	"context"

	protos "github.com/AmitSuresh/playground/playservices/v14/currency/v2/protos/currency"
	"go.uber.org/zap"
)

// CurrencyServerHandler implements protos.CurrencyServer.
type CurrencyServerHandler struct {
	logger *zap.Logger
	protos.UnimplementedCurrencyServer
}

// GetCurrencyServerHandler creates a new instance of CurrencyServerHandler.
func GetCurrencyServerHandler(logger *zap.Logger) protos.CurrencyServer {
	return &CurrencyServerHandler{
		logger: logger,
	}
}

// GetRate implements the GetRate RPC method.
func (csh *CurrencyServerHandler) GetRate(ctx context.Context, req *protos.RateRequest) (*protos.RateResponse, error) {
	csh.logger.Info("Handling GetRate", zap.Any("base", req.Base), zap.Any("destination", req.Destination))

	response := &protos.RateResponse{
		Rate: 0.5,
	}
	return response, nil
}
