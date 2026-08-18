ALTER TABLE users
    ADD COLUMN image_thumbnails_enabled BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN video_thumbnails_enabled BOOLEAN NOT NULL DEFAULT true;
