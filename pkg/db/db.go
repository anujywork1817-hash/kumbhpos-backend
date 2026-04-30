package db

import (
    "os"

    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

var DB *sqlx.DB

func Connect() error {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        // fallback for local development
        dsn = "host=" + os.Getenv("DB_HOST") +
            " port=" + os.Getenv("DB_PORT") +
            " user=" + os.Getenv("DB_USER") +
            " password=" + os.Getenv("DB_PASSWORD") +
            " dbname=" + os.Getenv("DB_NAME") +
            " sslmode=disable"
    }
    var err error
    DB, err = sqlx.Connect("postgres", dsn)
    return err
}
