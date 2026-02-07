// Number and date formatting utilities

export function formatNumber(num: number): string {
  if (num >= 1_000_000) {
    return (num / 1_000_000).toFixed(1).replace(/\.0$/, '') + 'M';
  }
  if (num >= 1_000) {
    return (num / 1_000).toFixed(1).replace(/\.0$/, '') + 'k';
  }
  return num.toString();
}

export function formatNumberWithCommas(num: number): string {
  return num.toLocaleString('en-US');
}

export function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

export function formatDateRange(from: string, to: string): string {
  return `${formatDate(from)} - ${formatDate(to)}`;
}

export function formatMonthYear(year: number, month: number): string {
  const date = new Date(year, month - 1);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
  });
}

export function parseDate(dateStr: string): Date {
  return new Date(dateStr);
}

export function getDayOfWeek(dateStr: string): number {
  return new Date(dateStr).getDay();
}

export function getMonthName(month: number): string {
  const date = new Date(2000, month - 1);
  return date.toLocaleDateString('en-US', { month: 'short' });
}
