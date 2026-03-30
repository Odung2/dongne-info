import Link from "next/link";

// 상단 네비게이션 — 로고 + 서비스명
export default function Header() {
  return (
    <header className="bg-white border-b border-gray-200">
      <div className="max-w-2xl mx-auto px-4 py-3 flex items-center justify-between">
        <Link href="/" className="text-lg font-bold text-blue-600">
          동깨
        </Link>
        <p className="text-xs text-gray-400">우리 동네 소식, 쉽게</p>
      </div>
    </header>
  );
}
