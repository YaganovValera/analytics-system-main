package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/YaganovValera/analytics-system/common/logger"
	"github.com/YaganovValera/analytics-system/services/analytics-api/internal/usecase"
	"go.uber.org/zap"
)

type AnalyzeHandler struct {
	Analyzer *usecase.Analyzer
	Log      *logger.Logger
}

func NewAnalyzeHandler(analyzer *usecase.Analyzer, log *logger.Logger) *AnalyzeHandler {
	return &AnalyzeHandler{
		Analyzer: analyzer,
		Log:      log.Named("http.analyze"),
	}
}

func (h *AnalyzeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var candles []usecase.Candle
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.Log.WithContext(ctx).Warn("failed to read body", zap.Error(err))
		h.respond(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	if err := json.Unmarshal(body, &candles); err != nil {
		h.Log.WithContext(ctx).Warn("invalid JSON", zap.Error(err))
		h.respond(w, http.StatusBadRequest, "invalid JSON format")
		return
	}

	resp, err := h.Analyzer.AnalyzeCandles(ctx, candles)
	if err != nil {
		h.Log.WithContext(ctx).Warn("analysis failed", zap.Error(err))
		if err.Error() == "empty input" {
			h.respond(w, http.StatusBadRequest, "empty candle array")
			return
		}
		if len(candles) > 5000 {
			h.respond(w, http.StatusRequestEntityTooLarge, "too many candles (max 5000)")
			return
		}
		h.respond(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.Log.WithContext(ctx).Info("analysis successful", zap.String("symbol", resp.Symbol), zap.Int("count", resp.Count))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AnalyzeHandler) respond(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
