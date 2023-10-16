package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/Sofja96/go-metrics.git/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"time"

	//	"github.com/Sofja96/go-metrics.git/store/storage/database"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	//"github.com/jackc/pgx/v5/pgxpool"
	//"sync"
)

//type counterMetric struct {
//	name  string
//	value int64
//}
//
//type gaugeMetric struct {
//	name  string
//	value float64
//}

var retry = 3
var delays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

type Postgres struct {
	DB *pgxpool.Pool
}

//func (pg *Postgres) AllMetrics() *memory.AllMetrics {
//	return nil
//}

func NewStorage(dsn string) (*Postgres, error) {
	dbc := &Postgres{}

	//if dsn == "" {
	//	dbc.DB = nil
	//	return dbc
	//}

	conn, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		dbc.DB = nil
		return nil, err
	} else {
		dbc.DB = conn
	}

	err = dbc.initDB(context.Background())
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("error init db: %w", err)
	}
	return dbc, nil
}

func (pg *Postgres) initDB(ctx context.Context) error {
	err := pg.DB.Ping(ctx)
	if err != nil {
		return fmt.Errorf("Unable to connect to database: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	_, err = pg.DB.Exec(ctx, "CREATE TABLE IF NOT EXISTS counter_metrics (name char(30) UNIQUE, value integer);")
	if err != nil {
		return fmt.Errorf("error occured on creating table gauge: %w", err)
	}

	_, err = pg.DB.Exec(ctx, "CREATE TABLE IF NOT EXISTS gauge_metrics (name char(30) UNIQUE, value double precision);")
	if err != nil {
		return fmt.Errorf("error occured on creating table counter: %w", err)
	}

	return nil
}

func (pg *Postgres) GetGaugeValue(id string) (float64, bool) {
	ctx := context.Background()
	//ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	raw := pg.DB.QueryRow(ctx, "SELECT value FROM gauge_metrics WHERE name = $1", id)
	var gm storage.GaugeMetric
	err := raw.Scan(&gm.Value)
	if err != nil {
		return 0, false
	}
	return gm.Value, true
}

func (pg *Postgres) UpdateGauge(name string, value float64) (float64, error) {
	ctx := context.Background()
	_, err := pg.DB.Exec(ctx, "INSERT INTO gauge_metrics(name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = $2", name, value)
	if err != nil {
		fmt.Errorf("error insert gauge: %w", err)
	}
	return value, nil
}
func (pg *Postgres) UpdateCounter(name string, value int64) (int64, error) {
	ctx := context.Background()
	var newValue int64
	raw := pg.DB.QueryRow(ctx, "INSERT INTO counter_metrics(name, value)VALUES ($1, $2)	ON CONFLICT(name)DO UPDATE SET value = counter_metrics.value + $2 RETURNING value", name, value)
	err := raw.Scan(&newValue)
	if err != nil {
		return 0, fmt.Errorf("error insert counter: %w", err)
	}
	return newValue, nil
}

func (pg *Postgres) GetCounterValue(id string) (int64, bool) {
	ctx := context.Background()
	//ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	raw := pg.DB.QueryRow(ctx, "SELECT value FROM counter_metrics WHERE name = $1", id)
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
	rowsGauge, err := pg.DB.Query(ctx, "SELECT name, value FROM gauge_metrics;")
	if err != nil {
		fmt.Errorf("error selecting all gauges: %w", err)
	}
	if err := rowsGauge.Err(); err != nil {
		fmt.Errorf("error selecting all gauges: %w", err)
	}
	defer rowsGauge.Close()

	for rowsGauge.Next() {
		var gm storage.GaugeMetric
		err = rowsGauge.Scan(&gm.Name, &gm.Value)
		if err != nil {
			fmt.Errorf("error scanning all gauges: %w", err)
		}
		pg.UpdateGauge(gm.Name, gm.Value)
		gauges = append(gauges, gm)
	}

	return gauges, nil
}

// GetAllCounters returns all counter metrics.
func (pg *Postgres) GetAllCounters() ([]storage.CounterMetric, error) {
	counters := make([]storage.CounterMetric, 0)
	//	var counters []storage.CounterMetric
	ctx := context.Background()
	rowsCounter, err := pg.DB.Query(ctx, "SELECT name, value FROM counter_metrics;")
	if err != nil {
		fmt.Errorf("error selecting all counter: %w", err)
	}
	if err := rowsCounter.Err(); err != nil {
		fmt.Errorf("error selecting all counter: %w", err)
	}
	defer rowsCounter.Close()

	for rowsCounter.Next() {
		var cm storage.CounterMetric
		err = rowsCounter.Scan(&cm.Name, &cm.Value)
		if err != nil {
			fmt.Errorf("error scanning all counter: %w", err)
		}
		pg.UpdateCounter(cm.Name, cm.Value)
		counters = append(counters, cm)
	}

	//
	//
	//err, _ := pg.DB.Query(ctx, "SELECT name, value FROM counter_metrics", &counters)
	//if err != nil {
	//	return nil, nil
	//}

	return counters, nil
}

func (pg *Postgres) batchUpdate(gauge []models.Metrics, counter []models.Metrics) error {
	ctx := context.Background()
	tx, err := pg.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error occured on creating tx on batchupdate: %w", err)
	}
	defer tx.Rollback(ctx)
	gaugeQuery := `INSERT INTO gauge_metrics(name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = $2`
	gaugebatch := &pgx.Batch{}
	for _, v := range gauge {
		gaugebatch.Queue(gaugeQuery, v.ID, v.Value)
	}

	err = pg.DB.SendBatch(ctx, gaugebatch).Close()
	if err != nil {
		return fmt.Errorf("error batch gauge update: %w", err)
	}

	counterQuery := `INSERT INTO counter_metrics(name, value)VALUES ($1, $2) ON CONFLICT(name)DO UPDATE SET value = counter_metrics.value + $2`
	counterbatch := &pgx.Batch{}
	for _, v := range counter {
		counterbatch.Queue(counterQuery, v.ID, v.Delta)
	}

	err = pg.DB.SendBatch(ctx, counterbatch).Close()
	if err != nil {
		return fmt.Errorf("error batch counter update: %w", err)
	}

	return tx.Commit(ctx)

	//
	//
	//
	//
	//encoder := json.NewEncoder(w)
	//results := make([]models.Metrics, len(metrics))
	//for _, v := range metrics {
	//	switch v.MType {
	//	case "gauge":
	//		pg.UpdateGauge(v.ID, *v.Value)
	//	case "counter":
	//		val, _ := pg.UpdateCounter(v.ID, *v.Delta)
	//		*v.Delta = val
	//	}
	//	results = append(results, v)
	//}
	//if err := encoder.Encode(results[0]); err != nil {
	//	return fmt.Errorf("error occured on encoding result of batchupdate :%w", err)
	//}
	//return tx.Commit(ctx)
	////encoder := json.NewEncoder(w)
}

func (pg *Postgres) BatchUpdate(metrics []models.Metrics) error {
	ctx := context.Background()
	var gauge []models.Metrics
	var counter []models.Metrics
	//tx, err := pg.DB.BeginTx(ctx, pgx.TxOptions{})
	//if err != nil {
	//	return fmt.Errorf("error occured on creating tx on batchupdate: %w", err)
	//}
	//defer tx.Rollback(ctx)
	//encoder := json.NewEncoder(w)
	//results := make([]models.Metrics, len(metrics))
	for _, v := range metrics {
		switch v.MType {
		case "gauge":
			gauge = append(gauge, v)
		//	pg.UpdateGauge(v.ID, *v.Value)
		case "counter":
			counter = append(counter, v)
			//val, _ := pg.UpdateCounter(v.ID, *v.Delta)
			//*v.Delta = val
		}
		//	results = append(results, v)
	}

	for i := 0; ; i++ {
		err := pg.batchUpdate(gauge, counter)
		if err == nil || i >= retry {
			return err
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			select {
			case <-time.After(delays[i]):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return err
		//if err := encoder.Encode(results[0]); err != nil {
		//	return fmt.Errorf("error occured on encoding result of batchupdate :%w", err)
	}
	//encoder := json.NewEncoder(w)
}

//func (pg *Postgres) AllMetrics() *memory.AllMetrics {
//	var counters []storage.CounterMetric
//	var gauges []storage.GaugeMetric
//	ctx := context.Background()
//	rows, err := pg.DB.Query(ctx, "SELECT name, value FROM gauge_metrics",&gauges)
//	if err != nil {
//		return nil
//	}
//	rows, err = pg.DB.Query(ctx, "SELECT name, value FROM counter_metrics", &counters)
//	if err != nil {
//		return nil
//	}
//
//	//err := r.db.Select(&metrics, `SELECT name, value FROM counters`)
//
//	return rows
//
//}

func (pg *Postgres) Ping() error {
	ctx := context.Background()
	return pg.DB.Ping(ctx)
}
