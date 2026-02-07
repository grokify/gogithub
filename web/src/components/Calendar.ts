// Contribution calendar heatmap generator

import type { RawJSON, AggregateJSON, CalendarWeek } from '../types';
import { getContributionColor, isDarkMode, CONTRIBUTION_COLORS } from '../utils/colors';
import { isRawJSON } from '../types';

const CELL_SIZE = 11;
const CELL_GAP = 3;
const MONTH_LABEL_HEIGHT = 15;
const DAY_LABEL_WIDTH = 28;

export function renderCalendar(
  container: HTMLElement,
  data: RawJSON | AggregateJSON
): void {
  // Only render if we have calendar week data
  if (!isRawJSON(data) || !data.calendar?.weeks?.length) {
    container.innerHTML = '<p style="color: var(--text-muted); text-align: center;">No calendar data available</p>';
    return;
  }

  const weeks = data.calendar.weeks;
  const svg = createCalendarSVG(weeks);

  container.innerHTML = '';
  container.appendChild(svg);

  // Add legend
  const legendContainer = document.getElementById('calendar-legend');
  if (legendContainer) {
    renderLegend(legendContainer);
  }
}

function createCalendarSVG(weeks: CalendarWeek[]): SVGElement {
  const isDark = isDarkMode();
  const ns = 'http://www.w3.org/2000/svg';

  const numWeeks = weeks.length;
  const width = DAY_LABEL_WIDTH + numWeeks * (CELL_SIZE + CELL_GAP) + 10;
  const height = MONTH_LABEL_HEIGHT + 7 * (CELL_SIZE + CELL_GAP) + 10;

  const svg = document.createElementNS(ns, 'svg');
  svg.setAttribute('width', width.toString());
  svg.setAttribute('height', height.toString());
  svg.setAttribute('viewBox', `0 0 ${width} ${height}`);
  svg.setAttribute('xmlns', ns);

  const textColor = isDark ? '#8b949e' : '#57606a';

  // Day labels (Mon, Wed, Fri)
  const dayLabels = ['', 'Mon', '', 'Wed', '', 'Fri', ''];
  dayLabels.forEach((label, i) => {
    if (!label) return;

    const text = document.createElementNS(ns, 'text');
    text.setAttribute('x', '0');
    text.setAttribute('y', (MONTH_LABEL_HEIGHT + i * (CELL_SIZE + CELL_GAP) + CELL_SIZE - 2).toString());
    text.setAttribute('fill', textColor);
    text.setAttribute('font-size', '9');
    text.setAttribute('font-family', '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif');
    text.textContent = label;
    svg.appendChild(text);
  });

  // Month labels
  const monthPositions = getMonthLabelPositions(weeks);
  monthPositions.forEach(({ month, x }) => {
    const text = document.createElementNS(ns, 'text');
    text.setAttribute('x', (DAY_LABEL_WIDTH + x * (CELL_SIZE + CELL_GAP)).toString());
    text.setAttribute('y', '10');
    text.setAttribute('fill', textColor);
    text.setAttribute('font-size', '9');
    text.setAttribute('font-family', '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif');
    text.textContent = month;
    svg.appendChild(text);
  });

  // Create a group for all cells
  const cellsGroup = document.createElementNS(ns, 'g');
  cellsGroup.setAttribute('transform', `translate(${DAY_LABEL_WIDTH}, ${MONTH_LABEL_HEIGHT})`);

  // Contribution cells
  weeks.forEach((week, weekIndex) => {
    week.days.forEach((day) => {
      const dayOfWeek = new Date(day.date).getDay();
      const x = weekIndex * (CELL_SIZE + CELL_GAP);
      const y = dayOfWeek * (CELL_SIZE + CELL_GAP);

      const rect = document.createElementNS(ns, 'rect');
      rect.setAttribute('x', x.toString());
      rect.setAttribute('y', y.toString());
      rect.setAttribute('width', CELL_SIZE.toString());
      rect.setAttribute('height', CELL_SIZE.toString());
      rect.setAttribute('rx', '2');
      rect.setAttribute('fill', getContributionColor(day.level, isDark));
      rect.setAttribute('data-date', day.date);
      rect.setAttribute('data-count', day.contribution_count.toString());

      // Add tooltip on hover
      rect.addEventListener('mouseenter', (e) => showTooltip(e, day.date, day.contribution_count));
      rect.addEventListener('mouseleave', hideTooltip);

      cellsGroup.appendChild(rect);
    });
  });

  svg.appendChild(cellsGroup);

  return svg;
}

function getMonthLabelPositions(weeks: CalendarWeek[]): { month: string; x: number }[] {
  const positions: { month: string; x: number }[] = [];
  const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
  let lastMonth = -1;

  weeks.forEach((week, i) => {
    // Use the first day of the week to determine the month
    const firstDay = week.days[0];
    if (!firstDay) return;

    const date = new Date(firstDay.date);
    const month = date.getMonth();

    if (month !== lastMonth) {
      positions.push({
        month: monthNames[month],
        x: i,
      });
      lastMonth = month;
    }
  });

  return positions;
}

function renderLegend(container: HTMLElement): void {
  const isDark = isDarkMode();
  const colors = isDark ? CONTRIBUTION_COLORS.dark : CONTRIBUTION_COLORS.light;

  container.innerHTML = `
    <span>Less</span>
    <div class="legend-squares">
      ${colors.map((color) => `<div class="legend-square" style="background: ${color}"></div>`).join('')}
    </div>
    <span>More</span>
  `;
}

let tooltipEl: HTMLElement | null = null;

function showTooltip(e: MouseEvent, date: string, count: number): void {
  if (!tooltipEl) {
    tooltipEl = document.createElement('div');
    tooltipEl.className = 'tooltip';
    document.body.appendChild(tooltipEl);
  }

  const formattedDate = new Date(date).toLocaleDateString('en-US', {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });

  const text = count === 0
    ? `No contributions on ${formattedDate}`
    : `${count} contribution${count === 1 ? '' : 's'} on ${formattedDate}`;

  tooltipEl.textContent = text;
  tooltipEl.style.display = 'block';

  const rect = (e.target as Element).getBoundingClientRect();
  tooltipEl.style.left = `${rect.left + rect.width / 2 - tooltipEl.offsetWidth / 2}px`;
  tooltipEl.style.top = `${rect.top - tooltipEl.offsetHeight - 8}px`;
}

function hideTooltip(): void {
  if (tooltipEl) {
    tooltipEl.style.display = 'none';
  }
}
