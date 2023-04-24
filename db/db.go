package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/initializer"
)

var pool *pgxpool.Pool

func init() {
	initializer.Register(initializer.MOD_DB, func(c *initializer.InitContext) {
		conf, err := pgxpool.ParseConfig(config.General.DB)
		log.FatalIfErr(err, "parsing config")

		db, err := pgxpool.ConnectConfig(context.Background(), conf)
		log.FatalIfErr(err, "connecting to pool")

		err = db.Ping(context.Background())
		log.FatalIfErr(err, "pinging the db")

		pool = db

		for _, sql := range setup {
			_, err = Exec(sql[0])
			log.FatalIfErr(err, sql[1])
		}
	}, nil)

	initializer.RegisterCloser(initializer.MOD_DB, func() {
		pool.Close()
	})
}

func Exec(sql string, args ...any) (pgconn.CommandTag, error) {
	v1, err := pool.Exec(context.Background(), sql, args...)
	if err != nil {
		log.Error("Couldn't exec '%s': %v", sql, err)
	}
	return v1, err
}

func Query(sql string, args ...any) (pgx.Rows, error) {
	rows, err := pool.Query(context.Background(), sql, args...)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Error("Couldn't fetch '%s': %v", sql, err)
	}
	return rows, err
}

func QueryRow(sql string, args []any, scanTarget ...any) error {
	row := pool.QueryRow(context.Background(), sql, args...)
	err := row.Scan(scanTarget...)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Error("Couldn't row fetch '%s': %v", sql, err)
	}
	return err
}

// Query row with 1 condition
func QueryRowID(sql string, arg any, scanTarget ...any) error {
	row := pool.QueryRow(context.Background(), sql, arg)
	err := row.Scan(scanTarget...)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Error("Couldn't row fetch '%s': %v", sql, err)
	}
	return err
}

func Exists(table string, conditions string, values ...any) bool {
	var ret bool
	err := pool.QueryRow(context.Background(), fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE %s)`, table, conditions), values...).Scan(&ret)
	if err != nil {
		log.Error("Couldn't row exist fetch from table '%s', conditions '%s': %v", table, conditions, err)
	}
	return ret
}

func Close() {
	pool.Close()
}
