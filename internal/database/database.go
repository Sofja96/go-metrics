package database

import (
	"context"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/storage"
	//	"github.com/Sofja96/go-metrics.git/internal/storage/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
	//"github.com/jackc/pgx/v5/pgxpool"
	//"sync"
)

type counterMetric struct {
	name  string
	value int64
}

type gaugeMetric struct {
	name  string
	value float64
}

type Postgres struct {
	DB *pgxpool.Pool
}

func NewStorage(dsn string) *Postgres {
	dbc := &Postgres{}

	if dsn == "" {
		dbc.DB = nil
		return dbc
	}

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
		return nil
	}
	return dbc
}

func CheckConnection(dbc *Postgres) error {
	if dbc.DB != nil {
		err := dbc.DB.Ping(context.Background())
		if err != nil {
			log.Println("Unable to connect to database:", err)
			//os.Exit(1)
			return err
		}
		return nil
	}
	return nil
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

//func (pg *Postgres) SaveGauge(name string, val float64) error {
//	return nil
//}
//
//func (pg *Postgres) UpdateCounter(ctx context.Context, name string, val int64) (int64, error) {
//	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
//	defer cancel()
//
//	err := pg.db.Ping(ctx)
//	if err != nil {
//		return 0, fmt.Errorf("error occured on db ping when updating counter: %w", err)
//	}
//
//	row := pg.db.QueryRow(ctx, "SELECT name FROM counter WHERE name = $1", name)
//	err = row.Scan()
//	switch err {
//	case sql.ErrNoRows:
//		_, err = pg.db.Exec(ctx, "INSERT INTO counter (name, value) VALUES ($1, $2)", name, val)
//		if err != nil {
//			return 0, fmt.Errorf("error occured in updating counter: %w", err)
//		}
//	case nil:
//		_, err = pg.db.Exec(ctx, "UPDATE counter SET value = $2 WHERE name = $1", name, val)
//		if err != nil {
//			return 0, fmt.Errorf("error occured in updating counter: %w", err)
//		}
//	default:
//		return 0, fmt.Errorf("error occured in updating counter: %w", err)
//	}
//
//	return 0, nil
//}
//
//func (pg *Postgres) GetGauger(name string) (float64, bool) {
//	return 0.0, true
//}
//
//func (pg *Postgres) GetCounter(name string) (int64, bool) {
//	return 0, true
//}
//
//func (pg *Postgres) GetAll() ([]byte, error) {
//
//	return nil, nil
//}

func SaveMetricsInStorage(s *storage.MemStorage, pg *Postgres) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	conn, err := pg.DB.Acquire(ctx)
	if err != nil {
		return
	}
	defer conn.Release()

	//err := pg.DB.Ping(ctx)
	//if err != nil {
	//	log.Println("Unable to connect to database:", err)
	//	return
	//}
	//if dbc.DB == nil {
	//	return
	//}
	//
	//ctx := context.Background()
	rowsCounter, err := conn.Query(ctx, "SELECT name, value FROM counter_metrics;")
	if err != nil {
		log.Println(err)
	}
	if err := rowsCounter.Err(); err != nil {
		log.Println(err)
	}
	defer rowsCounter.Close()

	for rowsCounter.Next() {
		var cm counterMetric
		err = rowsCounter.Scan(&cm.name, &cm.value)
		if err != nil {
			fmt.Println(err)
		}
		s.UpdateCounter(cm.name, cm.value)
	}

	rowsGauge, err := conn.Query(ctx, "SELECT name, value FROM gauge_metrics;")
	if err != nil {
		log.Println(err)
	}
	if err := rowsGauge.Err(); err != nil {
		log.Println(err)
	}
	defer rowsGauge.Close()

	for rowsGauge.Next() {
		var gm gaugeMetric
		err = rowsGauge.Scan(&gm.name, &gm.value)
		if err != nil {
			log.Println(err)
		}
		s.UpdateGauge(gm.name, gm.value)
	}
}

func Dump(s *storage.MemStorage, pg *Postgres, storeInterval int) {
	go func() {
		pollTicker := time.NewTicker(time.Duration(storeInterval) * time.Second)
		defer pollTicker.Stop()
		for range pollTicker.C {
			err := saveMetricsInBd(s, pg)
			if err != nil {
				return
			}
		}
	}()
}

func Truncate(pg *Postgres) error {
	ctx := context.Background()
	query := "TRUNCATE counter_metrics, gauge_metrics; "
	_, err := pg.DB.Exec(ctx, query)
	if err != nil {
		return nil
	}
	return err

}

//func (s *storage) SaveGauge(ctx context.Context, name string, val float64) error {
//	tx, err := s.BeginTx(ctx, nil)
//	if err != nil {
//		return fmt.Errorf("error occured on opening tx: %w", err)
//	}
//	defer tx.Rollback()
//
//	row := tx.QueryRowContext(ctx, "SELECT value FROM gauge WHERE name = $1", name)
//	err = row.Scan()
//	switch err {
//	case sql.ErrNoRows:
//		_, err = tx.ExecContext(ctx, "INSERT INTO gauge (name, value) VALUES ($1, $2)", name, val)
//		if err != nil {
//			return fmt.Errorf("error occured on inserting gauge: %w", err)
//		}
//	case nil:
//		_, err = tx.ExecContext(ctx, "UPDATE counter SET value = $2 WHERE name = $1", name, val)
//		if err != nil {
//			return fmt.Errorf("error occured on updating gauge: %w", err)
//		}
//	default:
//		return fmt.Errorf("error occured on selecting gauge: %w", err)
//	}
//
//	return tx.Commit()
//}

func saveMetricsInBd(s *storage.MemStorage, pg *Postgres) error {
	//var metrics storage.AllMetrics
	//metrics.Counter = s.CounterData
	//metrics.Gauge = s.GaugeData
	m := s.AllMetrics()
	//s.Lock()
	//defer s.Unlock()
	//var data storage.AllMetrics

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	conn, err := pg.DB.Acquire(ctx)
	if err != nil {
		return err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error occured on opening tx: %w", err)
	}
	defer conn.Release()
	defer tx.Rollback(ctx)

	for k, v := range m.Counter {

		_, err = tx.Exec(ctx, "INSERT INTO counter_metrics(name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = $2", k, v)
		if err != nil {
			log.Println(err)
		}

	}

	for k, v := range m.Gauge {

		_, err = tx.Exec(ctx, "INSERT INTO gauge_metrics(name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = $2", k, v)
		if err != nil {
			log.Println(err)
		}

	}

	return tx.Commit(ctx)
	//return nil
}

//func saveMetrics(s *storage.MemStorage, pg *Postgres) error {
//	m := s.AllMetrics()
//	ctx := context.Background()
//	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
//	defer cancel()
//
//	err := pg.DB.Ping(ctx)
//	if err != nil {
//		log.Println("Unable to connect to database:", err)
//		return err
//	}
//	//var query []string
//	query, _ := pg.DB.Query(ctx, "TRUNCATE counter_metrics, gauge_metrics")
//	for k, v := range m.Counter {
//		raw, _ := pg.DB.Query(ctx, "INSERT INTO counter_metrics(name, value) VALUES ($1, $2)", k, v)
//	//	query = append(query, raw)
//	}
//
//	for k, v := range m.Gauge {
//		query += fmt.Sprintf("INSERT INTO gauge_metrics (name, value) VALUES ('%s', %f); ", k, v)
//	}
//
//_, err := dbc.DB.Exec(context.Background(), query)
//if err != nil {
//	return err
//}
//return nil
//}

//func saveMetrics(s *storage.MemStorage, dbc *Postgres) error {
//	ctx := context.Background()
//	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
//	defer cancel()
//	var query string
//	var data storage.AllMetrics
//	//m := s.AllMetrics()
//	query = "TRUNCATE counter_metrics, gauge_metrics; "
//	if len(data.Counter) == 0 {
//		_, err := dbc.DB.Exec(ctx, "UPDATE gauge_metrics SET value = $2 WHERE name = $1")
//		if err != nil {
//			log.Println(err)
//		}
//	}
//	if len(data.Gauge) != 0 {
//		s.UpdateGaugeData(data.Gauge)
//	}
//
//	//for k, v := range m.Counter {
//	//
//	//	//query += append("INSERT INTO counter_metrics (name, value) VALUES ('%s', %d); ", k, v)
//	//	query += fmt.Sprintf("INSERT INTO counter_metrics (name, value) VALUES ('%s', %d); ", k, v)
//	//}
//	//
//	//for k, v := range m.Gauge {
//	//	query += fmt.Sprintf("INSERT INTO gauge_metrics (name, value) VALUES ('%s', %f); ", k, v)
//	//}
//
//	_, err := dbc.DB.Exec(context.Background(), query)
//	if err != nil {
//		return err
//	}
//	return nil
//}
