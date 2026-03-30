/**
 * 랜딩 페이지 (/)
 *
 * 구 선택 + 최근 인기 공고 미리보기.
 * 서버 컴포넌트 (SSR, SEO용).
 * 데이터: GET /api/announcements (강남구, 마포구 각각)
 */
import Link from "next/link";
import { DISTRICTS } from "@/types";
import { fetchAnnouncements } from "@/lib/api";
import AnnouncementCard from "@/components/announcement/AnnouncementCard";

export default async function HomePage() {
  // 각 구의 공고 수 + 최근 3건
  const gangnam = await fetchAnnouncements("강남구");
  const mapo = await fetchAnnouncements("마포구");
  const recent = [...gangnam, ...mapo]
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 3);

  return (
    <div>
      {/* 서비스 소개 */}
      <section className="text-center mb-8">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">
          우리 동네 소식, 곰고가 알려줘요 🐻
        </h1>
        <p className="text-gray-500 text-sm">
          재개발 · 재건축 · 도시계획 소식을 곰고가 쉽게 알려줘요
        </p>
      </section>

      {/* 구 선택 */}
      <section className="grid grid-cols-2 gap-3 mb-8">
        {DISTRICTS.map((d) => {
          const count = d.name === "강남구" ? gangnam.length : mapo.length;
          return (
            <Link
              key={d.slug}
              href={`/${d.slug}`}
              className="bg-white border border-gray-200 rounded-lg p-4 text-center hover:border-blue-300 hover:shadow-sm transition-all"
            >
              <p className="font-bold text-lg text-gray-900">{d.name}</p>
              <p className="text-sm text-gray-400 mt-1">공고 {count}건</p>
            </Link>
          );
        })}
      </section>

      {/* 최근 공고 미리보기 */}
      {recent.length > 0 && (
        <section>
          <h2 className="text-sm font-semibold text-gray-500 mb-3">
            최근 공고
          </h2>
          <div className="space-y-3">
            {recent.map((a) => (
              <AnnouncementCard key={a.id} announcement={a} />
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
