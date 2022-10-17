-- Undo the changes made in the up migration
ALTER TABLE lsif_dependency_repos
DROP COLUMN last_sync_error;
