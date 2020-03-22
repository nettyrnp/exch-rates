package repository

import (
	"context"
	"database/sql"
	qu "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"github.com/nettyrnp/exch-rates/api/sys/entity"
	"github.com/pkg/errors"
	"time"
)

type Config struct {
	Driver string
	DSN    string
}

type Repository interface {
	GetAverage(ctx context.Context, currency string, from, till time.Time) (float64, error)
	GetHistory(ctx context.Context, opts RatesQueryOpts) ([]entity.Average, int, error)
	GetMomental(ctx context.Context, currency string, moment time.Time) (float64, error)
	AddExchrate(ctx context.Context, e *entity.Exchrate) error
}

type RDBMSRepository struct {
	Name string
	db   *sql.DB
	Cfg  Config
}

func (r *RDBMSRepository) GetAverage(ctx context.Context, currency string, from, till time.Time) (float64, error) {
	var rate float64

	execErr := r.runInTx(func(tx *sql.Tx) error {
		var rate0 float64
		selectMax := qu.StatementBuilder.PlaceholderFormat(qu.Dollar).
			Select("AVG(rate)").
			From("exchange_rate")
		queryRows, args, err := selectMax.
			Where(qu.And{qu.Eq{"currency": currency}, qu.GtOrEq{"time": from}, qu.LtOrEq{"time": till}}).
			ToSql()
		if err != nil {
			return err
		}
		if err := tx.QueryRowContext(ctx, queryRows, args...).Scan(&rate0); err != nil {
			return err
		}

		rate = rate0
		return nil

	}, sql.LevelReadCommitted)

	if execErr != nil {
		return 0, execErr
	}
	return rate, nil
}

func (r *RDBMSRepository) GetHistory(ctx context.Context, opts RatesQueryOpts) ([]entity.Average, int, error) {
	var exchrates []entity.Average
	var total int

	execErr := r.runInTx(func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, "SELECT extract(epoch from time)::int/$1 AS AggregatedTime, avg(rate) "+
			"FROM exchange_rate "+
			"WHERE currency=$2 AND time>=$3 AND time<=$4 "+
			"GROUP BY AggregatedTime "+
			"ORDER BY AggregatedTime "+
			"LIMIT $5 OFFSET $6 ",
			opts.SecondsInInterval, opts.Currency, opts.From, opts.Till, opts.Limit, opts.Offset)
		if err != nil {
			return err
		}

		exchrates0, err := scanExchrateRows(rows, opts.SecondsInInterval, opts.Limit) // todo: return total
		if err != nil {
			return err
		}

		exchrates = exchrates0
		return nil

	}, sql.LevelReadCommitted)

	if execErr != nil {
		return nil, 0, execErr
	}
	return exchrates, total, nil
}

func (r *RDBMSRepository) GetMomental(ctx context.Context, currency string, moment time.Time) (float64, error) {
	var rate float64

	execErr := r.runInTx(func(tx *sql.Tx) error {
		var closestTime time.Time
		selectMax := qu.StatementBuilder.PlaceholderFormat(qu.Dollar). // todo: in single query with selectExchrates
										Select("MAX(time)").
										From("exchange_rate")
		queryRows, args, err := selectMax.
			Where(qu.And{qu.Eq{"currency": currency}, qu.LtOrEq{"time": moment}}).
			Limit(1).ToSql()
		if err != nil {
			return err
		}
		if err := tx.QueryRowContext(ctx, queryRows, args...).Scan(&closestTime); err != nil {
			return err
		}

		selectExchrates := qu.StatementBuilder.PlaceholderFormat(qu.Dollar).
			Select("rate").
			From("exchange_rate")
		query, args, err := selectExchrates.
			Where(qu.And{qu.Eq{"currency": currency}, qu.Eq{"time": closestTime}}).
			Limit(1).ToSql()
		if err != nil {
			return err
		}
		var rate0 float64
		if err := tx.QueryRowContext(ctx, query, args...).Scan(&rate0); err != nil {
			return err
		}

		rate = rate0
		return nil

	}, sql.LevelReadCommitted)

	if execErr != nil {
		return 0, execErr
	}
	return rate, nil
}

func (r *RDBMSRepository) AddExchrate(ctx context.Context, e *entity.Exchrate) error {
	return r.runInTx(func(tx *sql.Tx) error {
		psql := qu.StatementBuilder.PlaceholderFormat(qu.Dollar)
		query, args, err := psql.Insert("exchange_rate").Columns("time", "currency", "rate").
			Values(e.Time, e.Currency, e.Rate).
			ToSql()
		if err != nil {
			return err
		}
		if _, execErr := tx.ExecContext(ctx, query, args...); execErr != nil {
			return execErr
		}

		return nil

	}, sql.LevelSerializable)
}

func scanExchrateRows(rows *sql.Rows, secondsInInterval, limit uint64) ([]entity.Average, error) {
	exchrates := make([]entity.Average, 0, limit)
	defer rows.Close()

	for rows.Next() {
		e := &entity.Average{}
		var aggrTime int64

		if err := rows.Scan(&aggrTime, &e.Rate); err != nil {
			return nil, err
		}
		e.Time = time.Unix(aggrTime*int64(secondsInInterval), 0)
		exchrates = append(exchrates, *e)
	}
	return exchrates, nil
}

func (r *RDBMSRepository) Init() error {
	var err error
	r.db, err = connect(r.Cfg)
	if err != nil {
		return err
	}
	return nil
}

type dbExecutor func(tx *sql.Tx) error

func (r *RDBMSRepository) runInTx(executor dbExecutor, isoLevel sql.IsolationLevel) error {
	tx, err := r.db.BeginTx(context.Background(), &sql.TxOptions{Isolation: isoLevel})
	if err != nil {
		return err
	}

	if err := executor(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(err, rollbackErr.Error())
		}
		return err
	}

	return tx.Commit()
}

func connect(cfg Config) (*sql.DB, error) {
	db, openErr := sql.Open(cfg.Driver, cfg.DSN)
	if openErr != nil {
		return nil, openErr
	}

	if pingErr := db.Ping(); pingErr != nil {
		return nil, pingErr
	}

	return db, nil
}
