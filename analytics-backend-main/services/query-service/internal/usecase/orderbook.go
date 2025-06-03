package usecase

import (
	"context"
	"fmt"

	"github.com/YaganovValera/analytics-system/common/logger"
	"github.com/YaganovValera/analytics-system/services/query-service/internal/storage/timescaledb"

	marketdatapb "github.com/YaganovValera/analytics-system/proto/gen/go/v1/marketdata"
)

type GetOrderBookHandler struct {
	repo timescaledb.Repository
	log  *logger.Logger
}

func NewGetOrderBookHandler(repo timescaledb.Repository, log *logger.Logger) *GetOrderBookHandler {
	return &GetOrderBookHandler{
		repo: repo,
		log:  log.Named("usecase.orderbook"),
	}
}

func (h *GetOrderBookHandler) Handle(ctx context.Context, req *marketdatapb.GetOrderBookRequest) (*marketdatapb.GetOrderBookResponse, error) {
	if req.Symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if req.Start == nil || req.End == nil {
		return nil, fmt.Errorf("start and end must be specified")
	}
	start := req.Start.AsTime()
	end := req.End.AsTime()
	if !start.Before(end) {
		return nil, fmt.Errorf("start must be before end")
	}

	pageSize := 100
	if req.Pagination != nil && req.Pagination.PageSize > 0 {
		pageSize = int(req.Pagination.PageSize)
	}
	pageToken := ""
	if req.Pagination != nil {
		pageToken = req.Pagination.PageToken
	}

	snapshots, nextPage, err := h.repo.GetOrderBook(ctx, req.Symbol, start, end, pageSize, pageToken)
	if err != nil {
		return nil, fmt.Errorf("fetch orderbook: %w", err)
	}

	analysis := AnalyzeOrderBookSnapshots(snapshots)

	return &marketdatapb.GetOrderBookResponse{
		Snapshots:     snapshots,
		NextPageToken: nextPage,
		Analysis:      analysis,
	}, nil
}
