/**
 * 마포구 공고 상세 페이지 (/mapo/[id])
 *
 * 데이터: GET /api/announcements/:id + GET /api/announcements/:id/reactions
 * 서버 컴포넌트 (SSR, SEO)
 */
import { notFound } from "next/navigation";
import type { Metadata } from "next";
import Link from "next/link";
import { fetchAnnouncement, fetchReactions } from "@/lib/api";
import TypeBadge from "@/components/announcement/TypeBadge";
import StageBar from "@/components/announcement/StageBar";
import AreaChange from "@/components/announcement/AreaChange";
import InterestButton from "@/components/reaction/InterestButton";
import SubscribeForm from "@/components/subscribe/SubscribeForm";

interface PageProps {
  params: Promise<{ id: string }>;
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { id } = await params;
  const a = await fetchAnnouncement(id);
  if (!a) return { title: "공고를 찾을 수 없습니다" };

  return {
    title: `${a.title} | 마포구 공고 — 동네정보`,
    description: a.summary || `${a.district} ${a.title}`,
  };
}

export default async function MapoDetailPage({ params }: PageProps) {
  const { id } = await params;
  const [announcement, reactions] = await Promise.all([
    fetchAnnouncement(id),
    fetchReactions(id),
  ]);

  if (!announcement) notFound();
  const a = announcement;

  return (
    <div>
      <Link
        href="/mapo"
        className="text-sm text-gray-400 hover:text-gray-600 mb-4 inline-block"
      >
        ← 마포구 공고 목록
      </Link>

      <div className="flex gap-2 mb-3">
        <TypeBadge label={a.type} variant="type" />
        <TypeBadge label={a.action} variant="action" />
      </div>

      <h1 className="text-xl font-bold text-gray-900 mb-1">{a.title}</h1>
      {a.location && (
        <p className="text-sm text-gray-500 mb-4">{a.district} {a.location}</p>
      )}

      <hr className="my-4 border-gray-200" />

      {a.summary && (
        <section className="mb-4">
          <h2 className="text-sm font-semibold text-gray-700 mb-2">쉬운 설명</h2>
          <p className="text-gray-700 leading-relaxed">{a.summary}</p>
        </section>
      )}

      <StageBar stage={a.stage} />
      <AreaChange before={a.area_before} after={a.area_after} />

      <hr className="my-4 border-gray-200" />

      <InterestButton
        announcementId={a.id}
        initialCount={reactions.interest}
      />

      <div className="mt-4 space-y-2 text-xs text-gray-400">
        {a.raw_category && <p>분류: {a.raw_category}</p>}
        {a.source_url && (
          <a
            href={a.source_url}
            target="_blank"
            rel="noopener noreferrer"
            className="text-blue-400 hover:underline block"
          >
            원문 보기 →
          </a>
        )}
      </div>

      <div className="mt-6">
        <SubscribeForm district="마포구" />
      </div>
    </div>
  );
}
