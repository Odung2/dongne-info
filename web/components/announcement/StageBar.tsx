/**
 * StageBar — 재건축 7단계 진행 상황을 프로그레스 바로 표시.
 *
 * 조건부 렌더링:
 *   - "2/7단계 - 정비구역지정" → 프로그레스 바 (2/7 채움) + 텍스트
 *   - "확인 필요" 등 → 텍스트만
 *   - null/undefined → 렌더링 안 함 (return null)
 *
 * @param stage - "2/7단계 - 정비구역지정" 형태 또는 null
 */

const STAGE_LABELS = [
  "기본계획",
  "정비구역지정",
  "조합설립",
  "사업시행인가",
  "관리처분인가",
  "착공",
  "입주",
];

export default function StageBar({ stage }: { stage?: string | null }) {
  if (!stage) return null;

  // "2/7단계" 패턴에서 숫자 추출
  const match = stage.match(/(\d+)\/7단계/);
  const current = match ? parseInt(match[1]) : null;

  if (current === null) {
    // 단계 판별 불가 — 텍스트만 표시
    return (
      <div className="mt-4">
        <h3 className="text-sm font-semibold text-gray-700 mb-1">진행 단계</h3>
        <p className="text-sm text-gray-500">{stage}</p>
      </div>
    );
  }

  return (
    <div className="mt-4">
      <h3 className="text-sm font-semibold text-gray-700 mb-2">진행 단계</h3>

      {/* 프로그레스 바 */}
      <div className="flex gap-1 mb-2">
        {STAGE_LABELS.map((label, i) => (
          <div
            key={label}
            className={`h-2 flex-1 rounded-full ${
              i < current ? "bg-blue-500" : "bg-gray-200"
            }`}
            title={label}
          />
        ))}
      </div>

      {/* 현재 단계 텍스트 */}
      <p className="text-sm text-blue-600 font-medium">
        {current}/7단계 — {STAGE_LABELS[current - 1] || ""}
      </p>
    </div>
  );
}
