-- 동네정보 초기 스키마
-- 실행: Supabase SQL Editor에서 실행

-- 공고
CREATE TABLE announcements (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  district    VARCHAR(20) NOT NULL,
  type        VARCHAR(20) NOT NULL,
  action      VARCHAR(10) NOT NULL,
  title       TEXT NOT NULL,
  location    TEXT,
  summary     TEXT,
  stage       VARCHAR(50),
  related     TEXT,
  source_url  TEXT,
  source_id   VARCHAR(100) UNIQUE,
  created_at  TIMESTAMP DEFAULT NOW(),
  notified_at TIMESTAMP
);

-- 구독자
CREATE TABLE subscribers (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  contact     VARCHAR(100) NOT NULL,
  type        VARCHAR(10) NOT NULL,
  districts   TEXT[],
  created_at  TIMESTAMP DEFAULT NOW(),
  active      BOOLEAN DEFAULT TRUE
);

-- 관심 표시 / 찬반
CREATE TABLE reactions (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  announcement_id UUID REFERENCES announcements(id),
  type            VARCHAR(10) NOT NULL,
  session_id      VARCHAR(100),
  subscriber_id   UUID REFERENCES subscribers(id),
  created_at      TIMESTAMP DEFAULT NOW(),
  UNIQUE(announcement_id, session_id, type)
);

-- 주민 의견
CREATE TABLE comments (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  announcement_id UUID REFERENCES announcements(id),
  subscriber_id   UUID REFERENCES subscribers(id),
  body            TEXT NOT NULL,
  created_at      TIMESTAMP DEFAULT NOW()
);

-- 알림 이력
CREATE TABLE notification_logs (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  announcement_id UUID REFERENCES announcements(id),
  subscriber_id   UUID REFERENCES subscribers(id),
  sent_at         TIMESTAMP DEFAULT NOW(),
  status          VARCHAR(10)
);

-- 인덱스
CREATE INDEX idx_announcements_district ON announcements(district);
CREATE INDEX idx_announcements_type ON announcements(type);
CREATE INDEX idx_announcements_created_at ON announcements(created_at DESC);
CREATE INDEX idx_reactions_announcement_id ON reactions(announcement_id);
CREATE INDEX idx_comments_announcement_id ON comments(announcement_id);
CREATE INDEX idx_subscribers_active ON subscribers(active) WHERE active = TRUE;
