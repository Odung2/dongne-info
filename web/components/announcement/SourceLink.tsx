"use client";

import { useState } from "react";

interface Props {
  sourceUrl?: string | null; // 원문 링크 (없을 수 있음)
  rawText?: string | null;   // 도시계획 고시문 전문 (DB에 저장된 원본)
}

/**
 * SourceLink — 원문 보기 컴포넌트.
 *
 * 두 가지 모드:
 *   - rawText 있으면: "원문 보기" 접기/펼치기로 고시문 전문 표시 (도시계획)
 *   - sourceUrl만 있으면: 링크 표시 + 작동 안 될 수 있음 안내 (재개발)
 *   - 둘 다 없으면: 렌더링 안 함
 */
export default function SourceLink({ sourceUrl, rawText }: Props) {
  const [isOpen, setIsOpen] = useState(false);

  if (!sourceUrl && !rawText) return null;

  return (
    <div className="mt-4">
      {/* 도시계획: 고시문 전문 접기/펼치기 */}
      {rawText && rawText.length > 100 && (
        <div>
          <button
            onClick={() => setIsOpen(!isOpen)}
            className="text-sm text-blue-500 hover:text-blue-600 flex items-center gap-1"
            aria-expanded={isOpen}
            aria-label="원문 보기"
          >
            <span>{isOpen ? "▼" : "▶"}</span>
            <span>원문 보기</span>
          </button>
          {isOpen && (
            <div className="mt-2 p-3 bg-gray-50 rounded-lg text-xs text-gray-600 leading-relaxed whitespace-pre-wrap max-h-96 overflow-y-auto">
              {rawText}
            </div>
          )}
        </div>
      )}

      {/* 재개발: 외부 링크 (작동 안 될 수 있음) */}
      {sourceUrl && !rawText && (
        <div className="text-xs text-gray-400">
          <a
            href={sourceUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="text-blue-400 hover:underline"
          >
            원문 보기 (준비 중)
          </a>
          <p className="mt-1">
            링크가 열리지 않으면{" "}
            <a
              href="https://cleanup.seoul.go.kr"
              target="_blank"
              rel="noopener noreferrer"
              className="underline"
            >
              서울시 정비사업 정보몽땅
            </a>
            에서 검색해주세요.
          </p>
        </div>
      )}
    </div>
  );
}
