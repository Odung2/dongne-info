-- AI가 생성한 쉬운 제목 필드 추가
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS easy_title TEXT;
