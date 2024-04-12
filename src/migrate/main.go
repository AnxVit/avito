package main

import (
	"Auth-Reg/internal/config"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

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

func main() {
	flags.Parse(os.Args[1:])

	args := flags.Args()

	command := args[0]

	cfg := config.MustLoad()
	c := cfg.DB

	dbString := fmt.Sprintf(fmtDBString, c.Host, c.User, c.Password, c.DBName, c.Port)

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
		log.Fatalf("migrate %v: %v", command, err)
	}
}
