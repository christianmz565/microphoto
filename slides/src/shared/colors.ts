export const colors = {
  teal: "#14b8a6",
  red: "#f87171",
  purple: "#a78bfa",
  amber: "#fbbf24",
  green: "#34d399",
  blue: "#60a5fa",
  petrol: "#005F76",
} as const;

export type ColorKey = keyof typeof colors;
