export interface TimeRange {
  id: string;
  startMinutes: number;
  endMinutes: number;
}

export function hasOverlap(ranges: TimeRange[]): boolean {
  const sorted = [...ranges].sort((a, b) => a.startMinutes - b.startMinutes);
  for (let i = 1; i < sorted.length; i++) {
    if (sorted[i].startMinutes < sorted[i - 1].endMinutes) return true;
  }
  return false;
}
