"use client";

import { useState } from "react";

/**
 * StageBar — 재건축/도시계획 진행 상황을 프로그레스 바로 표시.
 *
 * 각 단계 칸을 클릭/호버하면 해당 단계 설명이 말풍선으로 표시된다.
 */

interface StageInfo {
  label: string;
  tooltip: string;
}

const REBUILD_STAGES: StageInfo[] = [
  { label: "기본계획", tooltip: "시/구에서 어디를 재건축할지 전체 계획을 세우는 단계예요" },
  { label: "정비구역지정", tooltip: "이 구역에서 재건축을 공식적으로 할 수 있게 지정하는 단계예요" },
  { label: "조합설립", tooltip: "주민들이 재건축 조합을 만들어서 사업 주체가 되는 단계예요" },
  { label: "사업시행인가", tooltip: "구체적인 설계와 사업 계획을 정부에서 승인하는 단계예요" },
  { label: "관리처분인가", tooltip: "기존 주민의 권리(분양권 등)를 정리하는 단계예요" },
  { label: "착공", tooltip: "실제로 기존 건물을 철거하고 새 건물을 짓기 시작하는 단계예요" },
  { label: "입주", tooltip: "새 건물이 완성되어 주민들이 들어가는 단계예요" },
];

const CITYPLAN_STAGES: StageInfo[] = [
  { label: "입안", tooltip: "도시계획 초안을 작성하는 단계예요" },
  { label: "주민열람", tooltip: "주민들이 계획안을 보고 의견을 낼 수 있는 단계예요" },
  { label: "위원회심의", tooltip: "전문가 위원회에서 계획을 검토하고 승인하는 단계예요" },
  { label: "결정고시", tooltip: "계획이 최종 확정되어 공식 발표되는 단계예요" },
  { label: "지형도면고시", tooltip: "확정된 계획을 공식 지도에 표시하는 마무리 단계예요" },
];

export default function StageBar({ stage }: { stage?: string | null }) {
  const [activeIdx, setActiveIdx] = useState<number | null>(null);

  if (!stage) return null;

  if (stage === "폐지됨") {
    return (
      <div className="mt-4">
        <h3 className="text-sm font-semibold text-gray-700 mb-1">진행 단계</h3>
        <p className="text-sm text-red-500 font-medium">사업 폐지됨</p>
      </div>
    );
  }

  const match7 = stage.match(/(\d+)\/7단계/);
  const match5 = stage.match(/(\d+)\/5단계/);
  const current = match7 ? parseInt(match7[1]) : match5 ? parseInt(match5[1]) : null;
  const stages = match7 ? REBUILD_STAGES : match5 ? CITYPLAN_STAGES : null;
  const total = match7 ? 7 : match5 ? 5 : null;

  if (current === null || stages === null || total === null) {
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
        {stages.map((s, i) => (
          <div key={s.label} className="relative flex-1">
            <button
              type="button"
              onClick={() => setActiveIdx(activeIdx === i ? null : i)}
              onMouseEnter={() => setActiveIdx(i)}
              onMouseLeave={() => setActiveIdx(null)}
              className={`w-full h-3 rounded-full cursor-help transition-all ${
                i < current
                  ? "bg-blue-500 hover:bg-blue-600"
                  : i === current - 1
                  ? "bg-blue-500 hover:bg-blue-600"
                  : "bg-gray-200 hover:bg-gray-300"
              }`}
              aria-label={`${i + 1}/${total}단계 ${s.label}`}
            />
            {/* 툴팁 말풍선 */}
            {activeIdx === i && (
              <span className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-3 py-2 bg-gray-800 text-white text-xs rounded-lg w-48 text-center z-50 shadow-lg">
                <span className="font-medium">{i + 1}단계: {s.label}</span>
                <br />
                <span className="opacity-80">{s.tooltip}</span>
                <span className="absolute top-full left-1/2 -translate-x-1/2 border-4 border-transparent border-t-gray-800" />
              </span>
            )}
          </div>
        ))}
      </div>

      {/* 현재 단계 텍스트 */}
      <p className="text-sm text-blue-600 font-medium">
        {current}/{total}단계 — {stages[current - 1]?.label || ""}
      </p>
      <p className="text-xs text-gray-400 mt-0.5">
        {stages[current - 1]?.tooltip || ""}
      </p>
    </div>
  );
}
