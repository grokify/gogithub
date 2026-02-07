// Main entry point for GitHub Stats Viewer

import type { RawJSON, AggregateJSON } from './types';
import { isRawJSON } from './types';
import { initFileLoader, validateData } from './components/FileLoader';
import { renderStatsCard } from './components/StatsCard';
import { renderCalendar } from './components/Calendar';
import { initMonthlyChart, updateChartTheme } from './components/MonthlyChart';
import { renderRepoList } from './components/RepoList';
import { initTheme, toggleDarkMode, isDarkMode } from './utils/colors';
import { downloadSVG, downloadPNG } from './utils/download';
import { formatDateRange } from './utils/format';

// Current loaded data
let currentData: RawJSON | AggregateJSON | null = null;
let statsCardSVG: SVGElement | null = null;

// Initialize app
async function init(): Promise<void> {
  // Set up theme
  initTheme();
  updateThemeButton();

  // Try to auto-load baked-in data.json (for static site deployment)
  const autoLoaded = await tryAutoLoadData();

  // If no auto-loaded data, set up file loader for interactive mode
  if (!autoLoaded) {
    initFileLoader(handleDataLoad);
  }

  // Set up theme toggle
  const themeToggle = document.getElementById('theme-toggle');
  if (themeToggle) {
    themeToggle.addEventListener('click', () => {
      toggleDarkMode();
      updateThemeButton();

      // Re-render components with new theme
      if (currentData) {
        renderAllComponents(currentData);
      }
    });
  }

  // Set up download buttons
  const downloadSvgBtn = document.getElementById('download-svg');
  const downloadPngBtn = document.getElementById('download-png');

  if (downloadSvgBtn) {
    downloadSvgBtn.addEventListener('click', () => {
      if (statsCardSVG && currentData) {
        downloadSVG(statsCardSVG, `${currentData.username}-github-stats.svg`);
      }
    });
  }

  if (downloadPngBtn) {
    downloadPngBtn.addEventListener('click', () => {
      if (statsCardSVG && currentData) {
        downloadPNG(statsCardSVG, `${currentData.username}-github-stats.png`);
      }
    });
  }
}

function handleDataLoad(data: RawJSON | AggregateJSON): void {
  currentData = data;

  // Hide file loader, show dashboard
  const fileLoader = document.getElementById('file-loader');
  const dashboard = document.getElementById('dashboard');

  if (fileLoader) fileLoader.classList.add('hidden');
  if (dashboard) dashboard.classList.remove('hidden');

  // Update header
  const usernameEl = document.getElementById('username');
  const dateRangeEl = document.getElementById('date-range');

  if (usernameEl) usernameEl.textContent = `@${data.username}`;
  if (dateRangeEl) dateRangeEl.textContent = formatDateRange(data.from, data.to);

  // Render all components
  renderAllComponents(data);
}

function renderAllComponents(data: RawJSON | AggregateJSON): void {
  // Stats card
  const statsCardContainer = document.getElementById('stats-card-container');
  if (statsCardContainer) {
    statsCardSVG = renderStatsCard(statsCardContainer, data);
  }

  // Calendar
  const calendarContainer = document.getElementById('calendar-container');
  if (calendarContainer) {
    renderCalendar(calendarContainer, data);
  }

  // Monthly chart
  initMonthlyChart(data);

  // Repository list (only for raw data with repos)
  const reposContainer = document.getElementById('repos-table-container');
  const reposSection = document.getElementById('repos-section');

  if (reposContainer && reposSection) {
    if (isRawJSON(data) && data.repos?.length) {
      reposSection.classList.remove('hidden');
      renderRepoList(reposContainer, data);
    } else {
      reposSection.classList.add('hidden');
    }
  }

  // Update chart theme
  updateChartTheme(data);
}

function updateThemeButton(): void {
  const themeToggle = document.getElementById('theme-toggle');
  if (themeToggle) {
    themeToggle.textContent = isDarkMode() ? '☀️' : '🌙';
  }
}

// Try to auto-load data.json for static site deployment mode
async function tryAutoLoadData(): Promise<boolean> {
  try {
    const response = await fetch('./data.json');
    if (!response.ok) {
      return false;
    }

    const data = await response.json();
    if (!validateData(data)) {
      console.warn('data.json exists but is not valid gogithub profile format');
      return false;
    }

    // Successfully loaded - hide file loader UI elements
    const fileLoader = document.getElementById('file-loader');
    const demoBtn = document.getElementById('demo-btn');
    if (fileLoader) fileLoader.classList.add('hidden');
    if (demoBtn) demoBtn.classList.add('hidden');

    handleDataLoad(data);
    return true;
  } catch {
    // No data.json or failed to load - this is normal for interactive mode
    return false;
  }
}

// Start app when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', init);
} else {
  init();
}
