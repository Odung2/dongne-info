interface TypeBadgeProps {
  label: string;              // 뱃지에 표시할 텍스트
  variant: "type" | "action"; // type=유형(재개발/재건축), action=조치(신설/변경/폐지)
}

const TYPE_COLORS: Record<string, string> = {
  "재개발": "bg-blue-100 text-blue-700",
  "재건축": "bg-green-100 text-green-700",
  "도시계획": "bg-purple-100 text-purple-700",
  "정비사업": "bg-gray-100 text-gray-700",
};

const ACTION_COLORS: Record<string, string> = {
  "신설": "bg-emerald-100 text-emerald-700",
  "변경": "bg-amber-100 text-amber-700",
  "폐지": "bg-red-100 text-red-700",
};

// 공고 유형/조치를 색상 뱃지로 표시
export default function TypeBadge({ label, variant }: TypeBadgeProps) {
  const colors = variant === "type" ? TYPE_COLORS : ACTION_COLORS;
  const colorClass = colors[label] || "bg-gray-100 text-gray-600";

  return (
    <span className={`inline-block px-2 py-0.5 rounded-full text-xs font-medium ${colorClass}`}>
      {label}
    </span>
  );
}
