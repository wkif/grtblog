-- +goose Up

ALTER TABLE photo
    ADD COLUMN media_type  VARCHAR(16) NOT NULL DEFAULT 'image',
    ADD COLUMN mime_type   VARCHAR(127),
    ADD COLUMN poster_url  TEXT,
    ADD COLUMN duration_ms BIGINT,
    ADD COLUMN width       INT,
    ADD COLUMN height      INT;

ALTER TABLE photo
    ADD CONSTRAINT ck_photo_media_type CHECK (media_type IN ('image', 'video')),
    ADD CONSTRAINT ck_photo_duration_ms CHECK (duration_ms IS NULL OR duration_ms >= 0),
    ADD CONSTRAINT ck_photo_width CHECK (width IS NULL OR width > 0),
    ADD CONSTRAINT ck_photo_height CHECK (height IS NULL OR height > 0);

CREATE INDEX idx_photo_album_media_type ON photo (album_id, media_type) WHERE deleted_at IS NULL;

-- +goose Down

DROP INDEX IF EXISTS idx_photo_album_media_type;

ALTER TABLE photo
    DROP CONSTRAINT IF EXISTS ck_photo_height,
    DROP CONSTRAINT IF EXISTS ck_photo_width,
    DROP CONSTRAINT IF EXISTS ck_photo_duration_ms,
    DROP CONSTRAINT IF EXISTS ck_photo_media_type;

ALTER TABLE photo
    DROP COLUMN IF EXISTS height,
    DROP COLUMN IF EXISTS width,
    DROP COLUMN IF EXISTS duration_ms,
    DROP COLUMN IF EXISTS poster_url,
    DROP COLUMN IF EXISTS mime_type,
    DROP COLUMN IF EXISTS media_type;
