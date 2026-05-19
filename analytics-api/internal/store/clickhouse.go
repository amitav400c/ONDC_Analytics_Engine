package store

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickHouse struct {
	conn driver.Conn
}

type FunnelStage struct {
	EventType    string  `json:"event_type"`
	Count        uint64  `json:"count"`
	UniqueOrders uint64  `json:"unique_orders"`
	Rate         float64 `json:"rate"` // Computed client-side or here
}

type CancellationRow struct {
	City  string `json:"city"`
	Count uint64 `json:"count"`
	Day   string `json:"day"`
}

type VolumePoint struct {
	Day   string `json:"day"`
	Count uint64 `json:"count"`
}

type RecentEvent struct {
	EventID   string  `json:"event_id"`
	EventType string  `json:"event_type"`
	City      string  `json:"city"`
	Timestamp string  `json:"timestamp"`
	OrderID   string  `json:"order_id"`
	BuyerHash string  `json:"buyer_hash"`
	Amount    float64 `json:"amount"`
}

func NewClickHouse(addr, database string) (*ClickHouse, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{Database: database},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout:  5 * time.Second,
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	})
	if err != nil {
		return nil, fmt.Errorf("clickhouse open: %w", err)
	}
	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("clickhouse ping: %w", err)
	}
	return &ClickHouse{conn: conn}, nil
}

func (ch *ClickHouse) Funnel(ctx context.Context, from, to string) ([]FunnelStage, error) {
	query := `
		SELECT event_type, count() AS cnt, uniqExact(order_id) AS uniq
		FROM ondc_events
		WHERE timestamp >= toDateTime64(?, 3) AND timestamp <= toDateTime64(?, 3)
		GROUP BY event_type
		ORDER BY cnt DESC`

	if from == "" {
		from = time.Now().AddDate(0, 0, -30).Format("2006-01-02 15:04:05")
	}
	if to == "" {
		to = time.Now().Format("2006-01-02 15:04:05")
	}

	rows, err := ch.conn.Query(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stages []FunnelStage
	for rows.Next() {
		var s FunnelStage
		if err := rows.Scan(&s.EventType, &s.Count, &s.UniqueOrders); err != nil {
			return nil, err
		}
		stages = append(stages, s)
	}

	// Compute conversion rates relative to the first stage
	if len(stages) > 0 {
		base := float64(stages[0].Count)
		for i := range stages {
			if base > 0 {
				stages[i].Rate = float64(stages[i].Count) / base * 100
			}
		}
	}
	return stages, nil
}

func (ch *ClickHouse) Cancellations(ctx context.Context, city, from, to string) ([]CancellationRow, error) {
	query := `
		SELECT city, count() AS cnt, toString(toStartOfDay(timestamp)) AS day
		FROM ondc_events
		WHERE event_type = 'on_cancel'
		  AND timestamp >= toDateTime64(?, 3) AND timestamp <= toDateTime64(?, 3)`

	if city != "" {
		query += ` AND city = ?`
	}
	query += ` GROUP BY city, day ORDER BY cnt DESC LIMIT 100`

	if from == "" {
		from = time.Now().AddDate(0, 0, -30).Format("2006-01-02 15:04:05")
	}
	if to == "" {
		to = time.Now().Format("2006-01-02 15:04:05")
	}

	var rows driver.Rows
	var err error
	if city != "" {
		rows, err = ch.conn.Query(ctx, query, from, to, city)
	} else {
		rows, err = ch.conn.Query(ctx, query, from, to)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []CancellationRow
	for rows.Next() {
		var r CancellationRow
		if err := rows.Scan(&r.City, &r.Count, &r.Day); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func (ch *ClickHouse) Volume(ctx context.Context, days int) ([]VolumePoint, error) {
	if days <= 0 {
		days = 7
	}
	query := `
		SELECT toString(toStartOfDay(timestamp)) AS day, count() AS cnt
		FROM ondc_events
		WHERE timestamp >= now() - INTERVAL ? DAY
		GROUP BY day
		ORDER BY day`

	rows, err := ch.conn.Query(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []VolumePoint
	for rows.Next() {
		var v VolumePoint
		if err := rows.Scan(&v.Day, &v.Count); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

func (ch *ClickHouse) RecentEvents(ctx context.Context, limit int) ([]RecentEvent, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query := `
		SELECT event_id, event_type, city, toString(timestamp), order_id, buyer_hash, amount
		FROM ondc_events
		ORDER BY timestamp DESC
		LIMIT ?`

	rows, err := ch.conn.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []RecentEvent
	for rows.Next() {
		var e RecentEvent
		if err := rows.Scan(&e.EventID, &e.EventType, &e.City, &e.Timestamp, &e.OrderID, &e.BuyerHash, &e.Amount); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (ch *ClickHouse) Ping(ctx context.Context) error {
	return ch.conn.Ping(ctx)
}
