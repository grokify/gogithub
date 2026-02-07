// TypeScript interfaces matching RawJSON from cmd/gogithub/cmd_profile.go

export interface RawJSON {
  username: string;
  from: string; // ISO date string
  to: string;
  generated_at: string;

  // Summary counts - GitHub official count from contributionsCollection
  total_commits: number;
  total_issues: number;
  total_prs: number;
  total_reviews: number;
  total_repos_created: number;
  restricted_contributions?: number;

  // Commits from default branch traversal
  commits_default_branch: number;

  // Code stats (from default branch traversal)
  total_additions: number;
  total_deletions: number;

  // Per-repo details
  repos: RepoJSON[];

  // Monthly breakdown
  monthly: MonthlyJSON[];

  // Calendar data
  calendar?: CalendarDataJSON;
}

export interface AggregateJSON {
  username: string;
  from: string;
  to: string;
  generated_at: string;

  total_commits: number;
  total_issues: number;
  total_prs: number;
  total_reviews: number;
  total_repos_created: number;
  restricted_contributions?: number;

  commits_default_branch: number;

  total_additions: number;
  total_deletions: number;

  repos_contributed_to: number;

  calendar?: CalendarStatsJSON;

  monthly?: MonthlyJSON[];
}

export interface CalendarDataJSON {
  total_contributions: number;
  weeks?: CalendarWeek[];
}

export interface CalendarWeek {
  start_date: string;
  days: CalendarDay[];
}

export interface CalendarDay {
  date: string;
  contribution_count: number;
  level: number;
}

export interface CalendarStatsJSON {
  total_contributions: number;
  days_with_contributions: number;
  longest_streak: number;
  current_streak: number;
}

export interface MonthlyJSON {
  year: number;
  month: number;
  month_name: string;
  commits: number;
  issues: number;
  prs: number;
  reviews: number;
  additions: number;
  deletions: number;
}

export interface RepoJSON {
  full_name: string;
  is_private: boolean;
  commits: number;
  additions: number;
  deletions: number;
}

// Type guard to check if data is RawJSON or AggregateJSON
export function isRawJSON(data: RawJSON | AggregateJSON): data is RawJSON {
  return 'repos' in data && Array.isArray(data.repos);
}
