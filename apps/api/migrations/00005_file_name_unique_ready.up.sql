DROP INDEX files_sibling_name;
DROP INDEX files_root_name;

CREATE UNIQUE INDEX files_sibling_name
    ON files (owner_id, folder_id, name)
    WHERE deleted_at IS NULL AND folder_id IS NOT NULL AND status = 'READY';

CREATE UNIQUE INDEX files_root_name
    ON files (owner_id, name)
    WHERE deleted_at IS NULL AND folder_id IS NULL AND status = 'READY';
