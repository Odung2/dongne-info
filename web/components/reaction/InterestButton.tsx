"use client";

import { useState, useEffect } from "react";
import { postReaction } from "@/lib/api";

interface Props {
  announcementId: string;
  initialCount: number; // 서버에서 미리 조회한 관심 수
}

/**
 * InterestButton — [관심 있어요] 버튼.
 *
 * 비로그인 유저도 사용 가능 (session_id 기반, 쿠키 저장).
 * 낙관적 업데이트: 클릭 즉시 숫자 +1, 실패 시 롤백.
 * 이미 누른 상태면 비활성화 (localStorage로 체크).
 */
export default function InterestButton({ announcementId, initialCount }: Props) {
  const [count, setCount] = useState(initialCount);
  const [hasReacted, setHasReacted] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  // 이미 반응했는지 localStorage에서 확인
  useEffect(() => {
    const reacted = localStorage.getItem(`interest_${announcementId}`);
    if (reacted) setHasReacted(true);
  }, [announcementId]);

  // session_id 생성/조회
  const getSessionId = (): string => {
    let sid = localStorage.getItem("session_id");
    if (!sid) {
      sid = crypto.randomUUID();
      localStorage.setItem("session_id", sid);
    }
    return sid;
  };

  const handleClick = async () => {
    if (hasReacted || isLoading) return;

    setIsLoading(true);
    setCount((prev) => prev + 1); // 낙관적 업데이트

    try {
      const ok = await postReaction(announcementId, "interest", getSessionId());
      if (ok) {
        setHasReacted(true);
        localStorage.setItem(`interest_${announcementId}`, "true");
      } else {
        setCount((prev) => prev - 1); // 롤백
      }
    } catch {
      setCount((prev) => prev - 1); // 롤백
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <button
      onClick={handleClick}
      disabled={hasReacted || isLoading}
      aria-label={hasReacted ? "이미 관심 표시함" : "관심 있어요"}
      className={`w-full py-3 rounded-lg font-medium text-sm transition-all flex items-center justify-center gap-2
        ${hasReacted
          ? "bg-blue-50 text-blue-500 border border-blue-200 cursor-default"
          : "bg-blue-500 text-white hover:bg-blue-600 active:bg-blue-700 cursor-pointer"
        }
        ${isLoading ? "opacity-70" : ""}
      `}
    >
      <span>{hasReacted ? "✓" : "📌"}</span>
      <span>관심 있어요</span>
      <span className="ml-1">{count}명</span>
    </button>
  );
}
