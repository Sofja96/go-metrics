package database

import (
	"context"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/Sofja96/go-metrics.git/internal/storage"
	"github.com/jackc/pgx/v5"
	"time"

	//	"github.com/Sofja96/go-metrics.git/store/storage/database"
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
		log.Println(err)
		dbc.DB = nil
	} else {
		//err := Truncate(dbc)
		//if err != nil {
		//	log.Println(err)
		//}
		dbc.DB = conn
	}

	err = dbc.initDB(context.Background())
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	return dbc, nil
}

func (pg *Postgres) initDB(ctx context.Context) error {
	err := pg.DB.Ping(ctx)
	if err != nil {
		log.Println("Unable to connect to database:", err)
		//os.Exit(1)
		return err
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
		log.Println(err)
	}
	return value, nil
}
func (pg *Postgres) UpdateCounter(name string, value int64) (int64, error) {
	ctx := context.Background()
	var newValue int64
	raw := pg.DB.QueryRow(ctx, "INSERT INTO counter_metrics(name, value)VALUES ($1, $2)	ON CONFLICT(name)DO UPDATE SET value = counter_metrics.value + $2 RETURNING value", name, value)
	err := raw.Scan(&newValue)
	if err != nil {
		return 0, err
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
		fmt.Println(err)
	}
	if err := rowsGauge.Err(); err != nil {
		fmt.Println(err)
	}
	defer rowsGauge.Close()

	for rowsGauge.Next() {
		var gm storage.GaugeMetric
		err = rowsGauge.Scan(&gm.Name, &gm.Value)
		if err != nil {
			fmt.Println(err)
		}
		pg.UpdateGauge(gm.Name, gm.Value)
		gauges = append(gauges, gm)
	}

	//gauges := make([]storage.GaugeMetric, 0)
	//var gauges []storage.GaugeMetric
	//var gauges []storage.GaugeMetric
	//ctx := context.Background()
	//_, err := pg.DB.Exec(ctx, "SELECT name, value FROM gauge_metrics", &gauges)
	//if err != nil {
	//	return nil, nil
	//}

	//gauges = append(gauges, storage.GaugeMetric{Name: name, Value: value})
	//err = raw.Scan(&gauges)
	//if err != nil {
	//	return nil, fmt.Errorf("error occured on scanning gauge: %w", err)
	//}
	//defer raw.Close()
	//
	//for raw.Next() {
	//	var cm storage.GaugeMetric
	//	err = raw.Scan(&cm.Name, &cm.Value)
	//	if err != nil {
	//		return nil, fmt.Errorf("error occured on scanning gauge: %w", err)
	//	}
	//	//m := shared.NewEmptyGaugeMetric()
	//	//if err := rows.Scan(&m.ID, m.Value); err != nil {
	//	//	return nil, fmt.Errorf("error occured on scanning gauge: %w", err)
	//	//}
	//	gauges = append(gauges, cm)
	//}

	return gauges, nil
}

// GetAllCounters returns all counter metrics.
func (pg *Postgres) GetAllCounters() ([]storage.CounterMetric, error) {
	counters := make([]storage.CounterMetric, 0)
	//	var counters []storage.CounterMetric
	ctx := context.Background()
	rowsCounter, err := pg.DB.Query(ctx, "SELECT name, value FROM counter_metrics;")
	if err != nil {
		fmt.Println(err)
	}
	if err := rowsCounter.Err(); err != nil {
		fmt.Println(err)
	}
	defer rowsCounter.Close()

	for rowsCounter.Next() {
		var cm storage.CounterMetric
		err = rowsCounter.Scan(&cm.Name, &cm.Value)
		if err != nil {
			fmt.Println(err)
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

func (pg *Postgres) BatchUpdate(metrics []models.Metrics) error {
	ctx := context.Background()
	tx, err := pg.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error occured on creating tx on batchupdate: %w", err)
	}
	defer tx.Rollback(ctx)
	results := make([]models.Metrics, len(metrics))
	for _, v := range metrics {
		switch v.MType {
		case "gauge":
			pg.UpdateGauge(v.ID, *v.Value)
		case "counter":
			val, _ := pg.UpdateCounter(v.ID, *v.Delta)
			*v.Delta = val
		}
		results = append(results, v)
	}
	return tx.Commit(ctx)
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
