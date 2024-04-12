-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS feature(
    id SERIAL PRIMARY KEY CHECK(0 < id)
);

CREATE TABLE IF NOT EXISTS tag(
    id SERIAL PRIMARY KEY CHECK(0 < id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS banner(
    id INT GENERATED ALWAYS AS IDENTITY,
    feature INT,
    content JSONB,
    access BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(id),
    CONSTRAINT fk_feature
        FOREIGN KEY(feature)
            REFERENCES feature(id)
            ON DELETE SET NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS bannerTag (
    BannerID int REFERENCES banner ON DELETE CASCADE,
    TagID int REFERENCES tag ON DELETE SET NULL,
    UNIQUE(BannerID, TagID)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bannerTag;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS banner;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS tag;
DROP TABLE IF EXISTS feature;
-- +goose StatementEnd
