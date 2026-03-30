-- "나한테 어떤 의미?" + "추천 액션" 필드 추가
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS impact TEXT;
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS action_tip TEXT;
