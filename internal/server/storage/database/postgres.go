package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/Sofja96/go-metrics.git/internal/server/storage"
)

type Postgres struct {
	DB *sqlx.DB
}

// NewStorage - создает хранилище БД
func NewStorage(dsn string) (*Postgres, error) {
	dbc := &Postgres{}

	conn, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	dbc.DB = conn

	err = dbc.InitDB(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error init db: %w", err)
	}
	return dbc, nil
}

func (pg *Postgres) InitDB(ctx context.Context) error {
	err := pg.DB.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	_, err = pg.DB.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS counter_metrics (name char(30) UNIQUE, value bigint);")
	if err != nil {
		return fmt.Errorf("error occured on creating table gauge: %w", err)
	}

	_, err = pg.DB.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS gauge_metrics (name char(30) UNIQUE, value double precision);")
	if err != nil {
		return fmt.Errorf("error occured on creating table counter: %w", err)
	}

	return nil
}

func (pg *Postgres) GetGaugeValue(id string) (float64, bool) {
	ctx := context.Background()
	raw := pg.DB.QueryRowContext(ctx, "SELECT value FROM gauge_metrics WHERE name = $1", id)
	var gm storage.GaugeMetric
	err := raw.Scan(&gm.Value)
	if err != nil {
		return 0, false
	}
	return gm.Value, true
}

func (pg *Postgres) UpdateGauge(name string, value float64) (float64, error) {
	ctx := context.Background()
	_, err := pg.DB.ExecContext(ctx, "INSERT INTO gauge_metrics(name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = $2", name, value)
	if err != nil {
		return 0, fmt.Errorf("error insert gauge: %w", err)
	}
	return value, nil
}
func (pg *Postgres) UpdateCounter(name string, value int64) (int64, error) {
	ctx := context.Background()
	var newValue int64
	raw := pg.DB.QueryRowContext(ctx, "INSERT INTO counter_metrics(name, value)VALUES ($1, $2)	ON CONFLICT(name)DO UPDATE SET value = counter_metrics.value + $2 RETURNING value", name, value)
	err := raw.Scan(&newValue)
	if err != nil {
		return 0, fmt.Errorf("error insert counter: %w", err)
	}
	return newValue, nil
}

func (pg *Postgres) GetCounterValue(id string) (int64, bool) {
	ctx := context.Background()
	raw := pg.DB.QueryRowContext(ctx, "SELECT value FROM counter_metrics WHERE name = $1", id)
	var cm storage.CounterMetric
	err := raw.Scan(&cm.Value)
	if err != nil {
		return 0, false
	}
	return cm.Value, true
}

func (pg *Postgres) GetAllGauges() ([]storage.GaugeMetric, error) {
	ctx := context.Background()
	gauges := make([]storage.GaugeMetric, 0)
	rowsGauge, err := pg.DB.QueryContext(ctx, "SELECT name, value FROM gauge_metrics;")
	if err != nil {
		return nil, fmt.Errorf("error selecting all gauges: %w", err)
	}
	if err := rowsGauge.Err(); err != nil {
		return nil, fmt.Errorf("error selecting all gauges: %w", err)
	}
	defer rowsGauge.Close()

	for rowsGauge.Next() {
		var gm storage.GaugeMetric
		err = rowsGauge.Scan(&gm.Name, &gm.Value)
		if err != nil {
			return nil, fmt.Errorf("error scanning all gauges: %w", err)
		}
		_, err := pg.UpdateGauge(gm.Name, gm.Value)
		if err != nil {
			return nil, fmt.Errorf("error update gauge: %v", err)
		}
		gauges = append(gauges, gm)
	}

	return gauges, nil
}

func (pg *Postgres) GetAllCounters() ([]storage.CounterMetric, error) {
	counters := make([]storage.CounterMetric, 0)
	ctx := context.Background()
	rowsCounter, err := pg.DB.QueryContext(ctx, "SELECT name, value FROM counter_metrics;")
	if err != nil {
		return nil, fmt.Errorf("error selecting all counter: %w", err)
	}
	if err := rowsCounter.Err(); err != nil {
		return nil, fmt.Errorf("error selecting all counter: %w", err)
	}
	defer rowsCounter.Close()

	for rowsCounter.Next() {
		var cm storage.CounterMetric
		err = rowsCounter.Scan(&cm.Name, &cm.Value)
		if err != nil {
			return nil, fmt.Errorf("error scanning all counter: %w", err)
		}
		_, err := pg.UpdateCounter(cm.Name, cm.Value)
		if err != nil {
			return nil, fmt.Errorf("error update counter: %v", err)
		}
		counters = append(counters, cm)
	}
	return counters, nil
}

func (pg *Postgres) BatchUpdate(metrics []models.Metrics) error {
	ctx := context.Background()
	tx, err := pg.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("error occured on creating tx on batchupdate: %w", err)
		return fmt.Errorf("error occured on creating tx on batchupdate: %w", err)
	}
	defer tx.Rollback()

	if len(metrics) == 0 {
		log.Println("no metrics provided")
		return fmt.Errorf("no metrics provided")
	}
	for _, v := range metrics {
		switch v.MType {
		case "gauge":
			_, err = pg.UpdateGauge(v.ID, *v.Value)
			if err != nil {
				log.Printf("error update gauge: %v", err)
				return fmt.Errorf("error update gauge: %v", err)
			}
		case "counter":
			val, err := pg.UpdateCounter(v.ID, *v.Delta)
			if err != nil {
				log.Printf("error update counter: %v", err)
				return fmt.Errorf("error update counter: %v", err)
			}
			*v.Delta = val
		default:
			log.Printf("unsopperted metrics type: %s", v.MType)
			return fmt.Errorf("unsopperted metrics type: %s", v.MType)
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("failed to commit transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	return nil
}

func (pg *Postgres) Ping() error {
	err := pg.DB.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}
