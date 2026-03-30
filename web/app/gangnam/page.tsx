/**
 * 강남구 공고 목록 페이지 (/gangnam)
 *
 * 데이터: GET /api/announcements?district=강남구
 * 서버 컴포넌트 (SSR, SEO용)
 */
import type { Metadata } from "next";
import { fetchAnnouncements } from "@/lib/api";
import AnnouncementList from "@/components/announcement/AnnouncementList";
import SubscribeForm from "@/components/subscribe/SubscribeForm";

export const metadata: Metadata = {
  title: "강남구 공고 — 동네정보",
  description: "강남구 재개발·재건축·도시계획 공고를 AI가 쉽게 설명해드려요.",
};

export default async function GangnamPage() {
  const announcements = await fetchAnnouncements("강남구");

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900 mb-4">강남구 공고</h1>
      <AnnouncementList announcements={announcements} />
      <div className="mt-6">
        <SubscribeForm district="강남구" />
      </div>
    </div>
  );
}
