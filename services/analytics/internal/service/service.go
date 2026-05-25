package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

type ChartPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
	X     string  `json:"x"`
	Y     float64 `json:"y"`
}

type ChartSeries struct {
	Label  string       `json:"label"`
	Points []ChartPoint `json:"points"`
}

type ChartResponse struct {
	ChartType string        `json:"chartType"`
	Title     string        `json:"title"`
	Unit      *string       `json:"unit"`
	Series    []ChartSeries `json:"series"`
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetChartTypes(tagID uint) ([]string, error) {
	rows := []struct {
		ChartType string
	}{}
	if err := s.db.Table("tag_chart_type_links").Select("chart_type").Where("tag_id = ?", tagID).Order("chart_type ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.ChartType)
	}
	return result, nil
}

func (s *Service) GetChart(userID, tagID uint, chartType, dateFrom, dateTo string) (*ChartResponse, error) {
	query := s.db.Table("diary_entry_metric_values AS mv").
		Joins("JOIN diary_entry_metrics m ON m.id = mv.diary_entry_metric_id").
		Joins("JOIN diary_entries e ON e.id = m.diary_entry_id").
		Joins("JOIN diary_entry_tags det ON det.diary_entry_id = e.id").
		Joins("JOIN tags t ON t.id = det.tag_id").
		Joins("JOIN dictionary_items metric_name ON metric_name.id = m.metric_type_id").
		Where("e.user_id = ? AND e.status <> ? AND t.id = ?", userID, "DELETED", tagID)
	if strings.TrimSpace(dateFrom) != "" {
		query = query.Where("COALESCE(e.when_started, e.created_at) >= ?", dateFrom)
	}
	if strings.TrimSpace(dateTo) != "" {
		query = query.Where("COALESCE(e.when_started, e.created_at) <= ?", dateTo)
	}
	rows := []struct {
		MetricLabel string
		Value       float64
		OccurredAt  time.Time
	}{}
	if err := query.Select("metric_name.label AS metric_label, mv.value, COALESCE(e.when_started, e.created_at) AS occurred_at").
		Order("occurred_at ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	byMetric := map[string][]ChartPoint{}
	for _, row := range rows {
		dateLabel := row.OccurredAt.UTC().Format("2006-01-02")
		byMetric[row.MetricLabel] = append(byMetric[row.MetricLabel], ChartPoint{
			Label: dateLabel,
			Value: row.Value,
			X:     dateLabel,
			Y:     row.Value,
		})
	}

	metricNames := make([]string, 0, len(byMetric))
	for label := range byMetric {
		metricNames = append(metricNames, label)
	}
	sort.Strings(metricNames)

	series := make([]ChartSeries, 0, len(metricNames))
	for _, label := range metricNames {
		series = append(series, ChartSeries{Label: label, Points: byMetric[label]})
	}
	return &ChartResponse{
		ChartType: chartType,
		Title:     humanizeChartType(chartType),
		Unit:      nil,
		Series:    series,
	}, nil
}

func humanizeChartType(chartType string) string {
	return strings.Title(strings.ToLower(strings.ReplaceAll(chartType, "_", " ")))
}

func (s *Service) EnsureAllowedChartType(tagID uint, chartType string) error {
	var count int64
	if err := s.db.Table("tag_chart_type_links").Where("tag_id = ? AND chart_type = ?", tagID, chartType).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("chart type %s is not linked to tag", chartType)
	}
	return nil
}
