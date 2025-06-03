// preprocessor/internal/aggregator/candle.go
package aggregator

import (
	"fmt"
	"time"

	"github.com/YaganovValera/analytics-system/common/interval"
)

// candleState — состояние in-progress бара.
type candleState struct {
	Candle    *Candle
	UpdatedAt time.Time
}

func (cs *candleState) update(ts time.Time, price, volume float64) {
	c := cs.Candle
	if price > c.High {
		c.High = price
	}
	if price < c.Low {
		c.Low = price
	}
	c.Close = price
	c.Volume += volume

	// Обновляем метку времени, только если новее
	cs.UpdatedAt = ts
}

func (cs *candleState) shouldFlush(now time.Time) bool {
	return !now.Before(cs.Candle.End)
}

// newCandleState создаёт состояние нового бара.
func newCandleState(symbol, iv string, ts time.Time, price, volume float64) (*candleState, error) {
	start, err := AlignToInterval(ts, iv)
	if err != nil {
		return nil, fmt.Errorf("align failed: %w", err)
	}
	dur, err := interval.Duration(interval.Interval(iv))
	if err != nil {
		return nil, fmt.Errorf("interval duration: %w", err)
	}
	end := start.Add(dur)

	return &candleState{
		Candle: &Candle{
			Symbol:   symbol,
			Interval: iv,
			Start:    start,
			End:      end,
			Open:     price,
			High:     price,
			Low:      price,
			Close:    price,
			Volume:   volume,
			Complete: false,
		},
		UpdatedAt: ts,
	}, nil
}
