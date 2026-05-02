package yahoo

import (
	"math"
)

type TAResult struct {
	RSI           float64
	MACD          float64
	MACDSignal    float64
	MACDHistogram float64
	MA20          float64
	MA50          float64
	ATR           float64
}

func ComputeTA(h *YahooHistoryData) TAResult {
	closes := h.Closes
	highs := h.Highs
	lows := h.Lows

	ta := TAResult{
		RSI:  calcRSI(closes, 14),
		MA20: calcSMA(closes, 20),
		MA50: calcSMA(closes, 50),
		ATR:  calcATR(highs, lows, closes, 14),
	}

	ta.MACD, ta.MACDSignal, ta.MACDHistogram = calcMACD(closes, 12, 26, 9)
	return ta
}

func ClassifySignal(price float64, ta TAResult) string {
	bullish := 0
	bearish := 0

	if ta.RSI < 45 {
		bullish++
	} else if ta.RSI > 60 {
		bearish++
	}

	if ta.MACD > ta.MACDSignal {
		bullish++
	} else {
		bearish++
	}

	if ta.MA20 > 0 && price > ta.MA20 {
		bullish++
	} else if ta.MA20 > 0 {
		bearish++
	}

	if bullish >= 2 {
		return "bullish"
	}
	if bearish >= 2 {
		return "bearish"
	}
	return "neutral"
}

func calcSMA(closes []float64, period int) float64 {
	if len(closes) < period {
		period = len(closes)
	}
	if period == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range closes[len(closes)-period:] {
		sum += v
	}
	return sum / float64(period)
}

func calcEMA(closes []float64, period int) []float64 {
	if len(closes) == 0 || period == 0 {
		return nil
	}
	k := 2.0 / float64(period+1)
	ema := make([]float64, len(closes))
	ema[0] = closes[0]
	for i := 1; i < len(closes); i++ {
		ema[i] = closes[i]*k + ema[i-1]*(1-k)
	}
	return ema
}

func calcRSI(closes []float64, period int) float64 {
	if len(closes) < period+1 {
		return 50
	}
	start := len(closes) - period - 1
	gains, losses := 0.0, 0.0
	for i := start + 1; i <= start+period; i++ {
		delta := closes[i] - closes[i-1]
		if delta > 0 {
			gains += delta
		} else {
			losses -= delta
		}
	}
	if losses == 0 {
		return 100
	}
	rs := (gains / float64(period)) / (losses / float64(period))
	return math.Round((100-(100/(1+rs)))*100) / 100
}

func calcMACD(closes []float64, fast, slow, signal int) (macdLine, signalLine, histogram float64) {
	if len(closes) < slow {
		return 0, 0, 0
	}
	emaFast := calcEMA(closes, fast)
	emaSlow := calcEMA(closes, slow)

	macdSeries := make([]float64, len(closes))
	for i := range closes {
		macdSeries[i] = emaFast[i] - emaSlow[i]
	}

	signalSeries := calcEMA(macdSeries, signal)
	last := len(closes) - 1
	macdLine = math.Round(macdSeries[last]*100) / 100
	signalLine = math.Round(signalSeries[last]*100) / 100
	histogram = math.Round((macdLine-signalLine)*100) / 100
	return
}

func calcATR(highs, lows, closes []float64, period int) float64 {
	n := len(closes)
	if n < 2 || len(highs) < n || len(lows) < n {
		return 0
	}
	trueRanges := make([]float64, n-1)
	for i := 1; i < n; i++ {
		hl := highs[i] - lows[i]
		hc := math.Abs(highs[i] - closes[i-1])
		lc := math.Abs(lows[i] - closes[i-1])
		trueRanges[i-1] = math.Max(hl, math.Max(hc, lc))
	}
	return math.Round(calcSMA(trueRanges, period)*100) / 100
}
