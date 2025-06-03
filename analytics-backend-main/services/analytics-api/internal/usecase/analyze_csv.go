// analytics-api/internal/usecase/analyze_csv.go
package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/YaganovValera/analytics-system/common/logger"
)

type Candle struct {
	Symbol    string    `json:"symbol"`
	OpenTime  time.Time `json:"open_time"`
	CloseTime time.Time `json:"close_time"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
}

type AnalyticsResponse struct {
	Symbol    string    `json:"symbol"`
	Interval  string    `json:"interval"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Count     int       `json:"count"`
	Analytics struct {
		AvgClose           float64 `json:"avg_close"`
		SumVolume          float64 `json:"sum_volume"`
		PriceChange        float64 `json:"price_change"`
		Volatility         float64 `json:"volatility"`
		UpCount            int     `json:"up_count"`
		DownCount          int     `json:"down_count"`
		UpRatio            float64 `json:"up_ratio"`
		DownRatio          float64 `json:"down_ratio"`
		MaxCandle          Candle  `json:"max_candle"`
		MinCandle          Candle  `json:"min_candle"`
		MostVolatileCandle Candle  `json:"most_volatile_candle"`
		MostVolumeCandle   Candle  `json:"most_volume_candle"`
		AvgBodySize        float64 `json:"avg_body_size"`
		AvgUpperWick       float64 `json:"avg_upper_wick"`
		AvgLowerWick       float64 `json:"avg_lower_wick"`
		MaxGapUp           float64 `json:"max_gap_up"`
		MaxGapDown         float64 `json:"max_gap_down"`
		MaxGapUpCandle     Candle  `json:"max_gap_up_candle"`
		MaxGapDownCandle   Candle  `json:"max_gap_down_candle"`
		BullishStreak      int     `json:"bullish_streak"`
		BearishStreak      int     `json:"bearish_streak"`
		PriceRangePercent  float64 `json:"price_range_percent"`
		DominantHour       int     `json:"dominant_hour"`
	} `json:"analytics"`
}

type Analyzer struct {
	log *logger.Logger
}

func NewAnalyzer(log *logger.Logger) *Analyzer {
	return &Analyzer{log: log.Named("analyzer")}
}

func (a *Analyzer) AnalyzeCandles(ctx context.Context, candles []Candle) (*AnalyticsResponse, error) {
	if len(candles) == 0 {
		return nil, errors.New("empty input")
	}
	if len(candles) > 5000 {
		return nil, fmt.Errorf("input exceeds limit: %d", len(candles))
	}
	for i, c := range candles {
		if c.OpenTime.IsZero() || c.CloseTime.IsZero() || c.Symbol == "" {
			return nil, fmt.Errorf("candle[%d]: missing required fields", i)
		}
	}
	sort.Slice(candles, func(i, j int) bool {
		return candles[i].OpenTime.Before(candles[j].OpenTime)
	})

	symbol := candles[0].Symbol
	start := candles[0].OpenTime
	end := candles[len(candles)-1].CloseTime
	count := len(candles)

	var sumClose, sumVolume float64
	var sumBody, sumWickTop, sumWickBot float64
	var upCount, downCount int
	maxHigh, minLow := candles[0].High, candles[0].Low
	maxCandle, minCandle := candles[0], candles[0]
	mostVolatile, mostVolume := candles[0], candles[0]
	maxVolatility := candles[0].High - candles[0].Low
	maxVol := candles[0].Volume
	maxGapUp, maxGapDown := 0.0, 0.0
	var maxGapUpCandle, maxGapDownCandle Candle
	hourVolume := make(map[int]float64)
	bullStreak, maxBullStreak := 0, 0
	bearStreak, maxBearStreak := 0, 0

	for i, c := range candles {
		close := c.Close
		open := c.Open
		high := c.High
		low := c.Low
		volume := c.Volume

		body := abs(close - open)
		upper := high - max(close, open)
		lower := min(close, open) - low

		sumClose += close
		sumVolume += volume
		sumBody += body
		sumWickTop += upper
		sumWickBot += lower

		hourVolume[c.OpenTime.Hour()] += volume

		if close > open {
			upCount++
			bullStreak++
			bearStreak = 0
			if bullStreak > maxBullStreak {
				maxBullStreak = bullStreak
			}
		} else if close < open {
			downCount++
			bearStreak++
			bullStreak = 0
			if bearStreak > maxBearStreak {
				maxBearStreak = bearStreak
			}
		}

		if high > maxHigh {
			maxHigh = high
			maxCandle = c
		}
		if low < minLow {
			minLow = low
			minCandle = c
		}
		if v := high - low; v > maxVolatility {
			maxVolatility = v
			mostVolatile = c
		}
		if volume > maxVol {
			maxVol = volume
			mostVolume = c
		}

		// Gap
		if i > 0 {
			prevClose := candles[i-1].Close
			gap := c.Open - prevClose
			if gap > 0 && gap > maxGapUp {
				maxGapUp = gap
				maxGapUpCandle = c
			}
			if gap < 0 && -gap > maxGapDown {
				maxGapDown = -gap
				maxGapDownCandle = c
			}
		}

	}

	avgClose := sumClose / float64(count)
	upRatio := float64(upCount) / float64(count)
	downRatio := float64(downCount) / float64(count)
	priceChange := candles[count-1].Close - candles[0].Open
	volatility := maxHigh - minLow
	priceRangePercent := (volatility / candles[0].Open) * 100

	dominantHour, maxVolHour := 0, 0.0
	for h, v := range hourVolume {
		if v > maxVolHour {
			dominantHour = h
			maxVolHour = v
		}
	}

	resp := &AnalyticsResponse{
		Symbol:   symbol,
		Interval: inferInterval(candles),
		Start:    start,
		End:      end,
		Count:    count,
	}
	resp.Analytics.AvgClose = avgClose
	resp.Analytics.SumVolume = sumVolume
	resp.Analytics.PriceChange = priceChange
	resp.Analytics.Volatility = volatility
	resp.Analytics.UpCount = upCount
	resp.Analytics.DownCount = downCount
	resp.Analytics.UpRatio = upRatio
	resp.Analytics.DownRatio = downRatio
	resp.Analytics.MaxCandle = maxCandle
	resp.Analytics.MinCandle = minCandle
	resp.Analytics.MostVolatileCandle = mostVolatile
	resp.Analytics.MostVolumeCandle = mostVolume
	resp.Analytics.AvgBodySize = sumBody / float64(count)
	resp.Analytics.AvgUpperWick = sumWickTop / float64(count)
	resp.Analytics.AvgLowerWick = sumWickBot / float64(count)
	resp.Analytics.MaxGapUp = maxGapUp
	resp.Analytics.MaxGapDown = maxGapDown
	resp.Analytics.MaxGapUpCandle = maxGapUpCandle
	resp.Analytics.MaxGapDownCandle = maxGapDownCandle
	resp.Analytics.BullishStreak = maxBullStreak
	resp.Analytics.BearishStreak = maxBearStreak
	resp.Analytics.PriceRangePercent = priceRangePercent
	resp.Analytics.DominantHour = dominantHour

	return resp, nil
}

func inferInterval(candles []Candle) string {
	if len(candles) < 2 {
		return "unknown"
	}
	delta := candles[1].OpenTime.Sub(candles[0].OpenTime)
	switch delta {
	case time.Minute:
		return "1m"
	case 5 * time.Minute:
		return "5m"
	case 15 * time.Minute:
		return "15m"
	case time.Hour:
		return "1h"
	case 4 * time.Hour:
		return "4h"
	case 24 * time.Hour:
		return "1d"
	default:
		return delta.String()
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
