/**
 * 마포구 공고 목록 페이지 (/mapo)
 *
 * 데이터: GET /api/announcements?district=마포구
 * 서버 컴포넌트 (SSR, SEO용)
 */
import type { Metadata } from "next";
import { fetchAnnouncements } from "@/lib/api";
import AnnouncementList from "@/components/announcement/AnnouncementList";
import SubscribeForm from "@/components/subscribe/SubscribeForm";

export const metadata: Metadata = {
  title: "마포구 공고 — 동네깨비",
  description: "마포구 재개발·재건축·도시계획 소식을 동깨가 쉽게 알려줘요.",
};

export default async function MapoPage() {
  const announcements = await fetchAnnouncements("마포구");

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900 mb-4">마포구 공고</h1>
      <AnnouncementList announcements={announcements} />
      <div className="mt-6">
        <SubscribeForm district="마포구" />
      </div>
    </div>
  );
}
