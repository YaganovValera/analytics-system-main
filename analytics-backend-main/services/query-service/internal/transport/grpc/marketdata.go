// query-service/internal/transport/grpc/marketdata.go

package grpc

import (
	"context"

	marketdatapb "github.com/YaganovValera/analytics-system/proto/gen/go/v1/marketdata"
	"github.com/YaganovValera/analytics-system/services/query-service/internal/metrics"
	"github.com/YaganovValera/analytics-system/services/query-service/internal/usecase"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type marketdataServer struct {
	marketdatapb.UnimplementedMarketDataServiceServer
	orderbook *usecase.GetOrderBookHandler
	log       *zap.Logger
}

// NewMarketDataServer возвращает gRPC сервер для MarketDataService.
func NewMarketDataServer(orderbook *usecase.GetOrderBookHandler) marketdatapb.MarketDataServiceServer {
	return &marketdataServer{
		orderbook: orderbook,
		log:       zap.L().Named("grpc.marketdata"),
	}
}

// GetOrderBook реализует MarketDataService.GetOrderBook
func (s *marketdataServer) GetOrderBook(ctx context.Context, req *marketdatapb.GetOrderBookRequest) (*marketdatapb.GetOrderBookResponse, error) {
	metrics.GRPCRequestsTotal.WithLabelValues("GetOrderBook").Inc()

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	resp, err := s.orderbook.Handle(ctx, req)
	if err != nil {
		s.log.Warn("GetOrderBook failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "orderbook fetch failed: %v", err)
	}

	return resp, nil
}
