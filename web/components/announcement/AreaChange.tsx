interface AreaChangeProps {
  before?: string | null; // 기존 면적(㎡), "0"이면 신설
  after?: string | null;  // 변경 후 면적(㎡)
}

// 면적 변동을 시각적으로 표시. 둘 다 없으면 렌더링 안 함.
export default function AreaChange({ before, after }: AreaChangeProps) {
  if (!before && !after) return null;
  if (!after) return null;

  const beforeNum = parseFloat(before || "0");
  const afterNum = parseFloat(after);
  const diff = afterNum - beforeNum;

  // 신설 (기존 면적 0)
  if (beforeNum === 0) {
    return (
      <div className="mt-4">
        <h3 className="text-sm font-semibold text-gray-700 mb-1">면적</h3>
        <p className="text-sm text-gray-600">
          신규 지정 — {formatArea(afterNum)}
        </p>
      </div>
    );
  }

  // 변경
  const diffLabel = diff > 0 ? `+${formatArea(diff)}` : formatArea(diff);
  const diffColor = diff > 0 ? "text-blue-600" : diff < 0 ? "text-red-600" : "text-gray-500";

  return (
    <div className="mt-4">
      <h3 className="text-sm font-semibold text-gray-700 mb-1">면적 변동</h3>
      <p className="text-sm text-gray-600">
        {formatArea(beforeNum)} → {formatArea(afterNum)}
        <span className={`ml-2 font-medium ${diffColor}`}>
          ({diffLabel})
        </span>
      </p>
    </div>
  );
}

function formatArea(n: number): string {
  return n.toLocaleString("ko-KR", { maximumFractionDigits: 1 }) + "㎡";
}
