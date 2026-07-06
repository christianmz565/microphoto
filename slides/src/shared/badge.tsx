export function Badge({ label, color }: { label: string; color: string }) {
  return (
    <span
      className="rounded-full px-1.5 py-0.5 text-lg font-bold uppercase tracking-wider"
      style={{ backgroundColor: `${color}30`, color }}
    >
      {label}
    </span>
  );
}
