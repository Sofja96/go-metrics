package database

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
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

type Postgres struct {
	db *pgxpool.Pool
}

func New(dsn string) *Postgres {
	dbc := &Postgres{}

	conn, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Println(err)
		dbc.db = nil
	} else {
		dbc.db = conn
	}
	return dbc
}

func CheckConnection(dbc *Postgres) error {
	if dbc.db != nil {
		err := dbc.db.Ping(context.Background())
		if err != nil {
			log.Println("Unable to connect to database:", err)
			//os.Exit(1)
			return err
		}
		return nil
	}

	return errors.New("Empty connection string")
}

//type Dbinstance struct {
//	DSN string
//}
//
//func DBConnection(dsn string) *Dbinstance {
//	db := &Dbinstance{}
//	db.DSN = dsn
//
//	return db
//}
//
//func CheckConnection(db *Dbinstance) error {
//	if db.DSN == "" {
//		err := errors.New("Empty connection string")
//		if err != nil {
//			return err
//		}
//	}
//	//err := dbc.db.Ping()
//	//if err != nil {
//	//	return err
//	//}
//	dbc, err := pgxpool.New(context.Background(), db.DSN)
//	//dbc, err := pgx.Open("pgx", db.DSN)
//	if err != nil {
//		fmt.Fprintln(os.Stderr, "Unable to connect to database:", err)
//		os.Exit(1)
//		//return err
//	}
//	defer dbc.Close()
//	return nil
//}
//
//func (pg *Postgres) Ping(ctx context.Context) error {
//	return pg.db.Ping(ctx)
//}
//
////
////func NewPG(ctx context.Context, connString string) (*postgres, error) {
////	pgOnce.Do(func() {
////		db, err := pgxpool.New(ctx, connString)
////		if err != nil {
////			return fmt.Errorf("unable to create connection pool: %w", err)
////		}
////
////		pgInstance = &postgres{db}
////	})
////
////	return pgInstance, nil
////}
////
////func (pg *postgres) Ping(ctx context.Context) error {
////	return pg.db.Ping(ctx)
////}
////
////func (pg *postgres) Close() {
////	pg.db.Close()
////}
