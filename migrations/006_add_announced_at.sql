-- 실제 고시/공고 날짜 필드 추가
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS announced_at TIMESTAMP;
