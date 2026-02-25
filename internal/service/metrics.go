package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
)

// MetricsService defines operations for querying email metrics.
type MetricsService interface {
	Get(ctx context.Context, teamID uuid.UUID, period string) (*dto.MetricsResponse, error)
	IncrementCounter(ctx context.Context, teamID uuid.UUID, eventType string) error
}

type metricsService struct {
	metricsRepo postgres.MetricsRepository
}

// NewMetricsService creates a new MetricsService.
func NewMetricsService(metricsRepo postgres.MetricsRepository) MetricsService {
	return &metricsService{
		metricsRepo: metricsRepo,
	}
}

func (s *metricsService) Get(ctx context.Context, teamID uuid.UUID, period string) (*dto.MetricsResponse, error) {
	now := time.Now().UTC()
	var from time.Time
	var periodType string
	var dateFormat string

	switch period {
	case "24h":
		from = now.Add(-24 * time.Hour).Truncate(time.Hour)
		periodType = model.PeriodTypeHourly
		dateFormat = "15:04"
	case "30d":
		from = now.AddDate(0, 0, -30).Truncate(24 * time.Hour)
		periodType = model.PeriodTypeDaily
		dateFormat = "Jan 2"
	default: // "7d"
		period = "7d"
		from = now.AddDate(0, 0, -7).Truncate(24 * time.Hour)
		periodType = model.PeriodTypeDaily
		dateFormat = "Jan 2"
	}

	timeSeries, err := s.metricsRepo.ListByTeam(ctx, teamID, periodType, from, now)
	if err != nil {
		return nil, fmt.Errorf("listing metrics: %w", err)
	}

	totals, err := s.metricsRepo.AggregateTotals(ctx, teamID, periodType, from, now)
	if err != nil {
		return nil, fmt.Errorf("aggregating metrics: %w", err)
	}

	data := make([]dto.MetricsDataPoint, len(timeSeries))
	for i, m := range timeSeries {
		data[i] = dto.MetricsDataPoint{
			Date:       m.PeriodStart.Format(dateFormat),
			Sent:       m.Sent,
			Delivered:  m.Delivered,
			Bounced:    m.Bounced,
			Failed:     m.Failed,
			Opened:     m.Opened,
			Clicked:    m.Clicked,
			Complained: m.Complained,
		}
	}

	resp := &dto.MetricsResponse{
		Period: period,
		From:   from,
		To:     now,
		Totals: dto.MetricsTotals{
			Sent:       totals.Sent,
			Delivered:  totals.Delivered,
			Bounced:    totals.Bounced,
			Failed:     totals.Failed,
			Opened:     totals.Opened,
			Clicked:    totals.Clicked,
			Complained: totals.Complained,
		},
		Data: data,
	}

	if totals.Sent > 0 {
		resp.Totals.DeliveryRate = float64(totals.Delivered) / float64(totals.Sent) * 100
		resp.Totals.OpenRate = float64(totals.Opened) / float64(totals.Sent) * 100
		resp.Totals.BounceRate = float64(totals.Bounced) / float64(totals.Sent) * 100
	}

	return resp, nil
}

func (s *metricsService) IncrementCounter(ctx context.Context, teamID uuid.UUID, eventType string) error {
	now := time.Now().UTC()

	// Increment both hourly and daily buckets.
	for _, pt := range []struct {
		periodType string
		start      time.Time
	}{
		{model.PeriodTypeHourly, now.Truncate(time.Hour)},
		{model.PeriodTypeDaily, now.Truncate(24 * time.Hour)},
	} {
		m := &model.EmailMetrics{
			TeamID:      teamID,
			PeriodStart: pt.start,
			PeriodType:  pt.periodType,
		}

		switch eventType {
		case model.EventSent:
			m.Sent = 1
		case model.EventDelivered:
			m.Delivered = 1
		case model.EventBounced:
			m.Bounced = 1
		case model.EventFailed:
			m.Failed = 1
		case model.EventOpened:
			m.Opened = 1
		case model.EventClicked:
			m.Clicked = 1
		case model.EventComplained:
			m.Complained = 1
		default:
			return nil // Unknown event type, skip
		}

		if err := s.metricsRepo.Upsert(ctx, m); err != nil {
			return fmt.Errorf("upserting metrics for %s/%s: %w", pt.periodType, eventType, err)
		}
	}

	return nil
}
