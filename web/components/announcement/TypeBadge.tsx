"use client";

import { useState } from "react";

interface TypeBadgeProps {
  label: string;              // 뱃지에 표시할 텍스트
  variant: "type" | "action"; // type=유형(재개발/재건축), action=조치(신설/변경/폐지)
}

const TYPE_COLORS: Record<string, string> = {
  "재개발": "bg-blue-100 text-blue-700",
  "재건축": "bg-green-100 text-green-700",
  "도시계획": "bg-purple-100 text-purple-700",
  "도심 재개발": "bg-cyan-100 text-cyan-700",
  "주거환경개선": "bg-orange-100 text-orange-700",
};

const ACTION_COLORS: Record<string, string> = {
  "신설": "bg-emerald-100 text-emerald-700",
  "변경": "bg-amber-100 text-amber-700",
  "폐지": "bg-red-100 text-red-700",
};

const TYPE_TOOLTIPS: Record<string, string> = {
  "재개발": "낡은 주거지역을 밀고 새로 짓는 사업이에요",
  "재건축": "기존 아파트를 허물고 새 아파트를 짓는 사업이에요",
  "도시계획": "도로, 공원, 용도지역 등 도시 계획을 변경하는 거예요",
  "도심 재개발": "낡은 도심 상업지역을 밀고 새로 짓는 사업이에요",
  "주거환경개선": "노후 주거지의 생활환경을 개선하는 사업이에요",
};

const ACTION_TOOLTIPS: Record<string, string> = {
  "신설": "새로운 구역이나 사업을 처음으로 지정한 거예요",
  "변경": "기존에 정해진 계획이나 구역을 수정한 거예요",
  "폐지": "기존 사업이나 구역 지정을 취소/해제한 거예요",
  "결정+지적": "도시계획을 확정하고 공식 지도에 표시한 거예요",
  "결정": "도시계획을 공식적으로 확정한 거예요",
  "실시": "확정된 계획을 실제로 착공할 수 있게 승인한 거예요",
};

// 공고 유형/조치를 색상 뱃지로 표시. 클릭하면 설명 툴팁.
export default function TypeBadge({ label, variant }: TypeBadgeProps) {
  const [showTooltip, setShowTooltip] = useState(false);
  const colors = variant === "type" ? TYPE_COLORS : ACTION_COLORS;
  const tooltips = variant === "type" ? TYPE_TOOLTIPS : ACTION_TOOLTIPS;
  const colorClass = colors[label] || "bg-gray-100 text-gray-600";
  const tooltip = tooltips[label] || "";

  if (!tooltip) {
    return (
      <span className={`inline-block px-2 py-0.5 rounded-full text-xs font-medium ${colorClass}`}>
        {label}
      </span>
    );
  }

  return (
    <span className="relative inline-block">
      <button
        type="button"
        onClick={() => setShowTooltip(!showTooltip)}
        onMouseEnter={() => setShowTooltip(true)}
        onMouseLeave={() => setShowTooltip(false)}
        className={`px-2 py-0.5 rounded-full text-xs font-medium cursor-help ${colorClass}`}
        aria-label={`${label}: ${tooltip}`}
      >
        {label}
      </button>
      {showTooltip && (
        <span className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-3 py-1.5 bg-gray-800 text-white text-xs rounded-lg whitespace-nowrap z-50 shadow-lg">
          {tooltip}
          <span className="absolute top-full left-1/2 -translate-x-1/2 border-4 border-transparent border-t-gray-800" />
        </span>
      )}
    </span>
  );
}
