import {
  CandlestickSeries,
  createChart,
  LineSeries,
  type IChartApi,
  type UTCTimestamp
} from "lightweight-charts";
import React, { useEffect, useRef } from "react";
import { createRoot } from "react-dom/client";

const bars = [
  { time: 1783344600 as UTCTimestamp, open: 100, high: 105, low: 99, close: 103 },
  { time: 1783344660 as UTCTimestamp, open: 103, high: 106, low: 101, close: 102 },
  { time: 1783344720 as UTCTimestamp, open: 102, high: 108, low: 102, close: 107 },
  { time: 1783344780 as UTCTimestamp, open: 107, high: 110, low: 106, close: 109 }
];

const rsi = [
  { time: 1783344600 as UTCTimestamp, value: 48 },
  { time: 1783344660 as UTCTimestamp, value: 55 },
  { time: 1783344720 as UTCTimestamp, value: 63 },
  { time: 1783344780 as UTCTimestamp, value: 69 }
];

function App() {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const chartRef = useRef<IChartApi | null>(null);

  useEffect(() => {
    if (!containerRef.current) {
      return;
    }

    const chart = createChart(containerRef.current, {
      width: 720,
      height: 420,
      layout: {
        background: { color: "#ffffff" },
        textColor: "#1f2937"
      },
      rightPriceScale: {
        borderVisible: false
      },
      timeScale: {
        borderVisible: false
      }
    });
    chartRef.current = chart;

    const candleSeries = chart.addSeries(CandlestickSeries, {
      upColor: "#16a34a",
      downColor: "#dc2626",
      borderVisible: false,
      wickUpColor: "#16a34a",
      wickDownColor: "#dc2626"
    });
    candleSeries.setData(bars);

    const rsiSeries = chart.addSeries(
      LineSeries,
      {
        color: "#2563eb",
        lineWidth: 2,
        title: "RSI"
      },
      1
    );
    rsiSeries.setData(rsi);

    chart.timeScale().fitContent();

    return () => {
      chart.remove();
      chartRef.current = null;
    };
  }, []);

  return (
    <main style={{ padding: 24, fontFamily: "system-ui, sans-serif" }}>
      <h1>TradingView Lite Smoke</h1>
      <div ref={containerRef} style={{ width: 720, height: 420 }} />
    </main>
  );
}

createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
