ALTER TABLE albums
    ADD COLUMN cover_file_id CHAR(26) NULL REFERENCES files(id);
