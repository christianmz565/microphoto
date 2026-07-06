import type { ReactNode } from "react";

export function SlideWrap({
  color,
  tag,
  className,
  variant = "default",
  children,
}: {
  color: string;
  tag: string;
  className?: string;
  variant?: "default" | "decorated";
  children: ReactNode;
}) {
  const isDecorated = variant === "decorated";

  return (
    <div
      className={`size-full text-white flex flex-col justify-between items-center relative overflow-hidden py-4 gap-4`}
    >
      <div
        className="pointer-events-none h-0.75 w-full relative z-10"
        style={{
          background: `linear-gradient(90deg, ${color}, transparent)`,
          ...(isDecorated
            ? { boxShadow: `0 0 10px ${color}40, 0 0 20px ${color}20` }
            : {}),
        }}
      />
      {isDecorated && (
        <>
          <div
            className="absolute top-5 left-5 w-2 h-2 rounded-full opacity-20"
            style={{ backgroundColor: color }}
          />
          <div
            className="absolute top-5 right-5 w-2 h-2 rounded-full opacity-20"
            style={{ backgroundColor: color }}
          />
          <div
            className="absolute bottom-20 left-5 w-2 h-2 rounded-full opacity-20"
            style={{ backgroundColor: color }}
          />
          <div
            className="absolute bottom-20 right-5 w-2 h-2 rounded-full opacity-20"
            style={{ backgroundColor: color }}
          />
        </>
      )}

      <div className="flex flex-col flex-1 w-full">
        <div
          className="mb-2 inline-block text-lg font-semibold uppercase tracking-[0.18em] relative z-10"
          style={{ color }}
        >
          {tag}
        </div>
        <div className={`flex-1 ${className} relative z-10`}>{children}</div>
      </div>
      <div
        className="pointer-events-none h-0.75 w-full relative z-10"
        style={{
          background: `linear-gradient(90deg, ${color}, transparent)`,
          ...(isDecorated
            ? { boxShadow: `0 0 10px ${color}40, 0 0 20px ${color}20` }
            : {}),
        }}
      />
    </div>
  );
}
