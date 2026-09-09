DROP TABLE file_versions;

ALTER TABLE files
    DROP COLUMN replaces_file_id;
