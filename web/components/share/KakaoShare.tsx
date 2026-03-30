"use client";

interface Props {
  title: string;      // 공고 사업명
  summary: string;    // AI 요약 (미리보기용)
  url: string;        // 공유할 페이지 URL (예: /gangnam/abc-123)
}

/**
 * KakaoShare — 카카오톡 공유 버튼.
 *
 * 카카오톡 공유 URL 방식 사용 (SDK 불필요, 앱 등록 불필요).
 * 클릭하면 카카오톡 앱으로 이동하여 텍스트 공유.
 * 모바일에서는 카카오톡 앱 열림, 데스크탑에서는 카카오톡 웹 열림.
 */
export default function KakaoShare({ title, summary, url }: Props) {
  const handleShare = () => {
    const fullUrl = typeof window !== "undefined"
      ? `${window.location.origin}${url}`
      : url;

    const text = `📌 ${title}\n\n${summary}\n\n자세히 보기 👉 ${fullUrl}`;

    // 카카오톡 공유 URL
    const kakaoUrl = `https://sharer.kakao.com/talk/friends/picker/link?url=${encodeURIComponent(fullUrl)}&text=${encodeURIComponent(text)}`;

    // 모바일이면 카카오톡 앱, 데스크탑이면 새 탭
    window.open(kakaoUrl, "_blank", "noopener,noreferrer");
  };

  // 클립보드 복사 (카카오톡 공유 안 될 때 fallback)
  const handleCopyLink = async () => {
    const fullUrl = typeof window !== "undefined"
      ? `${window.location.origin}${url}`
      : url;

    try {
      await navigator.clipboard.writeText(fullUrl);
      alert("링크가 복사됐어요! 카카오톡에 붙여넣기 해보세요.");
    } catch {
      // clipboard API 안 되는 환경
      prompt("아래 링크를 복사해서 공유하세요:", fullUrl);
    }
  };

  return (
    <div className="flex gap-2">
      <button
        onClick={handleShare}
        className="flex-1 py-2.5 rounded-lg text-sm font-medium bg-yellow-300 text-yellow-900 hover:bg-yellow-400 active:bg-yellow-500 transition-colors flex items-center justify-center gap-1.5"
        aria-label="카카오톡으로 공유"
      >
        <span>💬</span>
        <span>카카오톡 공유</span>
      </button>
      <button
        onClick={handleCopyLink}
        className="py-2.5 px-4 rounded-lg text-sm font-medium bg-gray-100 text-gray-600 hover:bg-gray-200 active:bg-gray-300 transition-colors"
        aria-label="링크 복사"
      >
        🔗
      </button>
    </div>
  );
}
