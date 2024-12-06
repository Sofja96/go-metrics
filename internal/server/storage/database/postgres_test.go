package database

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/Sofja96/go-metrics.git/internal/server/storage"
	storagemock "github.com/Sofja96/go-metrics.git/internal/server/storage/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

type mocks struct {
	DB      *sqlx.DB // мок для базы данных
	storage *storagemock.MockStorage
	Mock    sqlmock.Sqlmock // интерфейс для работы с mock SQL
}

func TestUpdateGauge(t *testing.T) {
	db, mock, err := sqlmock.New() // db - *sqlmock.DB, mock - sqlmock.Sqlmock
	if err != nil {
		t.Fatalf("failed to open mock database: %v", err)
	}
	defer db.Close()

	type (
		args struct {
			name  string
			value float64
		}
		mockBehavior func(m *mocks, args args)
	)
	tests := []struct {
		name          string
		args          args
		mockBehavior  mockBehavior
		expectedValue float64
		wantErr       bool
	}{
		{
			name: "Successful update",
			args: args{
				name:  "gauge",
				value: 0.12,
			},
			mockBehavior: func(m *mocks, args args) {
				expectedExec := `INSERT INTO 
    							gauge_metrics(name, value) 
								VALUES ($1, $2) ON CONFLICT (name)
								 DO UPDATE SET value = $2
								`
				mock.ExpectExec(regexp.QuoteMeta(expectedExec)).WithArgs(args.name, args.value).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedValue: 0.12,
			wantErr:       false,
		},
		{
			name: "Error update",
			args: args{
				name:  "gauge",
				value: 0.12,
			},
			mockBehavior: func(m *mocks, args args) {
				expectedExec := `INSERT INTO 
    							gauge_metrics(name, value) 
								VALUES ($1, $2) ON CONFLICT (name)
								 DO UPDATE SET value = $2
								`
				mock.ExpectExec(regexp.QuoteMeta(expectedExec)).WithArgs(args.name, args.value).
					WillReturnError(fmt.Errorf("error insert gauge"))
			},
			expectedValue: 0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			m := &mocks{}

			m.DB = sqlx.NewDb(db, "sqlmock")
			//m.Mock = mock //todo пока не поняло нужно или нет
			pg := &Postgres{DB: m.DB}

			tt.mockBehavior(m, tt.args)
			returnedValue, err := pg.UpdateGauge(tt.args.name, tt.args.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, returnedValue, "The returned value does not match the expected value")
			}

		})

	}
}

func TestUpdateCounter(t *testing.T) {
	db, mock, err := sqlmock.New() // db - *sqlmock.DB, mock - sqlmock.Sqlmock
	if err != nil {
		t.Fatalf("failed to open mock database: %v", err)
	}
	defer db.Close()

	type (
		args struct {
			name  string
			value int64
		}
		mockBehavior func(m *mocks, args args)
	)
	tests := []struct {
		name          string
		args          args
		mockBehavior  mockBehavior
		expectedValue int64
		wantErr       bool
	}{
		{
			name: "Successful update",
			args: args{
				name:  "counter",
				value: 2,
			},
			mockBehavior: func(m *mocks, args args) {

				rows := sqlmock.NewRows([]string{"value"}).AddRow(args.value)
				expectedExec := `INSERT INTO counter_metrics(name, value)VALUES ($1, $2) 
								ON CONFLICT(name)DO UPDATE SET value = counter_metrics.value 
								    + $2 RETURNING value
								`
				mock.ExpectQuery(regexp.QuoteMeta(expectedExec)).WithArgs(args.name, args.value).
					WillReturnRows(rows)
			},
			expectedValue: 2,
			wantErr:       false,
		},
		{
			name: "Error update",
			args: args{
				name:  "counter",
				value: 2,
			},
			mockBehavior: func(m *mocks, args args) {
				expectedExec := `INSERT INTO counter_metrics(name, value)VALUES ($1, $2) 
								ON CONFLICT(name)DO UPDATE SET value = counter_metrics.value 
								    + $2 RETURNING value
								`
				mock.ExpectQuery(regexp.QuoteMeta(expectedExec)).WithArgs(args.name, args.value).
					WillReturnError(fmt.Errorf("error insert counter"))
			},
			expectedValue: 0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			m := &mocks{
				DB: sqlx.NewDb(db, "sqlmock"),
			}
			pg := &Postgres{DB: m.DB}

			tt.mockBehavior(m, tt.args)
			returnedValue, err := pg.UpdateCounter(tt.args.name, tt.args.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, returnedValue, "The returned value does not match the expected value")
			}

		})

	}
}

func TestGetGaugeValue(t *testing.T) {
	db, mock, err := sqlmock.New() // db - *sqlmock.DB, mock - sqlmock.Sqlmock
	if err != nil {
		t.Fatalf("failed to open mock database: %v", err)
	}
	defer db.Close()

	type (
		args struct {
			id string
		}
		mockBehavior func(m *mocks, args args)
	)
	tests := []struct {
		name          string
		args          args
		mockBehavior  mockBehavior
		exists        bool // ожидается ли сущестование метрики
		expectedValue float64
		wantErr       bool
		mockError     error // ошибка, возвращаемая моками
	}{
		{
			name: "Gauge exists",
			args: args{
				id: "Alloc",
			},
			mockBehavior: func(m *mocks, args args) {

				rows := sqlmock.NewRows([]string{"value"}).AddRow(75.5)
				expectedExec := `SELECT value FROM gauge_metrics WHERE name = $1
								`
				mock.ExpectQuery(regexp.QuoteMeta(expectedExec)).WithArgs(args.id).
					WillReturnRows(rows)
			},
			expectedValue: 75.5,
			wantErr:       false,
			mockError:     nil,
			exists:        true,
		},
		{
			name: "Gauge does not exists",
			args: args{
				id: "Alloc",
			},
			mockBehavior: func(m *mocks, args args) {
				expectedExec := `SELECT value FROM gauge_metrics WHERE name = $1
								`
				mock.ExpectQuery(regexp.QuoteMeta(expectedExec)).WithArgs(args.id).
					WillReturnError(sql.ErrNoRows)
			},
			expectedValue: 0,
			wantErr:       true,
			mockError:     sql.ErrNoRows,
			exists:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			m := &mocks{
				DB: sqlx.NewDb(db, "sqlmock"),
			}
			pg := &Postgres{DB: m.DB}

			tt.mockBehavior(m, tt.args)
			returnedValue, exists := pg.GetGaugeValue(tt.args.id)

			if tt.exists {
				assert.True(t, exists, "Gauge should exists")
				assert.Equal(t, tt.expectedValue, returnedValue, "The returned value does not match the expected value")
			} else {
				assert.False(t, exists, "Gauge should not exist")
				assert.Equal(t, float64(0), returnedValue, "The returned value should be 0 for non-existing gauges")
			}

			if tt.wantErr {
				assert.Error(t, tt.mockError)
			} else {
				assert.NoError(t, tt.mockError)
			}

		})

	}
}

func TestGetCounterValue(t *testing.T) {
	db, mock, err := sqlmock.New() // db - *sqlmock.DB, mock - sqlmock.Sqlmock
	if err != nil {
		t.Fatalf("failed to open mock database: %v", err)
	}
	defer db.Close()

	type (
		args struct {
			id string
		}
		mockBehavior func(m *mocks, args args)
	)
	tests := []struct {
		name          string
		args          args
		mockBehavior  mockBehavior
		exists        bool // ожидается ли сущестование метрики
		expectedValue int64
		wantErr       bool
		mockError     error // ошибка, возвращаемая моками
	}{
		{
			name: "Counter exists",
			args: args{
				id: "Counter",
			},
			mockBehavior: func(m *mocks, args args) {

				rows := sqlmock.NewRows([]string{"value"}).AddRow(2)
				expectedExec := `SELECT value FROM counter_metrics WHERE name = $1
								`
				mock.ExpectQuery(regexp.QuoteMeta(expectedExec)).WithArgs(args.id).
					WillReturnRows(rows)
			},
			expectedValue: 2,
			wantErr:       false,
			mockError:     nil,
			exists:        true,
		},
		{
			name: "Counter does not exists",
			args: args{
				id: "Counter",
			},
			mockBehavior: func(m *mocks, args args) {
				expectedExec := `SELECT value FROM counter_metrics WHERE name = $1
								`
				mock.ExpectQuery(regexp.QuoteMeta(expectedExec)).WithArgs(args.id).
					WillReturnError(sql.ErrNoRows)
			},
			expectedValue: 0,
			wantErr:       true,
			mockError:     sql.ErrNoRows,
			exists:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			m := &mocks{
				DB: sqlx.NewDb(db, "sqlmock"),
			}

			pg := &Postgres{DB: m.DB}

			tt.mockBehavior(m, tt.args)
			returnedValue, exists := pg.GetCounterValue(tt.args.id)

			if tt.exists {
				assert.True(t, exists, "Gauge should exists")
				assert.Equal(t, tt.expectedValue, returnedValue, "The returned value does not match the expected value")
			} else {
				assert.False(t, exists, "Gauge should not exist")
				assert.Equal(t, int64(0), returnedValue, "The returned value should be 0 for non-existing gauges")
			}

			if tt.wantErr {
				assert.Error(t, tt.mockError)
			} else {
				assert.NoError(t, tt.mockError)
			}

		})

	}
}

func TestGetAllGauges(t *testing.T) {
	db, mock, err := sqlmock.New() // db - *sqlmock.DB, mock - sqlmock.Sqlmock
	if err != nil {
		t.Fatalf("failed to open mock database: %v", err)
	}
	defer db.Close()

	type (
		args struct {
			name1  string
			value1 float64
			name2  string
			value2 float64
		}
		mockBehavior func(m *mocks, args args)
	)
	tests := []struct {
		name           string
		args           args
		expectedGauges []storage.GaugeMetric
		mockBehavior   mockBehavior
		wantErr        bool
	}{
		{
			name: "Valid gauges",
			args: args{
				name1:  "cpu_usage",
				value1: 75.5,
				name2:  "memory_usage",
				value2: 60.0,
			},
			mockBehavior: func(m *mocks, args args) {
				rows := sqlmock.NewRows([]string{"name", "value"}).
					AddRow(args.name1, args.value1).
					AddRow(args.name2, args.value2)
				expectedExec := `SELECT name, value FROM gauge_metrics`
				mock.ExpectQuery(regexp.QuoteMeta(expectedExec)).WillReturnRows(rows)

				expectedInsert := `INSERT INTO 
    							gauge_metrics(name, value) 
								VALUES ($1, $2) ON CONFLICT (name)
								 DO UPDATE SET value = $2
								`
				mock.ExpectExec(regexp.QuoteMeta(expectedInsert)).WithArgs(args.name1, args.value1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(regexp.QuoteMeta(expectedInsert)).WithArgs(args.name2, args.value2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedGauges: []storage.GaugeMetric{
				{Name: "cpu_usage", Value: 75.5},
				{Name: "memory_usage", Value: 60.0},
			},
			wantErr: false,
		},
		{
			name: "No gauges",
			mockBehavior: func(m *mocks, args args) {
				rows := sqlmock.NewRows([]string{"name", "value"})

				expectedExec := `SELECT name, value FROM gauge_metrics
								`
				mock.ExpectQuery(regexp.QuoteMeta(expectedExec)).WillReturnRows(rows)
			},
			expectedGauges: []storage.GaugeMetric{},
			wantErr:        false,
		},
		{
			name: "Query error",
			mockBehavior: func(m *mocks, args args) {
				expectedExec := `SELECT name, value FROM gauge_metrics
								`
				mock.ExpectQuery(regexp.QuoteMeta(expectedExec)).WillReturnError(fmt.Errorf("error selecting all gauges"))
			},
			expectedGauges: nil,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			m := &mocks{
				DB: sqlx.NewDb(db, "sqlmock"),
			}
			pg := &Postgres{DB: m.DB}

			tt.mockBehavior(m, tt.args)
			gauges, err := pg.GetAllGauges()

			if tt.wantErr {
				assert.Error(t, err, "Expected an error but got nil")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedGauges, gauges, "The returned gauges do not match the expected gauges")
			}

		})

	}
}

func TestGetAllCounters(t *testing.T) {
	db, mock, err := sqlmock.New() // db - *sqlmock.DB, mock - sqlmock.Sqlmock
	if err != nil {
		t.Fatalf("failed to open mock database: %v", err)
	}
	defer db.Close()

	type (
		args struct {
			name1  string
			value1 int64
			name2  string
			value2 int64
		}
		mockBehavior func(m *mocks, args args)
	)
	tests := []struct {
		name             string
		expectedCounters []storage.CounterMetric
		mockBehavior     mockBehavior
		wantErr          bool
		args             args
	}{
		{
			name: "Valid counters",
			args: args{
				name1:  "counter1",
				value1: 2,
				name2:  "counter2",
				value2: 4,
			},
			mockBehavior: func(m *mocks, args args) {
				rows := sqlmock.NewRows([]string{"name", "value"}).
					AddRow(args.name1, args.value1).
					AddRow(args.name2, args.value2)
				expectedExec := `SELECT name, value FROM counter_metrics`
				mock.ExpectQuery(regexp.QuoteMeta(expectedExec)).WillReturnRows(rows)

				expectedInsert := `INSERT INTO counter_metrics(name, value)VALUES ($1, $2) 
								ON CONFLICT(name)DO UPDATE SET value = counter_metrics.value 
								    + $2 RETURNING value
								`
				rowsInsert := sqlmock.NewRows([]string{"value"}).
					AddRow(args.value1).AddRow(args.value2)
				mock.ExpectQuery(regexp.QuoteMeta(expectedInsert)).WithArgs(args.name1, args.value1).
					WillReturnRows(rowsInsert)
				mock.ExpectQuery(regexp.QuoteMeta(expectedInsert)).WithArgs(args.name2, args.value2).
					WillReturnRows(rowsInsert)
			},
			expectedCounters: []storage.CounterMetric{
				{Name: "counter1", Value: 2},
				{Name: "counter2", Value: 4},
			},
			wantErr: false,
		},
		{
			name: "No counters",
			mockBehavior: func(m *mocks, args args) {
				rows := sqlmock.NewRows([]string{"name", "value"})

				expectedExec := `SELECT name, value FROM counter_metrics`
				mock.ExpectQuery(regexp.QuoteMeta(expectedExec)).WillReturnRows(rows)
			},
			expectedCounters: []storage.CounterMetric{},
			wantErr:          false,
		},
		{
			name: "Query error",
			mockBehavior: func(m *mocks, args args) {
				expectedExec := `SELECT name, value FROM counter_metrics`
				mock.ExpectQuery(regexp.QuoteMeta(expectedExec)).WillReturnError(fmt.Errorf("error selecting all counter"))
			},
			expectedCounters: nil,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			m := &mocks{
				DB: sqlx.NewDb(db, "sqlmock"),
			}
			pg := &Postgres{DB: m.DB}

			tt.mockBehavior(m, tt.args)
			counters, err := pg.GetAllCounters()

			if tt.wantErr {
				assert.Error(t, err, "Expected an error but got nil")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCounters, counters, "The returned gauges do not match the expected gauges")
			}

		})

	}
}

func TestBatchUpdate(t *testing.T) {
	db, mock, err := sqlmock.New() // db - *sqlmock.DB, mock - sqlmock.Sqlmock
	if err != nil {
		t.Fatalf("failed to open mock database: %v", err)
	}
	defer db.Close()

	type (
		args struct {
			name1        string
			value1       float64
			name2        string
			value2       float64
			mtypeGauge   string
			nameConter   string
			valueCounter int64
			mtypeCounter string
			metrics      []models.Metrics
		}
		mockBehavior func(m *mocks, args args)
	)

	tests := []struct {
		name         string
		metrics      []models.Metrics
		mockBehavior mockBehavior
		wantErr      bool
		args         args
	}{
		{
			name: "Valid batch update",
			args: args{
				name1:        "cpu_usage",
				value1:       75.5,
				name2:        "memory_usage",
				value2:       60.0,
				mtypeGauge:   "gauge",
				nameConter:   "counter1",
				valueCounter: 10,
				mtypeCounter: "counter",
			},
			mockBehavior: func(m *mocks, args args) {
				mock.ExpectBegin()
				defer mock.ExpectRollback()

				results := make([]models.Metrics, len(args.metrics))
				for _, metric := range args.metrics {
					expectedInsert := ` INSERT INTO
										gauge_metrics(name, value)
										VALUES ($1, $2) ON CONFLICT (name)
										DO UPDATE SET value = $2
										`
					expectedInsertCounter := `INSERT INTO 
    										counter_metrics(name, value) VALUES ($1, $2) 
											ON CONFLICT(name)DO UPDATE SET 
											    value = counter_metrics.value 
								    		+ $2 RETURNING value
											`
					if metric.MType == args.mtypeGauge {
						mock.ExpectExec(regexp.QuoteMeta(expectedInsert)).
							WithArgs(metric.ID, *metric.Value).
							WillReturnResult(sqlmock.NewResult(1, 1))
						mock.ExpectCommit()

					} else if metric.MType == args.mtypeCounter {
						rowsInsertCounter := sqlmock.NewRows([]string{"value"}).
							AddRow(metric.ID).AddRow(*metric.Delta)
						mock.ExpectQuery(regexp.QuoteMeta(expectedInsertCounter)).WithArgs(metric.ID, *metric.Delta).
							WillReturnRows(rowsInsertCounter)

					}
					results = append(results, metric)
				}

				mock.ExpectCommit()
			},
			metrics: []models.Metrics{
				{ID: "cpu_usage", MType: "gauge", Value: ptrToFloat64(75.5)},
				{ID: "memory_usage", MType: "gauge", Value: ptrToFloat64(60.0)},
				{ID: "counter1", MType: "counter", Delta: ptrToInt64(10)},
			},
			wantErr: false,
		},
		{
			name: "Invalid transaction start",
			args: args{
				name1:        "cpu_usage",
				value1:       75.5,
				name2:        "memory_usage",
				value2:       60.0,
				mtypeGauge:   "gauge",
				nameConter:   "counter1",
				valueCounter: 10,
				mtypeCounter: "counter",
			},
			mockBehavior: func(m *mocks, args args) {
				mock.ExpectBegin().WillReturnError(fmt.Errorf("error occured on creating tx on batchupdate"))
			},
			metrics: []models.Metrics{
				{ID: "cpu_usage", MType: "gauge", Value: ptrToFloat64(75.5)},
				{ID: "memory_usage", MType: "gauge", Value: ptrToFloat64(60.0)},
				{ID: "counter1", MType: "counter", Delta: ptrToInt64(10)},
			},
			wantErr: true,
		},
		{
			name: "SQL execution error",
			args: args{
				name1:        "cpu_usage",
				value1:       75.5,
				name2:        "memory_usage",
				value2:       60.0,
				mtypeGauge:   "gauge",
				nameConter:   "counter1",
				valueCounter: 10,
				mtypeCounter: "counter",
			},
			mockBehavior: func(m *mocks, args args) {
				mock.ExpectBegin()
				expectedInsert := `INSERT INTO 
                            gauge_metrics(name, value) 
                            VALUES ($1, $2) ON CONFLICT (name)
                            DO UPDATE SET value = $2`
				mock.ExpectExec(regexp.QuoteMeta(expectedInsert)).
					WithArgs(args.name1, args.value1).
					WillReturnError(fmt.Errorf("insert error"))

				mock.ExpectRollback()
			},
			metrics: []models.Metrics{
				{ID: "cpu_usage", MType: "gauge", Value: ptrToFloat64(75.5)},
				{ID: "memory_usage", MType: "gauge", Value: ptrToFloat64(60.0)},
				{ID: "counter1", MType: "counter", Delta: ptrToInt64(10)},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			m := &mocks{
				DB: sqlx.NewDb(db, "sqlmock"),
			}
			pg := &Postgres{DB: m.DB}

			tt.mockBehavior(m, tt.args)

			var buf bytes.Buffer
			err := pg.BatchUpdate(&buf, tt.metrics)

			if tt.wantErr {
				assert.Error(t, err, "Expected an error but got nil")
			} else {
				assert.NoError(t, err)
			}
			// Проверка выполнения всех ожидаемых запросов
			//assert.NoError(t, mock.ExpectationsWereMet(), "Not all SQL expectations were met")
		})
	}
}

func TestBatchUpdate2(t *testing.T) {
	db, mock, err := sqlmock.New() // db - *sqlmock.DB, mock - sqlmock.Sqlmock
	if err != nil {
		t.Fatalf("failed to open mock database: %v", err)
	}
	defer db.Close()

	type (
		args struct {
			name1        string
			value1       float64
			name2        string
			value2       float64
			mtypeGauge   string
			nameConter   string
			valueCounter int64
			mtypeCounter string
			metrics      []models.Metrics
		}
		mockBehavior func(m *mocks, args args)
	)

	tests := []struct {
		name         string
		metrics      []models.Metrics
		mockBehavior mockBehavior
		wantErr      bool
		args         args
	}{
		{
			name: "Valid batch update",
			args: args{
				name1:        "cpu_usage",
				value1:       75.5,
				name2:        "memory_usage",
				value2:       60.0,
				mtypeGauge:   "gauge",
				nameConter:   "counter1",
				valueCounter: 10,
				mtypeCounter: "counter",
			},
			mockBehavior: func(m *mocks, args args) {
				mock.ExpectBegin()
				//defer mock.ExpectRollback()

				// Мокаем вставку/обновление метрик в базе данных
				results := make([]models.Metrics, len(args.metrics))
				for _, metric := range args.metrics {
					expectedInsert := `INSERT INTO 
    								gauge_metrics(name, value) 
									VALUES ($1, $2) ON CONFLICT (name)
									 DO UPDATE SET value = $2`
					expectedInsertCounter := `INSERT INTO 
    										counter_metrics(name, value) VALUES ($1, $2) 
											ON CONFLICT(name)DO UPDATE SET value = counter_metrics.value 
								    		+ $2 RETURNING value
											`
					if metric.MType == args.mtypeGauge {
						mock.ExpectExec(regexp.QuoteMeta(expectedInsert)).
							WithArgs(metric.ID, *metric.Value).
							WillReturnResult(sqlmock.NewResult(1, 1))
					} else if metric.MType == args.mtypeCounter {
						rowsInsertCounter := sqlmock.NewRows([]string{"value"}).
							AddRow(metric.ID).AddRow(*metric.Delta)
						mock.ExpectQuery(regexp.QuoteMeta(expectedInsertCounter)).WithArgs(metric.ID, *metric.Delta).
							WillReturnRows(rowsInsertCounter)
					}
					results = append(results, metric)
				}

				mock.ExpectCommit()
			},
			metrics: []models.Metrics{
				{ID: "cpu_usage", MType: "gauge", Value: ptrToFloat64(75.5)},
				{ID: "memory_usage", MType: "gauge", Value: ptrToFloat64(60.0)},
				{ID: "counter1", MType: "counter", Delta: ptrToInt64(10)},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок для DB и Storage
			c := gomock.NewController(t)
			defer c.Finish()

			m := &mocks{
				DB: sqlx.NewDb(db, "sqlmock"),
			}
			pg := &Postgres{DB: m.DB}

			// Настроим поведение мока
			tt.mockBehavior(m, tt.args)

			// Создаем буфер для записи результата
			var buf bytes.Buffer

			// Выполним тестируемый метод
			err := pg.BatchUpdate(&buf, tt.metrics)

			// Проверка на наличие ошибок
			if tt.wantErr {
				assert.Error(t, err, "Expected an error but got nil")
			} else {
				assert.NoError(t, err)
			}

			// Проверка выполнения всех ожидаемых запросов
			//assert.NoError(t, mock.ExpectationsWereMet(), "Not all SQL expectations were met")
		})
	}
}

func TestPing(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true)) // Включаем мониторинг пингов
	if err != nil {
		t.Fatalf("failed to open mock database: %v", err)
	}
	defer db.Close()

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
	}{
		{
			name: "successful ping",
			setupMock: func() {
				mock.ExpectPing()
			},
			wantErr: false,
		},
		{
			name: "failed ping",
			setupMock: func() {
				mock.ExpectPing().WillReturnError(fmt.Errorf("failed to ping database"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			pg := &Postgres{DB: sqlx.NewDb(db, "sqlmock")}

			err := pg.Ping()

			if tt.wantErr {
				assert.Error(t, err, "expected an error but got nil")
			} else {
				assert.NoError(t, err, "expected no error but got one")
			}

			assert.NoError(t, mock.ExpectationsWereMet(), "not all expectations were met")
		})
	}
}

func TestInitDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock database: %v", err)
	}
	defer db.Close()

	type mockBehavior func(mock sqlmock.Sqlmock)

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		wantErr      bool
	}{
		{
			name: "successful initDB",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				expectedExecCounter := `CREATE TABLE IF NOT EXISTS counter_metrics (name char(30) UNIQUE, value integer)`
				expectedExecGauge := `CREATE TABLE IF NOT EXISTS gauge_metrics (name char(30) UNIQUE, value double precision`
				mock.ExpectPing()
				mock.ExpectExec(regexp.QuoteMeta(expectedExecCounter)).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(regexp.QuoteMeta(expectedExecGauge)).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: false,
		},
		{
			name: "error creating counter_metrics table",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				expectedExecCounter := `CREATE TABLE IF NOT EXISTS counter_metrics (name char(30) UNIQUE, value integer)`
				mock.ExpectPing()
				mock.ExpectExec(regexp.QuoteMeta(expectedExecCounter)).
					WillReturnError(fmt.Errorf("failed to create table counter_metrics"))
			},
			wantErr: true,
		},
		{
			name: "error creating gauge_metrics table",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				expectedExecCounter := `CREATE TABLE IF NOT EXISTS counter_metrics (name char(30) UNIQUE, value integer)`
				expectedExecGauge := `CREATE TABLE IF NOT EXISTS gauge_metrics (name char(30) UNIQUE, value double precision`
				mock.ExpectPing()
				mock.ExpectExec(regexp.QuoteMeta(expectedExecCounter)).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(regexp.QuoteMeta(expectedExecGauge)).
					WillReturnError(fmt.Errorf("failed to create table gauge_metrics"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(mock)

			pg := &Postgres{DB: sqlx.NewDb(db, "sqlmock")}
			ctx := context.Background()

			err := pg.InitDB(ctx)

			if tt.wantErr {
				assert.Error(t, err, "expected an error but got nil")
			} else {
				assert.NoError(t, err, "expected no error but got one")
			}

			assert.NoError(t, mock.ExpectationsWereMet(), "not all expectations were met")
		})
	}
}

// Вспомогательная функция для указателя на float64
func ptrToFloat64(val float64) *float64 {
	return &val
}

// Вспомогательная функция для указателя на int64
func ptrToInt64(val int64) *int64 {
	return &val
}
