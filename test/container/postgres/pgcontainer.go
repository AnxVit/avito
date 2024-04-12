package pgcontainer

import (
	"context"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	testcontainers.Container
	Port string
	Host string
}

func (c PostgresContainer) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", "postgres", "1234", c.Host, c.Port, "test_banner")
}

func New(ctx context.Context) (*PostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "1234",
			"POSTGRES_DB":       "test_banner",
		},
		ExposedPorts: []string{"5432/tcp"},
		Image:        "postgres:alpine",
		WaitingFor: wait.ForExec([]string{"pg_isready"}).
			WithPollInterval(1 * time.Second).
			WithExitCodeMatcher(func(exitCode int) bool {
				return exitCode == 0
			}),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}
	return &PostgresContainer{
		Container: container,
		Port:      port.Port(),
		Host:      host,
	}, nil
}
