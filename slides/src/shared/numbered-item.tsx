export function NumberedItem({
  num,
  title,
  description,
  color,
}: {
  num: string;
  title: string;
  description: string;
  color: string;
}) {
  return (
    <div className="rounded-xl border border-white/10 bg-white/5 p-3">
      <div className="flex items-center gap-2">
        <span
          className="flex size-6 shrink-0 items-center justify-center rounded-full text-lg font-semibold"
          style={{ backgroundColor: `${color}25`, color }}
        >
          {num}
        </span>
        <p className="m-3! text-2xl font-semibold uppercase tracking-widest text-white">
          {title}
        </p>
      </div>
      <p className="m-3! text-2xl text-left text-gray-400 leading-relaxed">
        {description}
      </p>
    </div>
  );
}
