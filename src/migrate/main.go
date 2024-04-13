package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/AnxVit/avito/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const (
	dialect     = "pgx"
	fmtDBString = "host=%s user=%s password=%s dbname=%s port=%d sslmode=disable"
)

var (
	flags = flag.NewFlagSet("migrate", flag.ExitOnError)
	dir   = flags.String("dir", "migrations", "directory with mifgration files")
)

func GetDNS(c *config.DB) string {
	return fmt.Sprintf(fmtDBString, c.Host, c.User, c.Password, c.DBName, c.Port)
}

func main() {
	flags.Parse(os.Args[1:])

	args := flags.Args()

	command := args[0]

	cfg := config.MustLoad()
	c := cfg.DB

	dbString := GetDNS(&c)

	db, err := goose.OpenDBWithDriver(dialect, dbString)
	if err != nil {
		log.Fatalf(err.Error())
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf(err.Error())
		}
	}()

	if err := goose.RunContext(context.Background(), command, db, *dir, args[1:]...); err != nil {
		log.Printf("migrate %v: %v", command, err)
	}
}
