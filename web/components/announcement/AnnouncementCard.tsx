import Link from "next/link";
import { Announcement, getDistrictSlug } from "@/types";
import TypeBadge from "./TypeBadge";

interface Props {
  announcement: Announcement;
  interestCount?: number; // 관심 수 (서버에서 미리 조회)
}

// 공고 카드 — 목록 페이지에서 사용. 클릭하면 상세로 이동.
// easy_title이 있으면 쉬운 제목을 표시하고, impact를 부제로 보여준다.
export default function AnnouncementCard({ announcement: a, interestCount }: Props) {
  const slug = getDistrictSlug(a.district) || "gangnam";
  const displayTitle = a.easy_title || a.title;
  const summaryPreview = a.summary
    ? a.summary.length > 80 ? a.summary.slice(0, 80) + "..." : a.summary
    : null;

  return (
    <Link href={`/${slug}/${a.id}`} className="block">
      <article className="bg-white rounded-lg border border-gray-200 p-4 hover:border-blue-300 hover:shadow-sm transition-all">
        {/* 뱃지 */}
        <div className="flex gap-2 mb-2">
          <TypeBadge label={a.type} variant="type" />
          <TypeBadge label={a.action} variant="action" />
        </div>

        {/* 쉬운 제목 */}
        <h3 className="font-semibold text-gray-900 mb-1 leading-snug">
          {displayTitle}
        </h3>

        {/* 부제: impact (나한테 어떤 의미?) */}
        {a.impact && (
          <p className="text-sm text-blue-600 mb-2">{a.impact}</p>
        )}

        {/* AI 요약 미리보기 (impact 없을 때만) */}
        {!a.impact && summaryPreview && (
          <p className="text-sm text-gray-600 mb-3 leading-relaxed">
            {summaryPreview}
          </p>
        )}

        {/* 하단: 날짜 + 단계 + 관심 수 */}
        <div className="flex items-center justify-between text-xs text-gray-400">
          <div className="flex items-center gap-2">
            <time dateTime={a.created_at}>
              {new Date(a.created_at).toLocaleDateString("ko-KR")}
            </time>
            {a.stage && <span>· {a.stage}</span>}
          </div>
          {interestCount !== undefined && interestCount > 0 && (
            <span>관심 {interestCount}명</span>
          )}
        </div>
      </article>
    </Link>
  );
}
