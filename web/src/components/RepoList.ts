// Sortable repository list table

import type { RawJSON, RepoJSON } from '../types';
import { formatNumberWithCommas } from '../utils/format';
import { isRawJSON } from '../types';

type SortKey = 'full_name' | 'commits' | 'additions' | 'deletions';
type SortDirection = 'asc' | 'desc';

let currentSortKey: SortKey = 'commits';
let currentSortDirection: SortDirection = 'desc';

export function renderRepoList(container: HTMLElement, data: RawJSON): void {
  if (!isRawJSON(data) || !data.repos?.length) {
    container.innerHTML = '<p style="color: var(--text-muted); text-align: center;">No repository data available</p>';
    return;
  }

  renderTable(container, data.repos);
}

function renderTable(container: HTMLElement, repos: RepoJSON[]): void {
  const sorted = sortRepos(repos, currentSortKey, currentSortDirection);

  const table = document.createElement('table');
  table.className = 'repos-table';

  // Header
  const thead = document.createElement('thead');
  thead.innerHTML = `
    <tr>
      <th data-sort="full_name" class="${getSortClass('full_name')}">Repository</th>
      <th data-sort="commits" class="${getSortClass('commits')}">Commits</th>
      <th data-sort="additions" class="${getSortClass('additions')}">Additions</th>
      <th data-sort="deletions" class="${getSortClass('deletions')}">Deletions</th>
    </tr>
  `;
  table.appendChild(thead);

  // Body
  const tbody = document.createElement('tbody');

  // Show top 20 repos by default
  const topRepos = sorted.slice(0, 20);

  for (const repo of topRepos) {
    const tr = document.createElement('tr');

    // Repository name
    const tdName = document.createElement('td');
    const nameSpan = document.createElement('span');
    nameSpan.className = 'repo-name';
    nameSpan.textContent = repo.full_name;
    tdName.appendChild(nameSpan);

    if (repo.is_private) {
      const privateSpan = document.createElement('span');
      privateSpan.className = 'repo-private';
      privateSpan.textContent = '(private)';
      tdName.appendChild(privateSpan);
    }
    tr.appendChild(tdName);

    // Commits
    const tdCommits = document.createElement('td');
    tdCommits.textContent = formatNumberWithCommas(repo.commits);
    tr.appendChild(tdCommits);

    // Additions
    const tdAdditions = document.createElement('td');
    tdAdditions.className = 'additions';
    tdAdditions.textContent = '+' + formatNumberWithCommas(repo.additions);
    tr.appendChild(tdAdditions);

    // Deletions
    const tdDeletions = document.createElement('td');
    tdDeletions.className = 'deletions';
    tdDeletions.textContent = '-' + formatNumberWithCommas(repo.deletions);
    tr.appendChild(tdDeletions);

    tbody.appendChild(tr);
  }

  table.appendChild(tbody);

  // Add click handlers for sorting
  table.querySelectorAll('th[data-sort]').forEach((th) => {
    th.addEventListener('click', () => {
      const key = th.getAttribute('data-sort') as SortKey;

      if (key === currentSortKey) {
        currentSortDirection = currentSortDirection === 'asc' ? 'desc' : 'asc';
      } else {
        currentSortKey = key;
        currentSortDirection = key === 'full_name' ? 'asc' : 'desc';
      }

      renderTable(container, repos);
    });
  });

  container.innerHTML = '';
  container.appendChild(table);

  // Add "show more" if there are more repos
  if (sorted.length > 20) {
    const showMore = document.createElement('p');
    showMore.style.cssText = 'color: var(--text-muted); text-align: center; margin-top: var(--spacing-md); font-size: 0.875rem;';
    showMore.textContent = `Showing top 20 of ${sorted.length} repositories`;
    container.appendChild(showMore);
  }
}

function sortRepos(repos: RepoJSON[], key: SortKey, direction: SortDirection): RepoJSON[] {
  return [...repos].sort((a, b) => {
    let comparison = 0;

    if (key === 'full_name') {
      comparison = a.full_name.localeCompare(b.full_name);
    } else {
      comparison = a[key] - b[key];
    }

    return direction === 'asc' ? comparison : -comparison;
  });
}

function getSortClass(key: SortKey): string {
  if (key !== currentSortKey) return '';
  return currentSortDirection === 'asc' ? 'sorted-asc' : 'sorted-desc';
}
