package test

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AnxVit/avito/internal/config"
	"github.com/AnxVit/avito/internal/domain/models"
	"github.com/AnxVit/avito/internal/http-server/server"
	"github.com/AnxVit/avito/internal/storage/cache"
	"github.com/AnxVit/avito/internal/storage/postgres"
	pgcontainer "github.com/AnxVit/avito/tests/container/postgres"

	"github.com/AnxVit/avito/tests/migrate"

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

	localcache, err := cache.New(repo)
	s.Require().NoError(err)
	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	s.server = httptest.NewServer(server.New(cfgServer, repo, localcache, logger).Router)

	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.FilesMultiTables("fixtures/storage/1_banner.yaml"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
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

func (s *TestSuite) TestGetUserBannerpNotAuth() {
	header := http.Header{
		"token": []string{"token"},
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

	s.Assert().Equal(http.StatusUnauthorized, res.StatusCode)
}

func (s *TestSuite) TestGetUserBannerpNotAccess() {
	header := http.Header{
		"token": []string{"user_token"},
	}
	u, _ := url.Parse(s.server.URL + "/user_banner?tag_id=4&feature_id=2")
	req := &http.Request{
		Method: "GET",
		Header: header,
		URL:    u,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *TestSuite) TestGetUserBannerNotFound1() {
	header := http.Header{
		"token": []string{"user_token"},
	}
	u, _ := url.Parse(s.server.URL + "/user_banner?tag_id=6&feature_id=1")
	req := &http.Request{
		Method: "GET",
		Header: header,
		URL:    u,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusNotFound, res.StatusCode)
}

func (s *TestSuite) TestGetUserBannerNotFound2() {
	header := http.Header{
		"token": []string{"user_token"},
	}
	u, _ := url.Parse(s.server.URL + "/user_banner?tag_id=5&feature_id=3")
	req := &http.Request{
		Method: "GET",
		Header: header,
		URL:    u,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusNotFound, res.StatusCode)
}

func (s *TestSuite) TestGetUserBannerBadQuerry1() {
	header := http.Header{
		"token": []string{"user_token"},
	}
	u, _ := url.Parse(s.server.URL + "/user_banner?tag_id=5")
	req := &http.Request{
		Method: "GET",
		Header: header,
		URL:    u,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *TestSuite) TestGetUserBannerBadQuerry2() {
	header := http.Header{
		"token": []string{"user_token"},
	}
	u, _ := url.Parse(s.server.URL + "/user_banner?feature_id=1")
	req := &http.Request{
		Method: "GET",
		Header: header,
		URL:    u,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *TestSuite) TestGetBanner() {
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner?feature_id=2&tag_id=2")
	req := &http.Request{
		Method: "GET",
		Header: header,
		URL:    u,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var response []models.BannerDB
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Assert().Equal(int64(3), *response[0].ID)
}

func (s *TestSuite) TestGetBannerFeature() {
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner?feature_id=2")
	req := &http.Request{
		Method: "GET",
		Header: header,
		URL:    u,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var response []models.BannerDB
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	result := make([]int64, 0)
	for _, banner := range response {
		result = append(result, *banner.ID)
	}
	target := []int64{2, 3}

	s.Assert().ElementsMatch(target, result)
}

func (s *TestSuite) TestGetBannerTag() {
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner?tag_id=1")
	req := &http.Request{
		Method: "GET",
		Header: header,
		URL:    u,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var response []models.BannerDB
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	result := make([]int64, 0)
	for _, banner := range response {
		result = append(result, *banner.ID)
	}
	target := []int64{1, 3}
	s.Assert().ElementsMatch(target, result)
}

func (s *TestSuite) TestGetBannerLimit() {
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner?tag_id=1&limit=1")
	req := &http.Request{
		Method: "GET",
		Header: header,
		URL:    u,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var response []models.BannerDB
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Assert().Equal(int64(1), *response[0].ID)
}
func (s *TestSuite) TestGetBannerOffset() {
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner?tag_id=1&offset=1")
	req := &http.Request{
		Method: "GET",
		Header: header,
		URL:    u,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var response []models.BannerDB
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Assert().Equal(int64(3), *response[0].ID)
}

func (s *TestSuite) TestPostBanner() {
	requestBody := `{
		"tag_ids": [5], 
		"feature_id": 1, 
		"content": {
			"color": "yellow", 
			"object": "sun"
			},
		"is_active": true
		}`
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner")
	reader := io.NopCloser(strings.NewReader(requestBody))
	req := &http.Request{
		Method: "POST",
		Header: header,
		URL:    u,
		Body:   reader,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusCreated, res.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Assert().Equal("OK", response["status"].(string))
}

func (s *TestSuite) TestPostBannerNotAuth() {
	requestBody := `{
		"tag_ids": [5], 
		"feature_id": 1, 
		"content": {
			"color": "yellow", 
			"object": "sun"
			},
		"is_active": true
		}`
	header := http.Header{
		"token": []string{"token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner")
	reader := io.NopCloser(strings.NewReader(requestBody))
	req := &http.Request{
		Method: "POST",
		Header: header,
		URL:    u,
		Body:   reader,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusUnauthorized, res.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().Error(err)
}

func (s *TestSuite) TestPostBannerNotAccess() {
	requestBody := `{
		"tag_ids": [5], 
		"feature_id": 1, 
		"content": {
			"color": "yellow", 
			"object": "sun"
			},
		"is_active" :true
		}`
	header := http.Header{
		"token": []string{"user_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner")
	reader := io.NopCloser(strings.NewReader(requestBody))
	req := &http.Request{
		Method: "POST",
		Header: header,
		URL:    u,
		Body:   reader,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusForbidden, res.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().Error(err)
}

func (s *TestSuite) TestPostBannerBadRequest1() {
	requestBody := `{
		"tag_ids": [a],
		"feature_id": 1,
		"content": {
			"color": "yellow",
			"object": "sun"
			},
		"is_active": true
		}`
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner")
	reader := io.NopCloser(strings.NewReader(requestBody))
	req := &http.Request{
		Method: "POST",
		Header: header,
		URL:    u,
		Body:   reader,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)
	s.Require().Equal("unsupported type of value", response["error"].(string))
}

func (s *TestSuite) TestPostBannerBadRequest2() {
	requestBody := `{
		"tag_ids": 1,
		"feature_id": 1,
		"content": {
			"color": "yellow",
			"object": "sun"
			},
		"is_active": true
		}`
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner")
	reader := io.NopCloser(strings.NewReader(requestBody))
	req := &http.Request{
		Method: "POST",
		Header: header,
		URL:    u,
		Body:   reader,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)
	s.Require().Equal("invalid body", response["error"].(string))
}

func (s *TestSuite) TestPostBannerBadRequest3() {
	requestBody := `{
		"tag_id": [1],
		"feature": 1, 
		"content": {
			"color": "yellow", 
			"object": "sun"
			},
		"is_active": true
		}`
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner")
	reader := io.NopCloser(strings.NewReader(requestBody))
	req := &http.Request{
		Method: "POST",
		Header: header,
		URL:    u,
		Body:   reader,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)
	s.Require().Equal("invalid body", response["error"].(string))
}

func (s *TestSuite) TestPatchBanner() {
	requestBody := `{
		"tag_ids": [3, 4, 5],
		"feature_id": 1, 
		"content": {
			"color": "yellow", 
			"object": "sun"
			},
		"is_active": false
		}`
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner/1")
	reader := io.NopCloser(strings.NewReader(requestBody))
	req := &http.Request{
		Method: "PATCH",
		Header: header,
		URL:    u,
		Body:   reader,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)
	s.Require().Equal("OK", response["status"].(string))
}

func (s *TestSuite) TestPatchBannerNotFound() {
	requestBody := `{
		"tag_ids": [3, 4, 5],
		"feature_id": 1, 
		"content": {
			"color": "yellow", 
			"object": "sun"
			},
		"is_active": false
		}`
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner/100")
	reader := io.NopCloser(strings.NewReader(requestBody))
	req := &http.Request{
		Method: "PATCH",
		Header: header,
		URL:    u,
		Body:   reader,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusNotFound, res.StatusCode)
}

func (s *TestSuite) TestPatchBannerNullable() {
	requestBody := `
	{
		"tag_ids": null,
		"feature_id": null, 
		"content": null,
		"is_active": null
	}`
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner/2")
	reader := io.NopCloser(strings.NewReader(requestBody))
	req := &http.Request{
		Method: "PATCH",
		Header: header,
		URL:    u,
		Body:   reader,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusOK, res.StatusCode)
}

func (s *TestSuite) TestPatchBannerBadRequest1() {
	requestBody := `
	{
		"tag_ids": [3, 4, 5],
		"feature_id": 1, 
		"content": 1,
		"is_active": false
	}`
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner/2")
	reader := io.NopCloser(strings.NewReader(requestBody))
	req := &http.Request{
		Method: "PATCH",
		Header: header,
		URL:    u,
		Body:   reader,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)
	s.Assert().Equal("invalid body", response["error"].(string))
}

func (s *TestSuite) TestPatchBannerBadRequest2() {
	requestBody := `
	{
		"tag_ids": [a],
		"feature_id": 1, 
		"content": {
			"color": "yellow", 
			"object": "sun"
		},,
		"is_active": false
	}`
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner/2")
	reader := io.NopCloser(strings.NewReader(requestBody))
	req := &http.Request{
		Method: "PATCH",
		Header: header,
		URL:    u,
		Body:   reader,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)
	s.Assert().Equal("unsupported type of value", response["error"].(string))
}

func (s *TestSuite) TestPatchBannerBadRequest3() {
	requestBody := `
	{
		"tag_ids": 1,
		"feature_id": 1, 
		"content": {
			"color": "yellow", 
			"object": "sun"
		},,
		"is_active": false
	}`
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner/2")
	reader := io.NopCloser(strings.NewReader(requestBody))
	req := &http.Request{
		Method: "PATCH",
		Header: header,
		URL:    u,
		Body:   reader,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)
	s.Assert().Equal("unsupported type of value", response["error"].(string))
}

func (s *TestSuite) TestPatchBannerBadRequest4() {
	requestBody := `
	{
		"feature_id": 1, 
		"content": {
			"color": "yellow", 
			"object": "sun"
		},,
		"is_active": false
	}`
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner/2")
	reader := io.NopCloser(strings.NewReader(requestBody))
	req := &http.Request{
		Method: "PATCH",
		Header: header,
		URL:    u,
		Body:   reader,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)
	s.Assert().Equal("unsupported type of value", response["error"].(string))
}

func (s *TestSuite) TestDeleteBanner() {
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner/4")
	req := &http.Request{
		Method: "DELETE",
		Header: header,
		URL:    u,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusNoContent, res.StatusCode)

	// var response map[string]interface{}
	// err = json.NewDecoder(res.Body).Decode(&response)
	// s.Require().NoError(err)
	// s.Assert().Equal("not correct id", response["error"].(string))
}

func (s *TestSuite) TestDeleteBannerBadRequest1() {
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner/a")
	req := &http.Request{
		Method: "DELETE",
		Header: header,
		URL:    u,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)
	s.Assert().Equal("not correct id", response["error"].(string))
}

func (s *TestSuite) TestDeleteBannerBadRequest2() {
	header := http.Header{
		"token": []string{"admin_token"},
	}
	u, _ := url.Parse(s.server.URL + "/banner/5")
	req := &http.Request{
		Method: "DELETE",
		Header: header,
		URL:    u,
	}

	res, err := s.server.Client().Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Assert().Equal(http.StatusNotFound, res.StatusCode)
}
