-- 009_quota_notification.sql
-- Mengubah idempotency key dari UNIQUE (outbox_id) menjadi UNIQUE (outbox_id, user_id)
-- Alasan: Event broadcast seperti 'course.quota_changed' memiliki satu baris outbox_id
-- tetapi harus dikirimkan ke SELURUH mahasiswa peserta kelas tersebut.
-- Dengan UNIQUE (outbox_id, user_id), setiap mahasiswa memiliki claim idempoten masing-masing.

ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_outbox_id_key;
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_outbox_unique;
ALTER TABLE notifications ADD CONSTRAINT notifications_outbox_user_id_key UNIQUE (outbox_id, user_id);
