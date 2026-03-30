"use client";

import { useState } from "react";
import { Announcement } from "@/types";
import AnnouncementCard from "./AnnouncementCard";

interface Props {
  announcements: Announcement[];
}

const FILTER_TABS = ["전체", "재개발", "재건축", "도심 재개발", "도시계획"];

/**
 * AnnouncementList — 공고 목록 + 유형 필터 탭.
 *
 * 클라이언트 컴포넌트 (필터 탭 인터랙션).
 * 서버에서 전체 데이터를 받아서 클라이언트에서 필터링.
 * 데이터가 100건 미만이라 클라이언트 필터링으로 충분.
 */
export default function AnnouncementList({ announcements }: Props) {
  const [filter, setFilter] = useState("전체");

  const filtered = filter === "전체"
    ? announcements
    : announcements.filter((a) => a.type === filter);

  return (
    <div>
      {/* 필터 탭 */}
      <div className="flex gap-2 mb-4 overflow-x-auto" role="tablist">
        {FILTER_TABS.map((tab) => (
          <button
            key={tab}
            role="tab"
            aria-selected={filter === tab}
            onClick={() => setFilter(tab)}
            className={`px-3 py-1.5 rounded-full text-sm whitespace-nowrap transition-colors
              ${filter === tab
                ? "bg-blue-500 text-white"
                : "bg-gray-100 text-gray-600 hover:bg-gray-200"
              }`}
          >
            {tab}
            <span className="ml-1 text-xs opacity-70">
              {tab === "전체"
                ? announcements.length
                : announcements.filter((a) => a.type === tab).length}
            </span>
          </button>
        ))}
      </div>

      {/* 공고 목록 */}
      <div className="space-y-3">
        {filtered.length > 0 ? (
          filtered.map((a) => (
            <AnnouncementCard key={a.id} announcement={a} />
          ))
        ) : (
          <p className="text-center text-gray-400 py-8">
            해당하는 공고가 없어요.
          </p>
        )}
      </div>
    </div>
  );
}
