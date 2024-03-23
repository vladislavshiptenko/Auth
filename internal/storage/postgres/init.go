package postgres

import (
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"log"
	"os"
)

type Storage struct {
	db         *sql.DB
	sqlBuilder *sq.StatementBuilderType
}

func New(path string) (*Storage, error) {
	// TODO: параметризовать конфиг для подключения к бд
	const op = "storage.postgres.New"

	connStr := "user=postgres password=postgres dbname=testdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	content := make([]byte, 100000)
	n, err := file.Read(content)
	if err != nil {
		log.Fatal(err)
	}

	content = content[:n]

	_, err = db.Exec(string(content))
	if err != nil {
		log.Fatal(err)
	}

	sqlBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)

	return &Storage{db: db, sqlBuilder: &sqlBuilder}, nil
}
