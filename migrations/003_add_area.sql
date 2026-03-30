-- 면적 변동 정보 저장용 컬럼 추가
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS area_before TEXT;
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS area_after TEXT;
