package test

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/AnxVit/avito/internal/config"
	"github.com/AnxVit/avito/internal/http-server/server"
	"github.com/AnxVit/avito/internal/storage/cache"
	"github.com/AnxVit/avito/internal/storage/postgres"
	pgcontainer "github.com/AnxVit/avito/test/container/postgres"

	"github.com/AnxVit/avito/test/migrate"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	psqlContainer *pgcontainer.PostgresContainer
	server        *httptest.Server
}

func (s *TestSuite) SetupSuite() {
	// create db container
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()

	psqlContainer, err := pgcontainer.New(ctx)
	s.Require().NoError(err)
	s.psqlContainer = psqlContainer

	err = migrate.Migrate(psqlContainer.GetDSN())
	s.Require().NoError(err)

	port, _ := strconv.Atoi(psqlContainer.Port)
	cfgDB := &config.DB{
		User:     "postgres",
		Password: "1234",
		Host:     psqlContainer.Host,
		Port:     port,
		DBName:   "test_banner",
	}
	cfgServer := &config.Server{
		Host:    "localhost",
		Port:    "8082",
		Timeout: 1 * time.Second,
	}

	repo, err := postgres.New(cfgDB)
	s.Require().NoError(err)

	storage, err := cache.New(repo)
	s.Require().NoError(err)
	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	s.server = httptest.NewServer(server.New(cfgServer, storage, logger).Router)
}

func (s *TestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.psqlContainer.Terminate(ctx))

	s.server.Close()
}

func TestSuite_Run(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestGetUserBanner() {
	//------------------------------------------------
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.FilesMultiTables("fixtures/storage/1_banner.yaml"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())

	// test body of below ----------------------------
	header := http.Header{
		"token": []string{"user_token"},
	}
	u, _ := url.Parse(s.server.URL + "/user_banner?tag_id=1&feature_id=1")
	req := &http.Request{
		Method: "GET",
		Header: header,
		URL:    u,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)
	s.Assert().Equal("blue", response["color"].(string))
	s.Assert().Equal("sky", response["object"].(string))
}
