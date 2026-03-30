import { Announcement, ReactionSummary } from "@/types";

/**
 * Go API 호출 유틸.
 *
 * 서버 컴포넌트: GO_API_URL (비노출, 서버→서버 직접)
 * 클라이언트 컴포넌트: NEXT_PUBLIC_API_URL (브라우저→Go API)
 */
const SERVER_API = process.env.GO_API_URL || "http://localhost:8080";
const CLIENT_API =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

/** 서버 컴포넌트용 — 공고 목록 조회 */
export async function fetchAnnouncements(
  district: string,
  type?: string
): Promise<Announcement[]> {
  const params = new URLSearchParams({ district });
  if (type && type !== "전체") params.set("type", type);

  const res = await fetch(`${SERVER_API}/api/announcements?${params}`, {
    next: { revalidate: 300 }, // 5분 캐시
  });

  if (!res.ok) return [];
  const data = await res.json();
  return data.data ?? [];
}

/** 서버 컴포넌트용 — 공고 상세 조회 */
export async function fetchAnnouncement(
  id: string
): Promise<Announcement | null> {
  const res = await fetch(`${SERVER_API}/api/announcements/${id}`, {
    next: { revalidate: 300 },
  });

  if (!res.ok) return null;
  const data = await res.json();
  return data.data ?? null;
}

/** 서버 컴포넌트용 — 반응 집계 조회 */
export async function fetchReactions(
  id: string
): Promise<ReactionSummary> {
  const res = await fetch(`${SERVER_API}/api/announcements/${id}/reactions`, {
    next: { revalidate: 60 }, // 1분 캐시
  });

  if (!res.ok) return { interest: 0, agree: 0, disagree: 0 };
  const data = await res.json();
  return data.data ?? { interest: 0, agree: 0, disagree: 0 };
}

/** 클라이언트용 — 관심 표시 */
export async function postReaction(
  announcementId: string,
  type: string,
  sessionId: string
): Promise<boolean> {
  const res = await fetch(
    `${CLIENT_API}/api/announcements/${announcementId}/react`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ type, session_id: sessionId }),
    }
  );
  return res.ok;
}

/** 클라이언트용 — 구독 등록 */
export async function postSubscribe(
  contact: string,
  districts: string[]
): Promise<boolean> {
  const res = await fetch(`${CLIENT_API}/api/subscribe`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ contact, type: "email", districts }),
  });
  return res.ok;
}
