/**
 * 강남구 공고 상세 페이지 (/gangnam/[id])
 *
 * 데이터: GET /api/announcements/:id + GET /api/announcements/:id/reactions
 * 서버 컴포넌트 (SSR, SEO — 공고 상세는 구글 검색 유입의 핵심)
 */
import { notFound } from "next/navigation";
import type { Metadata } from "next";
import Link from "next/link";
import { fetchAnnouncement, fetchReactions } from "@/lib/api";
import TypeBadge from "@/components/announcement/TypeBadge";
import SourceLink from "@/components/announcement/SourceLink";
import StageBar from "@/components/announcement/StageBar";
import AreaChange from "@/components/announcement/AreaChange";
import InterestButton from "@/components/reaction/InterestButton";
import KakaoShare from "@/components/share/KakaoShare";
import SubscribeForm from "@/components/subscribe/SubscribeForm";

interface PageProps {
  params: Promise<{ id: string }>;
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { id } = await params;
  const a = await fetchAnnouncement(id);
  if (!a) return { title: "공고를 찾을 수 없습니다" };

  return {
    title: `${a.title} | 강남구 공고 — 곰고`,
    description: a.summary || `${a.district} ${a.title}`,
  };
}

export default async function GangnamDetailPage({ params }: PageProps) {
  const { id } = await params;
  const [announcement, reactions] = await Promise.all([
    fetchAnnouncement(id),
    fetchReactions(id),
  ]);

  if (!announcement) notFound();
  const a = announcement;

  return (
    <div>
      {/* 뒤로가기 */}
      <Link
        href="/gangnam"
        className="text-sm text-gray-400 hover:text-gray-600 mb-4 inline-block"
      >
        ← 강남구 공고 목록
      </Link>

      {/* 뱃지 */}
      <div className="flex gap-2 mb-3">
        <TypeBadge label={a.type} variant="type" />
        <TypeBadge label={a.action} variant="action" />
      </div>

      {/* 쉬운 제목 + 원본 고시명 */}
      <h1 className="text-xl font-bold text-gray-900 mb-1">
        {a.easy_title || a.title}
      </h1>
      {a.easy_title && (
        <p className="text-xs text-gray-400 mb-1">원본: {a.title}</p>
      )}
      {a.location && (
        <p className="text-sm text-gray-500 mb-4">{a.district} {a.location}</p>
      )}

      <hr className="my-4 border-gray-200" />

      {/* AI 쉬운 설명 */}
      {a.summary && (
        <section className="mb-4">
          <h2 className="text-sm font-semibold text-gray-700 mb-2">쉬운 설명</h2>
          <p className="text-gray-700 leading-relaxed">{a.summary}</p>
        </section>
      )}

      {/* 나한테 어떤 의미? + 추천 액션 */}
      {(a.impact || a.action_tip) && (
        <section className="mb-4 bg-blue-50 rounded-lg p-4">
          {a.impact && (
            <div className="mb-2">
              <h3 className="text-sm font-semibold text-blue-700 mb-1">나한테 어떤 의미?</h3>
              <p className="text-sm text-blue-600">{a.impact}</p>
            </div>
          )}
          {a.action_tip && (
            <div>
              <h3 className="text-sm font-semibold text-blue-700 mb-1">이런 걸 해볼 수 있어요</h3>
              <p className="text-sm text-blue-600">{a.action_tip}</p>
            </div>
          )}
        </section>
      )}

      {/* 진행 단계 */}
      <StageBar stage={a.stage} />

      {/* 면적 변동 */}
      <AreaChange before={a.area_before} after={a.area_after} />

      <hr className="my-4 border-gray-200" />

      {/* 관심 버튼 + 카카오 공유 */}
      <div className="space-y-3">
        <InterestButton
          announcementId={a.id}
          initialCount={reactions.interest}
        />
        <KakaoShare
          title={a.title}
          summary={a.summary || ""}
          url={`/gangnam/${a.id}`}
        />
      </div>

      {/* 원문/분류 */}
      <div className="mt-4 space-y-2 text-xs text-gray-400">
        {a.raw_category && <p>분류: {a.raw_category}</p>}
      </div>
      <SourceLink sourceUrl={a.source_url} />

      {/* 구독 */}
      <div className="mt-6">
        <SubscribeForm district="강남구" />
      </div>
    </div>
  );
}
