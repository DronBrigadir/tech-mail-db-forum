package database

import "github.com/jackc/pgx"

type TxOrDb interface {
	QueryRow(string, ...interface{}) *pgx.Row
	Query(string, ...interface{}) (*pgx.Rows, error)
	Exec(string, ...interface{}) (commandTag pgx.CommandTag, err error)
}

var config = pgx.ConnConfig{
	Host:     "localhost",
	Port:     5432,
	Database: "docker",
	User:     "docker",
	Password: "docker",
}

var Connection *pgx.ConnPool

func Init() error {
	var err error
	Connection, err = pgx.NewConnPool(
		pgx.ConnPoolConfig{
			ConnConfig:     config,
			MaxConnections: 50,
		})

	return err
}
