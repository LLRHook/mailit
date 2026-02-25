package dto

import "time"

// MetricsDataPoint represents a single point in the time-series.
type MetricsDataPoint struct {
	Date       string `json:"date"`
	Sent       int    `json:"sent"`
	Delivered  int    `json:"delivered"`
	Bounced    int    `json:"bounced"`
	Failed     int    `json:"failed"`
	Opened     int    `json:"opened"`
	Clicked    int    `json:"clicked"`
	Complained int    `json:"complained"`
}

// MetricsTotals holds aggregate totals for the period.
type MetricsTotals struct {
	Sent         int     `json:"sent"`
	Delivered    int     `json:"delivered"`
	Bounced      int     `json:"bounced"`
	Failed       int     `json:"failed"`
	Opened       int     `json:"opened"`
	Clicked      int     `json:"clicked"`
	Complained   int     `json:"complained"`
	DeliveryRate float64 `json:"delivery_rate"`
	OpenRate     float64 `json:"open_rate"`
	BounceRate   float64 `json:"bounce_rate"`
}

// MetricsResponse is the response for GET /metrics.
type MetricsResponse struct {
	Period string             `json:"period"`
	From   time.Time          `json:"from"`
	To     time.Time          `json:"to"`
	Totals MetricsTotals      `json:"totals"`
	Data   []MetricsDataPoint `json:"data"`
}
