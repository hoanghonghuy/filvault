DO $$
DECLARE
    constraint_name TEXT;
BEGIN
    SELECT con.conname
    INTO constraint_name
    FROM pg_constraint con
    JOIN pg_class rel ON rel.oid = con.conrelid
    JOIN pg_class referenced ON referenced.oid = con.confrelid
    WHERE rel.relname = 'file_upload_cleanup_jobs'
      AND referenced.relname = 'files'
      AND con.contype = 'f'
    LIMIT 1;

    IF constraint_name IS NOT NULL THEN
        EXECUTE format('ALTER TABLE file_upload_cleanup_jobs DROP CONSTRAINT %I', constraint_name);
    END IF;
END $$;
