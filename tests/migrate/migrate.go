package migrate

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/pressly/goose/v3"
)

func Migrate(dsn string) error {
	db, err := goose.OpenDBWithDriver("postgres", dsn)
	if err != nil {
		log.Fatalf(err.Error())
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf(err.Error())
		}
	}()

	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	return goose.RunContext(context.Background(), "up", db, dir+"/../migrations")
}
