-- 원본 API 분류 정보 저장용 컬럼 추가
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS raw_category TEXT;
