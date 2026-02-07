// SVG stats card generator

import type { RawJSON, AggregateJSON, CalendarStatsJSON } from '../types';
import { formatNumber } from '../utils/format';
import { isDarkMode, STATS_COLORS } from '../utils/colors';
import { isRawJSON } from '../types';

interface StatItem {
  label: string;
  value: number;
  icon: string;
  color: string;
}

export function renderStatsCard(
  container: HTMLElement,
  data: RawJSON | AggregateJSON
): SVGElement {
  const isDark = isDarkMode();

  // Calculate calendar stats if we have raw data with calendar
  let calendarStats: CalendarStatsJSON | undefined;
  if (isRawJSON(data) && data.calendar?.weeks) {
    calendarStats = calculateCalendarStats(data);
  } else if (!isRawJSON(data) && data.calendar) {
    calendarStats = data.calendar;
  }

  const stats: StatItem[] = [
    {
      label: 'Total Commits',
      value: data.total_commits,
      icon: '📝',
      color: STATS_COLORS.commits,
    },
    {
      label: 'Pull Requests',
      value: data.total_prs,
      icon: '🔀',
      color: STATS_COLORS.prs,
    },
    {
      label: 'Issues',
      value: data.total_issues,
      icon: '🔴',
      color: STATS_COLORS.issues,
    },
    {
      label: 'Code Reviews',
      value: data.total_reviews,
      icon: '👀',
      color: STATS_COLORS.reviews,
    },
    {
      label: 'Lines Added',
      value: data.total_additions,
      icon: '➕',
      color: STATS_COLORS.additions,
    },
    {
      label: 'Lines Deleted',
      value: data.total_deletions,
      icon: '➖',
      color: STATS_COLORS.deletions,
    },
  ];

  // Add streak if available
  if (calendarStats?.longest_streak) {
    stats.push({
      label: 'Longest Streak',
      value: calendarStats.longest_streak,
      icon: '🔥',
      color: '#f97316',
    });
  }

  const svg = createStatsCardSVG(data.username, stats, isDark);
  container.innerHTML = '';
  container.appendChild(svg);

  return svg;
}

function createStatsCardSVG(
  username: string,
  stats: StatItem[],
  isDark: boolean
): SVGElement {
  const width = 400;
  const headerHeight = 50;
  const statHeight = 36;
  const padding = 20;
  const height = headerHeight + stats.length * statHeight + padding * 2;

  const bgColor = isDark ? '#0d1117' : '#ffffff';
  const borderColor = isDark ? '#30363d' : '#e1e4e8';
  const textColor = isDark ? '#c9d1d9' : '#24292f';
  const secondaryColor = isDark ? '#8b949e' : '#57606a';

  const ns = 'http://www.w3.org/2000/svg';
  const svg = document.createElementNS(ns, 'svg');
  svg.setAttribute('width', width.toString());
  svg.setAttribute('height', height.toString());
  svg.setAttribute('viewBox', `0 0 ${width} ${height}`);
  svg.setAttribute('xmlns', ns);

  // Background with border
  const bg = document.createElementNS(ns, 'rect');
  bg.setAttribute('width', width.toString());
  bg.setAttribute('height', height.toString());
  bg.setAttribute('rx', '6');
  bg.setAttribute('fill', bgColor);
  bg.setAttribute('stroke', borderColor);
  bg.setAttribute('stroke-width', '1');
  svg.appendChild(bg);

  // Header
  const header = document.createElementNS(ns, 'text');
  header.setAttribute('x', padding.toString());
  header.setAttribute('y', (padding + 20).toString());
  header.setAttribute('fill', textColor);
  header.setAttribute('font-size', '18');
  header.setAttribute('font-weight', '600');
  header.setAttribute('font-family', '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif');
  header.textContent = `${username}'s GitHub Stats`;
  svg.appendChild(header);

  // Stats
  stats.forEach((stat, i) => {
    const y = headerHeight + padding + i * statHeight;

    // Icon
    const icon = document.createElementNS(ns, 'text');
    icon.setAttribute('x', padding.toString());
    icon.setAttribute('y', (y + 20).toString());
    icon.setAttribute('font-size', '14');
    icon.textContent = stat.icon;
    svg.appendChild(icon);

    // Label
    const label = document.createElementNS(ns, 'text');
    label.setAttribute('x', (padding + 26).toString());
    label.setAttribute('y', (y + 20).toString());
    label.setAttribute('fill', secondaryColor);
    label.setAttribute('font-size', '14');
    label.setAttribute('font-family', '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif');
    label.textContent = stat.label;
    svg.appendChild(label);

    // Value
    const value = document.createElementNS(ns, 'text');
    value.setAttribute('x', (width - padding).toString());
    value.setAttribute('y', (y + 20).toString());
    value.setAttribute('fill', stat.color);
    value.setAttribute('font-size', '14');
    value.setAttribute('font-weight', '600');
    value.setAttribute('font-family', '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif');
    value.setAttribute('text-anchor', 'end');
    value.textContent = formatNumber(stat.value);
    svg.appendChild(value);
  });

  return svg;
}

function calculateCalendarStats(data: RawJSON): CalendarStatsJSON {
  if (!data.calendar?.weeks) {
    return {
      total_contributions: 0,
      days_with_contributions: 0,
      longest_streak: 0,
      current_streak: 0,
    };
  }

  let totalContributions = 0;
  let daysWithContributions = 0;
  let longestStreak = 0;
  let currentStreak = 0;
  let tempStreak = 0;

  // Flatten all days
  const allDays = data.calendar.weeks.flatMap((w) => w.days);

  for (const day of allDays) {
    totalContributions += day.contribution_count;

    if (day.contribution_count > 0) {
      daysWithContributions++;
      tempStreak++;
      longestStreak = Math.max(longestStreak, tempStreak);
    } else {
      tempStreak = 0;
    }
  }

  // Calculate current streak from the end
  for (let i = allDays.length - 1; i >= 0; i--) {
    if (allDays[i].contribution_count > 0) {
      currentStreak++;
    } else {
      // Allow one day gap for "today" not having contributions yet
      if (i === allDays.length - 1) {
        continue;
      }
      break;
    }
  }

  return {
    total_contributions: totalContributions,
    days_with_contributions: daysWithContributions,
    longest_streak: longestStreak,
    current_streak: currentStreak,
  };
}
