package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/AnxVit/avito/internal/config"
	"github.com/AnxVit/avito/internal/domain/models"
	"github.com/AnxVit/avito/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	DB *pgxpool.Pool
}

func New(storage *config.DB) (*Repo, error) {
	const op = "storage.postgres.New"
	psqlInfo := fmt.Sprintf("user=%s password=%s host=%s "+
		"port=%d dbname=%s sslmode=disable",
		storage.User, storage.Password, storage.Host, storage.Port, storage.DBName)
	db, err := pgxpool.New(context.Background(), psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Repo{
		DB: db,
	}, nil
}

func (s *Repo) GetUserBanner(tag, feature int, admin bool) (map[string]interface{}, error) {
	const op = "storage.postgres.GetUserBanner"

	var banner map[string]interface{}
	var access *bool
	err := s.DB.QueryRow(context.Background(),
		`SELECT 
			content,
		 	access 
		FROM banner
		WHERE feature = $1 AND id = ANY(
			SELECT 
				BannerID
			FROM 
				bannertag
			WHERE TagID = $2
			);`, feature, tag).Scan(&banner, &access)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrBannerNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if access == nil || (!*access && !admin) {
		return nil, storage.ErrNotAccess
	}
	return banner, nil
}

func (s *Repo) GetBanner(tag, feature, limit, offset string) ([]models.BannerDB, error) {
	const op = "storage.postgres.GetBanner"

	var buffer bytes.Buffer

	query := `
	SELECT 
		id,
		array_agg(bannertag.tagid) tag,
		feature,
		content,
		access,
		created_at,
		updated_at
		FROM banner
		INNER JOIN bannertag ON bannertag.BannerID = banner.id
	`
	buffer.WriteString(query)
	if feature != "" {
		buffer.WriteString(" WHERE feature = " + feature)
	}
	buffer.WriteString(" GROUP BY id")
	if tag != "" {
		buffer.WriteString(` HAVING ` + tag + ` = ANY(array_agg(bannertag.TagID))`)
	}
	buffer.WriteString(" ORDER BY id")
	if limit != "" {
		buffer.WriteString(" LIMIT " + limit)
	}

	if offset != "" {
		buffer.WriteString(" OFFSET " + offset)
	}
	buffer.WriteString(";")

	rows, err := s.DB.Query(context.Background(), buffer.String())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var banners []models.BannerDB

	for rows.Next() {
		var banner models.BannerDB
		err = rows.Scan(&banner.ID, &banner.Tag, &banner.Feature, &banner.Content, &banner.Access, &banner.Created, &banner.Updated)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		banners = append(banners, banner)
	}
	return banners, nil
}

func (s *Repo) PostBanner(banner *models.BannerPost) (int64, error) {
	const op = "storage.postgres.PostBanner"

	tx, err := s.DB.Begin(context.Background())
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(context.Background())

	execQuery := `INSERT INTO banner(feature, content, access)
			VALUES ($1, $2, $3)
			RETURNING id;`

	row := tx.QueryRow(context.Background(), execQuery, banner.Feature, banner.Content, banner.Access)

	var id int64

	err = row.Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	execQuery = `INSERT INTO bannertag(bannerid, tagid) VALUES ($1, $2)`
	for _, tag := range banner.Tag {
		_, err = tx.Exec(context.Background(), execQuery, id, tag)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	_ = tx.Commit(context.Background())
	return id, nil
}

func (s *Repo) PatchBanner(id string, banner *models.BannerPatch) error {
	const op = "storage.postgres.PatchBanner"

	tx, err := s.DB.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(context.Background())

	updateQuery := `UPDATE banner `
	i := 0
	var buffer bytes.Buffer
	buffer.WriteString(updateQuery)

	if banner.Feature.Defined {
		buffer.WriteString("SET ")
		if banner.Feature.Value == nil {
			buffer.WriteString(`feature = NULL`)
		} else {
			buffer.WriteString(`feature = ` + strconv.Itoa(int(*banner.Feature.Value)))
		}
		i++
	}
	if banner.Content.Defined {
		if i > 0 {
			buffer.WriteString(", ")
		} else {
			buffer.WriteString("SET ")
		}
		if banner.Content.Value == nil {
			buffer.WriteString(`content = NULL`)
		} else {
			b, _ := json.Marshal(banner.Content.Value)
			buffer.WriteString(`content = '` + string(b) + "'")
		}
		i++
	}

	if banner.Access.Defined {
		if i > 0 {
			buffer.WriteString(", ")
		} else {
			buffer.WriteString("SET ")
		}
		if banner.Access.Value == nil {
			buffer.WriteString(`access = NULL`)
		} else {
			buffer.WriteString(`access = ` + strconv.FormatBool(*banner.Access.Value))
		}
	}
	if i > 0 {
		buffer.WriteString(", ")
	} else {
		buffer.WriteString("SET ")
	}
	buffer.WriteString(` updated_at = NOW()`)
	buffer.WriteString(` WHERE id = ` + id + `;`)

	res, err := tx.Exec(context.Background(), buffer.String())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected := res.RowsAffected()

	if rowsAffected == 0 {
		return storage.ErrBannerNotFound
	}

	if banner.Tag.Defined {
		_, err = tx.Exec(context.Background(), "DELETE FROM bannertag WHERE bannerid = "+id)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		execQuery := `INSERT INTO bannertag(bannerid, tagid) VALUES ($1, $2)`
		if banner.Tag.Value == nil {
			_, err = tx.Exec(context.Background(), execQuery, id, nil)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		} else {
			for _, tag := range *banner.Tag.Value {
				_, err = tx.Exec(context.Background(), execQuery, id, tag)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
			}
		}
	}

	_ = tx.Commit(context.Background())

	return nil
}

func (s *Repo) DeleteBanner(id string) error {
	const op = "storage.postgres.DeleteBanner"
	res, err := s.DB.Exec(context.Background(), "DELETE FROM banner WHERE id = "+id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected := res.RowsAffected()

	if rowsAffected == 0 {
		return storage.ErrBannerNotFound
	}

	return err
}
