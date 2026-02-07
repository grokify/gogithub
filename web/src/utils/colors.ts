// GitHub contribution color palette and theme utilities

export const CONTRIBUTION_COLORS = {
  light: ['#ebedf0', '#9be9a8', '#40c463', '#30a14e', '#216e39'],
  dark: ['#161b22', '#0e4429', '#006d32', '#26a641', '#39d353'],
};

export const STATS_COLORS = {
  commits: '#40c463',
  prs: '#8957e5',
  issues: '#da3633',
  reviews: '#f0883e',
  additions: '#3fb950',
  deletions: '#f85149',
};

export const CHART_COLORS = {
  commits: 'rgba(64, 196, 99, 0.8)',
  prs: 'rgba(137, 87, 229, 0.8)',
  issues: 'rgba(218, 54, 51, 0.8)',
  reviews: 'rgba(240, 136, 62, 0.8)',
  additions: 'rgba(63, 185, 80, 0.8)',
  deletions: 'rgba(248, 81, 73, 0.8)',
};

export function getContributionColor(level: number, isDark: boolean = false): string {
  const palette = isDark ? CONTRIBUTION_COLORS.dark : CONTRIBUTION_COLORS.light;
  return palette[Math.min(level, 4)] || palette[0];
}

export function isDarkMode(): boolean {
  return document.documentElement.getAttribute('data-theme') === 'dark';
}

export function toggleDarkMode(): boolean {
  const html = document.documentElement;
  const isDark = html.getAttribute('data-theme') === 'dark';
  html.setAttribute('data-theme', isDark ? 'light' : 'dark');
  localStorage.setItem('theme', isDark ? 'light' : 'dark');
  return !isDark;
}

export function initTheme(): void {
  const saved = localStorage.getItem('theme');
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
  const theme = saved || (prefersDark ? 'dark' : 'light');
  document.documentElement.setAttribute('data-theme', theme);
}
