package server

import (
	"context"
	"io"
	"time"

	"github.com/AmitSuresh/playground/playservices/v14/currency/data"
	protos "github.com/AmitSuresh/playground/playservices/v14/currency/protos/currency"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CurrencyServerHandler implements protos.CurrencyServer.
type CurrencyServerHandler struct {
	l   *zap.Logger
	e   *data.ExchangeRatesHandler
	sub map[protos.Currency_SubscribeRatesServer][]*protos.RateRequest

	protos.UnimplementedCurrencyServer
}

// GetCurrencyServerHandler creates a new instance of CurrencyServerHandler.
func GetCurrencyServerHandler(e *data.ExchangeRatesHandler, log *zap.Logger) protos.CurrencyServer {
	c := &CurrencyServerHandler{
		l:   log,
		e:   e,
		sub: make(map[protos.Currency_SubscribeRatesServer][]*protos.RateRequest),
	}
	go c.handleUpdates()
	return c
}

func (c *CurrencyServerHandler) handleUpdates() {
	ru := c.e.MonitorRates(3 * time.Second)
	for range ru {
		c.l.Info("Initialized streaming via handleUpdates")

		for k, v := range c.sub {

			for _, req := range v {
				r, err := c.e.GetRates(req.GetBase().String(), req.GetDestination().String())
				if err != nil {
					c.l.Error("unable to get rates", zap.Error(err), zap.Any("base", req.GetBase().String()), zap.Any("destination", req.GetDestination().String()))
				}

				err = k.Send(&protos.StreamingRateResponse{
					Message: &protos.StreamingRateResponse_RateResponse{
						RateResponse: &protos.RateResponse{Base: req.Base, Destination: req.Destination, Rate: r},
					},
				})
				c.l.Info("sent")

				if err != nil {
					c.l.Error("unable to send rates", zap.Error(err), zap.Any("base", req.GetBase().String()), zap.Any("destination", req.GetDestination().String()))
				}
			}
		}

	}
}

// GetRate implements the GetRate RPC method.
func (c *CurrencyServerHandler) GetRate(ctx context.Context, req *protos.RateRequest) (*protos.RateResponse, error) {
	c.l.Debug("Handling GetRate", zap.Any("base", req.Base), zap.Any("destination", req.Destination))

	if req.Base == req.Destination {
		err := status.Newf(
			codes.InvalidArgument,
			"base currency %s cannot be the same as the destination currency%s ",
			req.Base.String(), req.Destination.String(),
		)
		err, e := err.WithDetails(req)
		if e != nil {
			return nil, e
		}
		return nil, err.Err()
	}

	rate, err := c.e.GetRates(req.GetBase().String(), req.GetDestination().String())
	if err != nil {
		return nil, err
	}

	response := &protos.RateResponse{
		Base:        req.Base,
		Destination: req.Destination,
		Rate:        rate,
	}
	return response, nil
}

func (c *CurrencyServerHandler) SubscribeRates(srv protos.Currency_SubscribeRatesServer) error {
	for {
		req, err := srv.Recv()
		if err == io.EOF {
			c.l.Info("client has closed the connection")
			break
		}
		if err != nil {
			c.l.Error("unable to read from client", zap.Error(err))
			return nil
		}
		c.l.Info("Handle client request", zap.Any("base", req.Base.String()), zap.Any("dest", req.Destination.String()))

		rreq, ok := c.sub[srv]
		if !ok {
			rreq = []*protos.RateRequest{}
		}
		// check if already in the subscribe list and return a custom gRPC error
		for _, r := range rreq {
			// if we already have subscribe to this currency return an error
			if r.Base == req.Base && r.Destination == req.Destination {
				c.l.Error("Subscription already active", zap.Any("base", req.Base.String()), zap.Any("dest", req.Destination.String()))

				grpcError := status.Newf(codes.AlreadyExists, "Subscription already active for rate")
				grpcError, err = grpcError.WithDetails(req)
				if err != nil {
					c.l.Error("Unable to add metadata to error message", zap.Any("error", err))
					continue
				}

				// Can't return error as that will terminate the connection, instead must send an error which
				// can be handled by the client Recv stream.
				srv.Send(&protos.StreamingRateResponse{
					Message: &protos.StreamingRateResponse_Error{
						Error: grpcError.Proto()},
				})
			}
		}

		// all ok add to the collection
		rreq = append(rreq, req)
		c.sub[srv] = rreq
	}
	return nil
}
