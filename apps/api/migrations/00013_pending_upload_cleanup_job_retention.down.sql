ALTER TABLE file_upload_cleanup_jobs
    ADD CONSTRAINT file_upload_cleanup_jobs_file_id_fkey
    FOREIGN KEY (file_id) REFERENCES files (id) ON DELETE CASCADE;
