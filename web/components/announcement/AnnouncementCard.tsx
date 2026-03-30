import Link from "next/link";
import { Announcement, getDistrictSlug } from "@/types";
import TypeBadge from "./TypeBadge";

interface Props {
  announcement: Announcement;
  interestCount?: number; // 관심 수 (서버에서 미리 조회)
}

// 공고 카드 — 목록 페이지에서 사용. 클릭하면 상세로 이동.
export default function AnnouncementCard({ announcement: a, interestCount }: Props) {
  const slug = getDistrictSlug(a.district) || "gangnam";
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

        {/* 사업명 */}
        <h3 className="font-semibold text-gray-900 mb-1 leading-snug">
          {a.title}
        </h3>

        {/* 위치 */}
        {a.location && (
          <p className="text-sm text-gray-500 mb-2">{a.location}</p>
        )}

        {/* AI 요약 미리보기 */}
        {summaryPreview && (
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
