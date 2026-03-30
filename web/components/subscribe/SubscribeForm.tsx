"use client";

import { useState } from "react";
import { postSubscribe } from "@/lib/api";

interface Props {
  district: string; // 현재 보고 있는 구 (자동 선택)
}

/**
 * SubscribeForm — 인라인 구독 폼.
 *
 * 목록 페이지 하단과 상세 페이지 하단에 배치.
 * 이메일 입력 + 현재 구 자동 선택.
 * useState + HTML5 validation으로 처리.
 */
export default function SubscribeForm({ district }: Props) {
  const [email, setEmail] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (isSubmitting) return;

    setIsSubmitting(true);
    setError("");

    try {
      const ok = await postSubscribe(email, [district]);
      if (ok) {
        setSuccess(true);
        setEmail("");
      } else {
        setError("구독에 실패했어요. 다시 시도해주세요.");
      }
    } catch {
      setError("네트워크 오류가 발생했어요.");
    } finally {
      setIsSubmitting(false);
    }
  };

  if (success) {
    return (
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 text-center">
        <p className="text-blue-700 font-medium">구독 완료!</p>
        <p className="text-sm text-blue-500 mt-1">
          {district} 새 공고가 나오면 알려드릴게요.
        </p>
      </div>
    );
  }

  return (
    <div className="bg-white border border-gray-200 rounded-lg p-4">
      <p className="text-sm font-semibold text-gray-700 mb-2">
        {district} 새 공고 알림 받기
      </p>
      <form onSubmit={handleSubmit} className="flex gap-2">
        <input
          type="email"
          required
          placeholder="이메일 주소"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-blue-400"
          aria-label="이메일 주소 입력"
        />
        <button
          type="submit"
          disabled={isSubmitting}
          className="px-4 py-2 bg-blue-500 text-white rounded-lg text-sm font-medium hover:bg-blue-600 disabled:opacity-50"
        >
          {isSubmitting ? "..." : "구독"}
        </button>
      </form>
      {error && <p className="text-red-500 text-xs mt-2">{error}</p>}
    </div>
  );
}
