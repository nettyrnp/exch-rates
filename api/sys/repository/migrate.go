package repository

import (
	migrate "github.com/rubenv/sql-migrate"
)

var migrations = &migrate.MemoryMigrationSource{
	Migrations: []*migrate.Migration{
		{
			Id: "00001_initial_migration",
			Up: []string{
				`CREATE TABLE exchange_rate
					(
					  id SERIAL PRIMARY KEY,
					  time TIMESTAMP NOT NULL,
					  currency VARCHAR(3) NOT NULL,
					  rate NUMERIC NOT NULL,
					  created_at TIMESTAMP NOT NULL default NOW()
					);`,

				"CREATE INDEX exchange_rate_idx ON exchange_rate (time,currency);",

				`INSERT INTO exchange_rate (time, currency, rate)
					VALUES 
						('2020-03-10 15:07:30'::timestamp,'USD', 79.38426),
						('2020-03-10 16:07:30'::timestamp,'USD', 79.4),
						('2020-03-10 17:07:30'::timestamp,'USD', 79.58426),
						('2020-03-10 18:07:30'::timestamp,'USD', 79.48426),

						('2020-03-15 15:07:30'::timestamp,'USD', 77.38426),
						('2020-03-15 16:07:30'::timestamp,'USD', 77.4),
						('2020-03-15 17:07:30'::timestamp,'USD', 77.58426),
						('2020-03-15 18:07:30'::timestamp,'USD', 77.48426),

						('2020-03-20 15:07:30'::timestamp,'USD', 66.7),
						('2020-03-20 15:07:00'::timestamp,'USD', 66.78888),
						('2020-03-20 15:08:00'::timestamp,'USD', 67.8),
						('2020-03-20 15:08:30'::timestamp,'USD', 67.89999),
						('2020-03-20 15:09:00'::timestamp,'USD', 68.9),

						('2020-03-21 17:17:30'::timestamp,'EUR', 86.7),
						('2020-03-21 17:17:00'::timestamp,'EUR', 86.78888),
						('2020-03-21 17:18:00'::timestamp,'EUR', 87.8),
						('2020-03-21 17:18:30'::timestamp,'EUR', 87.89999),
						('2020-03-21 17:19:00'::timestamp,'EUR', 88.9)
					;`,
			},
			Down: []string{
				"DROP INDEX IF EXISTS exchange_rate_idx;",
				"DROP TABLE IF EXISTS exchange_rate;",
			},
		},
	},
}

func (r *RDBMSRepository) MigrateUp() error {
	_, err := migrate.Exec(r.db, r.Cfg.Driver, migrations, migrate.Up)
	return err
}

func (r *RDBMSRepository) MigrateDown() error {
	_, err := migrate.Exec(r.db, r.Cfg.Driver, migrations, migrate.Down)
	return err
}
