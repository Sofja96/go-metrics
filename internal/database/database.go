package database

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	//"github.com/jackc/pgx/v5/pgxpool"
	//"sync"
)

//type postgres struct {
//	db *pgxpool.Pool
//}
//
//var (
//	pgInstance *postgres
//	pgOnce     sync.Once
//)
//
//func NewPG(ctx context.Context, connString string) (*postgres, error) {
//	pgOnce.Do(func() {
//		db, err := pgxpool.New(ctx, connString)
//		if err != nil {
//			return fmt.Errorf("unable to create connection pool: %w", err)
//		}
//
//		pgInstance = &postgres{db}
//	})
//
//	return pgInstance, nil
//}
//
//func (pg *postgres) Ping(ctx context.Context) error {
//	return pg.db.Ping(ctx)
//}
//
//func (pg *postgres) Close() {
//	pg.db.Close()
//}

// urlExample := "postgres://metrics:userpassword@localhost:5432/metrics"
//
//	type postgres struct {
//		db *pgxpool.Pool
//	}
type Dbinstance struct {
	DSN string
}

func DBConnection(dsn string) *Dbinstance {
	db := &Dbinstance{}
	db.DSN = dsn

	return db
}

func CheckConnection(db *Dbinstance) error {
	if db.DSN == "" {
		return errors.New("Empty connection string")
	}
	dbc, err := pgxpool.New(context.Background(), db.DSN)
	//dbc, err := pgx.Open("pgx", db.DSN)
	if err != nil {
		return err
	}
	defer dbc.Close()
	return nil
}

//
//func NewPG(ctx context.Context, connString string) (*postgres, error) {
//	pgOnce.Do(func() {
//		db, err := pgxpool.New(ctx, connString)
//		if err != nil {
//			return fmt.Errorf("unable to create connection pool: %w", err)
//		}
//
//		pgInstance = &postgres{db}
//	})
//
//	return pgInstance, nil
//}
//
//func (pg *postgres) Ping(ctx context.Context) error {
//	return pg.db.Ping(ctx)
//}
//
//func (pg *postgres) Close() {
//	pg.db.Close()
//}
