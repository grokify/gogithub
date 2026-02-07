// File loader component with drag-drop support

import type { RawJSON, AggregateJSON } from '../types';

export type LoadCallback = (data: RawJSON | AggregateJSON) => void;

export function initFileLoader(onLoad: LoadCallback): void {
  const dropZone = document.getElementById('drop-zone');
  const fileInput = document.getElementById('file-input') as HTMLInputElement;
  const demoBtn = document.getElementById('demo-btn');

  if (!dropZone || !fileInput) {
    console.error('File loader elements not found');
    return;
  }

  // Click to open file picker
  dropZone.addEventListener('click', () => fileInput.click());

  // File input change
  fileInput.addEventListener('change', () => {
    const file = fileInput.files?.[0];
    if (file) {
      loadFile(file, onLoad);
    }
  });

  // Drag and drop events
  dropZone.addEventListener('dragenter', (e) => {
    e.preventDefault();
    dropZone.classList.add('drag-over');
  });

  dropZone.addEventListener('dragover', (e) => {
    e.preventDefault();
    dropZone.classList.add('drag-over');
  });

  dropZone.addEventListener('dragleave', (e) => {
    e.preventDefault();
    dropZone.classList.remove('drag-over');
  });

  dropZone.addEventListener('drop', (e) => {
    e.preventDefault();
    dropZone.classList.remove('drag-over');

    const file = e.dataTransfer?.files[0];
    if (file) {
      loadFile(file, onLoad);
    }
  });

  // Demo button
  if (demoBtn) {
    demoBtn.addEventListener('click', () => loadDemo(onLoad));
  }
}

async function loadFile(file: File, onLoad: LoadCallback): Promise<void> {
  if (!file.name.endsWith('.json') && file.type !== 'application/json') {
    showError('Please select a JSON file');
    return;
  }

  try {
    const text = await file.text();
    const data = JSON.parse(text);

    if (!validateData(data)) {
      showError('Invalid gogithub profile JSON format');
      return;
    }

    onLoad(data);
  } catch (err) {
    showError('Failed to parse JSON file');
    console.error(err);
  }
}

async function loadDemo(onLoad: LoadCallback): Promise<void> {
  try {
    const response = await fetch('./sample-data.json');
    if (!response.ok) {
      throw new Error('Demo data not found');
    }
    const data = await response.json();

    if (!validateData(data)) {
      showError('Invalid demo data format');
      return;
    }

    onLoad(data);
  } catch (err) {
    showError('Failed to load demo data');
    console.error(err);
  }
}

export function validateData(data: unknown): data is RawJSON | AggregateJSON {
  if (typeof data !== 'object' || data === null) {
    return false;
  }

  const obj = data as Record<string, unknown>;

  // Check required fields
  const requiredFields = [
    'username',
    'from',
    'to',
    'total_commits',
    'total_issues',
    'total_prs',
    'total_reviews',
  ];

  for (const field of requiredFields) {
    if (!(field in obj)) {
      console.error(`Missing required field: ${field}`);
      return false;
    }
  }

  return true;
}

function showError(message: string): void {
  alert(message); // Simple error display
}
