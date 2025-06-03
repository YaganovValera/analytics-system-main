// query-service/internal/usecase/orderbook_analysis.go
package usecase

import (
	marketdatapb "github.com/YaganovValera/analytics-system/proto/gen/go/v1/marketdata"
)

// AnalyzeOrderBookSnapshots проводит аналитический разбор массива снапшотов стаканов.
func AnalyzeOrderBookSnapshots(snapshots []*marketdatapb.OrderBookSnapshot) *marketdatapb.OrderBookAnalysis {
	if len(snapshots) == 0 {
		return nil
	}

	var (
		spreadSum                    float64
		bidDepthSum                  float64
		askDepthSum                  float64
		imbalanceStart, imbalanceEnd float64
		bidSlopeSum, askSlopeSum     float64
		maxBidVolume, maxAskVolume   float64
		maxBidPrice, maxAskPrice     float64
		totalSnapshots               = float64(len(snapshots))
	)

	for idx, snap := range snapshots {
		bids := snap.Bids
		asks := snap.Asks

		if len(bids) == 0 || len(asks) == 0 {
			continue
		}

		// 1. Spread
		bestBid := bids[0].Price
		bestAsk := asks[0].Price
		mid := (bestBid + bestAsk) / 2
		spread := (bestAsk - bestBid) / mid * 100
		spreadSum += spread

		// 2. Top 10 depth
		bidDepth, askDepth := sumVolumes(bids, 10), sumVolumes(asks, 10)
		bidDepthSum += bidDepth
		askDepthSum += askDepth

		// 3. Imbalance
		totalBid := sumVolumes(bids, 20)
		totalAsk := sumVolumes(asks, 20)
		total := totalBid + totalAsk
		imbalance := 0.0
		if total > 0 {
			imbalance = (totalBid - totalAsk) / total
		}
		if idx == 0 {
			imbalanceStart = imbalance
		}
		if idx == len(snapshots)-1 {
			imbalanceEnd = imbalance
		}

		// 4. Max wall
		for _, b := range bids {
			if b.Quantity > maxBidVolume {
				maxBidVolume = b.Quantity
				maxBidPrice = b.Price
			}
		}
		for _, a := range asks {
			if a.Quantity > maxAskVolume {
				maxAskVolume = a.Quantity
				maxAskPrice = a.Price
			}
		}

		// 5. Slope (approx linear decay of volume by depth index)
		bidSlopeSum += slope(bids, 10)
		askSlopeSum += slope(asks, 10)
	}

	return &marketdatapb.OrderBookAnalysis{
		AvgSpreadPercent:  spreadSum / totalSnapshots,
		AvgBidVolumeTop10: bidDepthSum / totalSnapshots,
		AvgAskVolumeTop10: askDepthSum / totalSnapshots,
		ImbalanceStart:    imbalanceStart,
		ImbalanceEnd:      imbalanceEnd,
		MaxBidWallPrice:   maxBidPrice,
		MaxBidWallVolume:  maxBidVolume,
		MaxAskWallPrice:   maxAskPrice,
		MaxAskWallVolume:  maxAskVolume,
		BidSlope:          bidSlopeSum / totalSnapshots,
		AskSlope:          askSlopeSum / totalSnapshots,
	}
}

func sumVolumes(levels []*marketdatapb.OrderBookLevel, limit int) float64 {
	vol := 0.0
	for i := 0; i < len(levels) && i < limit; i++ {
		vol += levels[i].Quantity
	}
	return vol
}

func slope(levels []*marketdatapb.OrderBookLevel, depth int) float64 {
	n := float64(min(len(levels), depth))
	if n < 2 {
		return 0
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i := 0; i < int(n); i++ {
		x := float64(i)
		y := levels[i].Quantity
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	numerator := n*sumXY - sumX*sumY
	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0
	}
	return numerator / denominator // чем ниже — тем резче спад
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
