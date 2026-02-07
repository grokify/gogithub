// Monthly activity chart using Chart.js

import {
  Chart,
  CategoryScale,
  LinearScale,
  BarElement,
  LineElement,
  PointElement,
  Title,
  Tooltip,
  Legend,
  BarController,
  LineController,
} from 'chart.js';
import type { RawJSON, AggregateJSON, MonthlyJSON } from '../types';
import { CHART_COLORS } from '../utils/colors';
import { formatMonthYear } from '../utils/format';

// Register Chart.js components
Chart.register(
  CategoryScale,
  LinearScale,
  BarElement,
  LineElement,
  PointElement,
  Title,
  Tooltip,
  Legend,
  BarController,
  LineController
);

let currentChart: Chart | null = null;

export function initMonthlyChart(data: RawJSON | AggregateJSON): void {
  const tabs = document.querySelectorAll('.chart-tabs .tab');

  tabs.forEach((tab) => {
    tab.addEventListener('click', () => {
      tabs.forEach((t) => t.classList.remove('active'));
      tab.classList.add('active');

      const chartType = tab.getAttribute('data-chart');
      if (chartType === 'activity') {
        renderActivityChart(data.monthly || []);
      } else {
        renderCodeChart(data.monthly || []);
      }
    });
  });

  // Initial render
  renderActivityChart(data.monthly || []);
}

export function renderActivityChart(monthly: MonthlyJSON[]): void {
  if (!monthly.length) {
    showNoData();
    return;
  }

  const canvas = document.getElementById('monthly-chart') as HTMLCanvasElement;
  if (!canvas) return;

  if (currentChart) {
    currentChart.destroy();
  }

  const labels = monthly.map((m) => formatMonthYear(m.year, m.month));
  const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
  const gridColor = isDark ? 'rgba(139, 148, 158, 0.2)' : 'rgba(0, 0, 0, 0.1)';
  const textColor = isDark ? '#8b949e' : '#57606a';

  currentChart = new Chart(canvas, {
    type: 'bar',
    data: {
      labels,
      datasets: [
        {
          label: 'Commits',
          data: monthly.map((m) => m.commits),
          backgroundColor: CHART_COLORS.commits,
          borderRadius: 4,
        },
        {
          label: 'Pull Requests',
          data: monthly.map((m) => m.prs),
          backgroundColor: CHART_COLORS.prs,
          borderRadius: 4,
        },
        {
          label: 'Issues',
          data: monthly.map((m) => m.issues),
          backgroundColor: CHART_COLORS.issues,
          borderRadius: 4,
        },
        {
          label: 'Reviews',
          data: monthly.map((m) => m.reviews),
          backgroundColor: CHART_COLORS.reviews,
          borderRadius: 4,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: {
        intersect: false,
        mode: 'index',
      },
      scales: {
        x: {
          stacked: true,
          grid: {
            display: false,
          },
          ticks: {
            color: textColor,
          },
        },
        y: {
          stacked: true,
          beginAtZero: true,
          grid: {
            color: gridColor,
          },
          ticks: {
            color: textColor,
          },
        },
      },
      plugins: {
        legend: {
          position: 'top',
          labels: {
            color: textColor,
            usePointStyle: true,
            padding: 16,
          },
        },
        tooltip: {
          backgroundColor: isDark ? '#21262d' : '#ffffff',
          titleColor: isDark ? '#c9d1d9' : '#24292f',
          bodyColor: isDark ? '#8b949e' : '#57606a',
          borderColor: isDark ? '#30363d' : '#d0d7de',
          borderWidth: 1,
        },
      },
    },
  });
}

export function renderCodeChart(monthly: MonthlyJSON[]): void {
  if (!monthly.length) {
    showNoData();
    return;
  }

  const canvas = document.getElementById('monthly-chart') as HTMLCanvasElement;
  if (!canvas) return;

  if (currentChart) {
    currentChart.destroy();
  }

  const labels = monthly.map((m) => formatMonthYear(m.year, m.month));
  const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
  const gridColor = isDark ? 'rgba(139, 148, 158, 0.2)' : 'rgba(0, 0, 0, 0.1)';
  const textColor = isDark ? '#8b949e' : '#57606a';

  currentChart = new Chart(canvas, {
    type: 'line',
    data: {
      labels,
      datasets: [
        {
          label: 'Additions',
          data: monthly.map((m) => m.additions),
          borderColor: CHART_COLORS.additions,
          backgroundColor: 'rgba(63, 185, 80, 0.1)',
          fill: true,
          tension: 0.3,
          pointRadius: 4,
          pointHoverRadius: 6,
        },
        {
          label: 'Deletions',
          data: monthly.map((m) => m.deletions),
          borderColor: CHART_COLORS.deletions,
          backgroundColor: 'rgba(248, 81, 73, 0.1)',
          fill: true,
          tension: 0.3,
          pointRadius: 4,
          pointHoverRadius: 6,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: {
        intersect: false,
        mode: 'index',
      },
      scales: {
        x: {
          grid: {
            display: false,
          },
          ticks: {
            color: textColor,
          },
        },
        y: {
          beginAtZero: true,
          grid: {
            color: gridColor,
          },
          ticks: {
            color: textColor,
          },
        },
      },
      plugins: {
        legend: {
          position: 'top',
          labels: {
            color: textColor,
            usePointStyle: true,
            padding: 16,
          },
        },
        tooltip: {
          backgroundColor: isDark ? '#21262d' : '#ffffff',
          titleColor: isDark ? '#c9d1d9' : '#24292f',
          bodyColor: isDark ? '#8b949e' : '#57606a',
          borderColor: isDark ? '#30363d' : '#d0d7de',
          borderWidth: 1,
          callbacks: {
            label: (context) => {
              const value = context.raw as number;
              return `${context.dataset.label}: ${value.toLocaleString()}`;
            },
          },
        },
      },
    },
  });
}

function showNoData(): void {
  const canvas = document.getElementById('monthly-chart') as HTMLCanvasElement;
  if (!canvas) return;

  if (currentChart) {
    currentChart.destroy();
    currentChart = null;
  }

  const container = document.getElementById('chart-container');
  if (container) {
    container.innerHTML = '<p style="color: var(--text-muted); text-align: center; padding: 40px;">No monthly data available</p>';
  }
}

// Re-render chart when theme changes
export function updateChartTheme(data: RawJSON | AggregateJSON): void {
  const activeTab = document.querySelector('.chart-tabs .tab.active');
  const chartType = activeTab?.getAttribute('data-chart');

  if (chartType === 'activity') {
    renderActivityChart(data.monthly || []);
  } else {
    renderCodeChart(data.monthly || []);
  }
}
