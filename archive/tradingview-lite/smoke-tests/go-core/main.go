package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cinar/indicator/v2/helper"
	"github.com/cinar/indicator/v2/trend"
	"github.com/gofiber/fiber/v2"
	yahoo "github.com/z-Wind/yahoofinance"
)

type result struct {
	Check  string `json:"check"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

func printResult(check, status, detail string) {
	payload, _ := json.Marshal(result{Check: check, Status: status, Detail: detail})
	fmt.Println(string(payload))
}

func main() {
	testFiber()
	testIndicator()
	testYahoo()
}

func testFiber() {
	app := fiber.New()
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res, err := app.Test(req)
	if err != nil {
		printResult("fiber", "fail", err.Error())
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		printResult("fiber", "fail", fmt.Sprintf("status=%d", res.StatusCode))
		return
	}

	printResult("fiber", "pass", "health route returned 200")
}

func testIndicator() {
	closes := []float64{1, 2, 3, 4, 5, 6}
	sma := trend.NewSmaWithPeriod[float64](3)
	values := helper.ChanToSlice(sma.Compute(helper.SliceToChan(closes)))
	if len(values) == 0 {
		printResult("indicator_sma", "fail", "empty SMA output")
		return
	}

	printResult("indicator_sma", "pass", fmt.Sprintf("count=%d last=%.4f", len(values), values[len(values)-1]))
}

func testYahoo() {
	client := &http.Client{Timeout: 15 * time.Second}
	service, err := yahoo.New(client)
	if err != nil {
		printResult("yahoo_client_init", "fail", err.Error())
		return
	}

	quote, err := service.Quote.RegularMarketPrice("AAPL").Do()
	if err != nil {
		printResult("yahoo_quote", "fail", err.Error())
		return
	}
	if len(quote.Chart.Result) == 0 {
		printResult("yahoo_quote", "fail", "empty chart result")
		return
	}
	meta := quote.Chart.Result[0].Meta
	printResult("yahoo_quote", "pass", fmt.Sprintf("symbol=%s price=%.4f granularity=%s", meta.Symbol, meta.RegularMarketPrice, meta.DataGranularity))

	history, err := service.History.Period("AAPL", "5d", "1d").Do()
	if err != nil {
		printResult("yahoo_history_1d", "fail", err.Error())
		return
	}
	if len(history.Chart.Result) == 0 {
		printResult("yahoo_history_1d", "fail", "empty chart result")
		return
	}
	first := history.Chart.Result[0]
	printResult("yahoo_history_1d", "pass", fmt.Sprintf("timestamps=%d quote_series=%d", len(first.Timestamp), len(first.Indicators.Quote)))

	minute, err := service.History.Period("AAPL", "1d", "1m").Do()
	if err != nil {
		printResult("yahoo_history_1m", "fail", err.Error())
		return
	}
	if len(minute.Chart.Result) == 0 || len(minute.Chart.Result[0].Indicators.Quote) == 0 {
		printResult("yahoo_history_1m", "fail", "empty chart result")
		return
	}
	minuteResult := minute.Chart.Result[0]
	quoteSeries := minuteResult.Indicators.Quote[0]
	lastClose := 0.0
	zeroOHLC := 0
	for i := range quoteSeries.Close {
		if quoteSeries.Open[i] == 0 || quoteSeries.High[i] == 0 || quoteSeries.Low[i] == 0 || quoteSeries.Close[i] == 0 {
			zeroOHLC++
		}
		lastClose = quoteSeries.Close[i]
	}
	printResult("yahoo_history_1m", "pass", fmt.Sprintf("timestamps=%d closes=%d last_close=%.4f zero_ohlc_rows=%d", len(minuteResult.Timestamp), len(quoteSeries.Close), lastClose, zeroOHLC))
}
