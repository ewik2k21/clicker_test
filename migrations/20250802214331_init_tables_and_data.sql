-- +goose Up
-- +goose StatementBegin
CREATE TABLE banners (
                         id UUID PRIMARY KEY,
                         name VARCHAR(255) NOT NULL
);

-- create table
CREATE TABLE clicks (
                        timestamp TIMESTAMP NOT NULL,
                        banner_id UUID REFERENCES banners(id),
                        count INT NOT NULL,
                        PRIMARY KEY (timestamp, banner_id)
);

-- index for clicks
CREATE INDEX idx_clicks_banner_id_timestamp ON clicks(banner_id, timestamp);

-- 100 banners
DO $$
    BEGIN
        FOR i IN 1..100 LOOP
                INSERT INTO banners (id, name) VALUES (gen_random_uuid(), 'Banner ' || i);
            END LOOP;
    END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_clicks_banner_id_timestamp;
DROP TABLE clicks;
DROP TABLE banners;
-- +goose StatementEnd
