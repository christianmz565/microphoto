import type { ReactNode } from "react";

export function InfoBox({
  title,
  color,
  children,
}: {
  title: string;
  color: string;
  children: ReactNode;
}) {
  return (
    <div className="rounded-xl border border-white/10 bg-white/5 p-5">
      <p
        className="text-lg font-bold uppercase tracking-widest mb-2"
        style={{ color }}
      >
        {title}
      </p>
      {children}
    </div>
  );
}
