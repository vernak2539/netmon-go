package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vernak2539/netmon-go/internal/models"
	_ "modernc.org/sqlite"
)

type DB struct {
	db *sql.DB
}

// Open opens a connection to the SQLite database and initializes the schema.
func Open(path string) (*DB, error) {
	if path == "" {
		return nil, fmt.Errorf("database path cannot be empty")
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	d := &DB{db: db}
	if err := d.createSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return d, nil
}

func (d *DB) createSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS metrics (
			id             TEXT PRIMARY KEY,
			download       REAL NOT NULL,
			upload         REAL NOT NULL,
			ping           REAL NOT NULL,
			share          TEXT,
			client         TEXT NOT NULL,
			server         TEXT NOT NULL,
			bytes_sent     INTEGER NOT NULL,
			bytes_received INTEGER NOT NULL,
			timestamp      DATETIME NOT NULL DEFAULT (datetime('now'))
		);`,
		`CREATE TABLE IF NOT EXISTS device_scans (
			id             TEXT PRIMARY KEY,
			ips            TEXT NOT NULL,
			latencies      TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS speedtest (
			id              TEXT PRIMARY KEY,
			device_scans_id TEXT UNIQUE REFERENCES device_scans(id) ON DELETE CASCADE,
			metrics_id      TEXT UNIQUE REFERENCES metrics(id) ON DELETE CASCADE
		);`,
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, query := range queries {
		if _, err := tx.Exec(query); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) AddMetric(ctx context.Context, metric *models.NetworkMetric) error {
	query := `
		INSERT INTO metrics (
			id, download, upload, ping, timestamp, share, client, server,
			bytes_sent, bytes_received
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err := d.db.ExecContext(ctx, query,
		metric.ID.String(),
		metric.Download,
		metric.Upload,
		metric.Ping,
		metric.Timestamp.Format(time.RFC3339),
		metric.Share,
		metric.Client,
		metric.Server,
		metric.BytesSent,
		metric.BytesReceived,
	)
	return err
}

func (d *DB) AddDevices(ctx context.Context, devices []models.NetworkDevice) (uuid.UUID, error) {
	scanID, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("generating UUID: %w", err)
	}

	ips := make([]string, len(devices))
	latencies := make([]float64, len(devices))
	for i, dev := range devices {
		ips[i] = dev.IP
		latencies[i] = dev.LatencyMs
	}

	ipsJSON, err := json.Marshal(ips)
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshaling IPs: %w", err)
	}

	latenciesJSON, err := json.Marshal(latencies)
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshaling latencies: %w", err)
	}

	query := `
		INSERT INTO device_scans (id, ips, latencies)
		VALUES (?, ?, ?);
	`
	_, err = d.db.ExecContext(ctx, query, scanID.String(), string(ipsJSON), string(latenciesJSON))
	if err != nil {
		return uuid.Nil, err
	}

	return scanID, nil
}

func (d *DB) AddSpeedTest(ctx context.Context, speedtest *models.SpeedTest) error {
	query := `
		INSERT INTO speedtest (id, metrics_id, device_scans_id)
		VALUES (?, ?, ?);
	`
	_, err := d.db.ExecContext(ctx, query,
		speedtest.ID.String(),
		speedtest.MetricID.String(),
		speedtest.DeviceScanID.String(),
	)
	return err
}

func (d *DB) GetMetrics(ctx context.Context) ([]models.NetworkMetric, error) {
	query := `
		SELECT id, download, upload, ping, timestamp, share, client, server, bytes_sent, bytes_received
		FROM (
			SELECT id, download, upload, ping, timestamp, share, client, server, bytes_sent, bytes_received
			FROM metrics
			WHERE timestamp > datetime('now', '-24 hours')
			ORDER BY timestamp DESC
			LIMIT 24
		)
		ORDER BY timestamp ASC;
	`
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []models.NetworkMetric
	for rows.Next() {
		var m models.NetworkMetric
		var idStr, timeStr string
		err := rows.Scan(
			&idStr,
			&m.Download,
			&m.Upload,
			&m.Ping,
			&timeStr,
			&m.Share,
			&m.Client,
			&m.Server,
			&m.BytesSent,
			&m.BytesReceived,
		)
		if err != nil {
			return nil, err
		}

		m.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("parsing UUID %s: %w", idStr, err)
		}

		m.Timestamp, err = time.Parse(time.RFC3339, timeStr)
		if err != nil {
			// Fallback in case SQLite defaults filled it as standard datetime string
			m.Timestamp, err = time.Parse("2006-01-02 15:04:05", timeStr)
			if err != nil {
				return nil, fmt.Errorf("parsing timestamp %s: %w", timeStr, err)
			}
		}

		metrics = append(metrics, m)
	}

	return metrics, nil
}

func (d *DB) GetMetricsWithDeviceCounts(ctx context.Context) ([]models.NetworkMetric, []int, error) {
	query := `
		SELECT id, download, upload, ping, timestamp, share, client, server, bytes_sent, bytes_received, ips
		FROM (
			SELECT m.id, m.download, m.upload, m.ping, m.timestamp, m.share, m.client, m.server,
			       m.bytes_sent, m.bytes_received, ds.ips
			FROM metrics m
			JOIN speedtest st ON st.metrics_id = m.id
			JOIN device_scans ds ON ds.id = st.device_scans_id
			WHERE m.timestamp > datetime('now', '-24 hours')
			ORDER BY m.timestamp DESC
			LIMIT 24
		)
		ORDER BY timestamp ASC;
	`
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var metrics []models.NetworkMetric
	var deviceCounts []int

	for rows.Next() {
		var m models.NetworkMetric
		var idStr, timeStr, ipsJSON string
		err := rows.Scan(
			&idStr,
			&m.Download,
			&m.Upload,
			&m.Ping,
			&timeStr,
			&m.Share,
			&m.Client,
			&m.Server,
			&m.BytesSent,
			&m.BytesReceived,
			&ipsJSON,
		)
		if err != nil {
			return nil, nil, err
		}

		m.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing UUID %s: %w", idStr, err)
		}

		m.Timestamp, err = time.Parse(time.RFC3339, timeStr)
		if err != nil {
			m.Timestamp, err = time.Parse("2006-01-02 15:04:05", timeStr)
			if err != nil {
				return nil, nil, fmt.Errorf("parsing timestamp %s: %w", timeStr, err)
			}
		}

		var ips []string
		if err := json.Unmarshal([]byte(ipsJSON), &ips); err != nil {
			return nil, nil, fmt.Errorf("unmarshaling ips: %w", err)
		}

		metrics = append(metrics, m)
		deviceCounts = append(deviceCounts, len(ips))
	}

	return metrics, deviceCounts, nil
}
