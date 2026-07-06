export function DotList({ items, color }: { items: string[]; color: string }) {
  return (
    <div className="space-y-1">
      {items.map((i) => (
        <div key={i} className="flex items-baseline gap-2 text-2xl">
          <span className="text-2xl font-bold" style={{ color }}>
            •
          </span>
          <span className="text-gray-300">{i}</span>
        </div>
      ))}
    </div>
  );
}
