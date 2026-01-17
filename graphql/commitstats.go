package graphql

import (
	"context"
	"sort"
	"time"

	"github.com/shurcooL/githubv4"
)

// Visibility represents repository visibility for filtering.
type Visibility int

const (
	// VisibilityAll includes both public and private repositories.
	VisibilityAll Visibility = iota
	// VisibilityPublic includes only public repositories.
	VisibilityPublic
	// VisibilityPrivate includes only private repositories.
	VisibilityPrivate
)

// CommitStats represents detailed commit statistics including additions and deletions.
type CommitStats struct {
	Username     string
	From         time.Time
	To           time.Time
	Visibility   Visibility
	TotalCommits int
	Additions    int
	Deletions    int
	ByMonth      []MonthlyCommitStats
	ByRepo       []RepoCommitStats
}

// MonthlyCommitStats represents commit statistics for a specific month.
type MonthlyCommitStats struct {
	Year      int
	Month     time.Month
	Commits   int
	Additions int
	Deletions int
}

// YearMonth returns a formatted string like "2024-01".
func (mcs MonthlyCommitStats) YearMonth() string {
	return time.Date(mcs.Year, mcs.Month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
}

// RepoCommitStats represents commit statistics for a specific repository.
type RepoCommitStats struct {
	Owner     string
	Name      string
	IsPrivate bool
	Commits   int
	Additions int
	Deletions int
}

// repositoriesContributedToQuery fetches repositories user has contributed to.
type repositoriesContributedToQuery struct {
	User struct {
		ID                        githubv4.ID
		RepositoriesContributedTo struct {
			PageInfo struct {
				HasNextPage bool
				EndCursor   githubv4.String
			}
			Nodes []struct {
				Owner struct {
					Login githubv4.String
				}
				Name      githubv4.String
				IsPrivate githubv4.Boolean
			}
		} `graphql:"repositoriesContributedTo(first: 100, after: $cursor, contributionTypes: COMMIT, includeUserRepositories: true)"`
	} `graphql:"user(login: $login)"`
}

// commitHistoryQuery fetches commit history for a specific repository.
type commitHistoryQuery struct {
	Repository struct {
		DefaultBranchRef struct {
			Target struct {
				Commit struct {
					History struct {
						PageInfo struct {
							HasNextPage bool
							EndCursor   githubv4.String
						}
						Nodes []struct {
							Additions     githubv4.Int
							Deletions     githubv4.Int
							CommittedDate githubv4.DateTime
						}
					} `graphql:"history(first: 100, after: $cursor, author: {id: $authorId}, since: $since, until: $until)"`
				} `graphql:"... on Commit"`
			}
		}
	} `graphql:"repository(owner: $owner, name: $name)"`
}

// GetCommitStats retrieves detailed commit statistics including additions and deletions.
// This method iterates through all repositories the user has contributed to and aggregates
// commit data. Use the visibility parameter to filter by public/private repositories.
func GetCommitStats(ctx context.Context, client *githubv4.Client, username string, from, to time.Time, visibility Visibility) (*CommitStats, error) {
	// First, get the user's ID for author filtering
	userID, err := getUserID(ctx, client, username)
	if err != nil {
		return nil, err
	}

	// Get all repositories the user has contributed to
	repos, err := getContributedRepositories(ctx, client, username, visibility)
	if err != nil {
		return nil, err
	}

	stats := &CommitStats{
		Username:   username,
		From:       from,
		To:         to,
		Visibility: visibility,
		ByMonth:    []MonthlyCommitStats{},
		ByRepo:     []RepoCommitStats{},
	}

	monthlyMap := make(map[string]*MonthlyCommitStats)

	// Iterate through each repository and get commit history
	for _, repo := range repos {
		repoStats, monthData, err := getRepoCommitStats(ctx, client, repo.Owner, repo.Name, repo.IsPrivate, userID, from, to)
		if err != nil {
			// Skip repos we can't access (might have lost access)
			continue
		}

		if repoStats.Commits > 0 {
			stats.ByRepo = append(stats.ByRepo, *repoStats)
			stats.TotalCommits += repoStats.Commits
			stats.Additions += repoStats.Additions
			stats.Deletions += repoStats.Deletions

			// Aggregate monthly data
			for ym, data := range monthData {
				if existing, ok := monthlyMap[ym]; ok {
					existing.Commits += data.Commits
					existing.Additions += data.Additions
					existing.Deletions += data.Deletions
				} else {
					monthlyMap[ym] = &MonthlyCommitStats{
						Year:      data.Year,
						Month:     data.Month,
						Commits:   data.Commits,
						Additions: data.Additions,
						Deletions: data.Deletions,
					}
				}
			}
		}
	}

	// Convert monthly map to sorted slice
	stats.ByMonth = monthlyMapToSlice(monthlyMap)

	// Sort repos by commit count descending
	sort.Slice(stats.ByRepo, func(i, j int) bool {
		return stats.ByRepo[i].Commits > stats.ByRepo[j].Commits
	})

	return stats, nil
}

// GetCommitStatsByVisibility returns separate stats for public, private, and combined.
func GetCommitStatsByVisibility(ctx context.Context, client *githubv4.Client, username string, from, to time.Time) (all, public, private *CommitStats, err error) {
	all, err = GetCommitStats(ctx, client, username, from, to, VisibilityAll)
	if err != nil {
		return nil, nil, nil, err
	}

	public, err = GetCommitStats(ctx, client, username, from, to, VisibilityPublic)
	if err != nil {
		return nil, nil, nil, err
	}

	private, err = GetCommitStats(ctx, client, username, from, to, VisibilityPrivate)
	if err != nil {
		return nil, nil, nil, err
	}

	return all, public, private, nil
}

// userIDQuery fetches a user's node ID.
type userIDQuery struct {
	User struct {
		ID githubv4.ID
	} `graphql:"user(login: $login)"`
}

// getUserID retrieves the GraphQL node ID for a username.
func getUserID(ctx context.Context, client *githubv4.Client, username string) (githubv4.ID, error) {
	var query userIDQuery
	variables := map[string]any{
		"login": githubv4.String(username),
	}

	if err := client.Query(ctx, &query, variables); err != nil {
		return "", err
	}

	return query.User.ID, nil
}

// repoInfo holds basic repository information.
type repoInfo struct {
	Owner     string
	Name      string
	IsPrivate bool
}

// getContributedRepositories fetches all repositories a user has contributed commits to.
func getContributedRepositories(ctx context.Context, client *githubv4.Client, username string, visibility Visibility) ([]repoInfo, error) {
	var repos []repoInfo
	var cursor *githubv4.String

	for {
		var query repositoriesContributedToQuery
		variables := map[string]any{
			"login":  githubv4.String(username),
			"cursor": cursor,
		}

		if err := client.Query(ctx, &query, variables); err != nil {
			return nil, err
		}

		for _, node := range query.User.RepositoriesContributedTo.Nodes {
			isPrivate := bool(node.IsPrivate)

			// Filter by visibility
			switch visibility {
			case VisibilityPublic:
				if isPrivate {
					continue
				}
			case VisibilityPrivate:
				if !isPrivate {
					continue
				}
			}

			repos = append(repos, repoInfo{
				Owner:     string(node.Owner.Login),
				Name:      string(node.Name),
				IsPrivate: isPrivate,
			})
		}

		if !query.User.RepositoriesContributedTo.PageInfo.HasNextPage {
			break
		}
		cursor = &query.User.RepositoriesContributedTo.PageInfo.EndCursor
	}

	return repos, nil
}

// getRepoCommitStats fetches commit statistics for a specific repository.
func getRepoCommitStats(ctx context.Context, client *githubv4.Client, owner, name string, isPrivate bool, authorID githubv4.ID, from, to time.Time) (*RepoCommitStats, map[string]*MonthlyCommitStats, error) {
	repoStats := &RepoCommitStats{
		Owner:     owner,
		Name:      name,
		IsPrivate: isPrivate,
	}
	monthlyData := make(map[string]*MonthlyCommitStats)

	var cursor *githubv4.String

	for {
		var query commitHistoryQuery
		variables := map[string]any{
			"owner":    githubv4.String(owner),
			"name":     githubv4.String(name),
			"authorId": authorID,
			"since":    githubv4.GitTimestamp{Time: from},
			"until":    githubv4.GitTimestamp{Time: to},
			"cursor":   cursor,
		}

		if err := client.Query(ctx, &query, variables); err != nil {
			return nil, nil, err
		}

		history := query.Repository.DefaultBranchRef.Target.Commit.History

		for _, commit := range history.Nodes {
			additions := int(commit.Additions)
			deletions := int(commit.Deletions)
			committedDate := commit.CommittedDate.Time

			repoStats.Commits++
			repoStats.Additions += additions
			repoStats.Deletions += deletions

			// Aggregate by month
			ym := committedDate.Format("2006-01")
			if existing, ok := monthlyData[ym]; ok {
				existing.Commits++
				existing.Additions += additions
				existing.Deletions += deletions
			} else {
				monthlyData[ym] = &MonthlyCommitStats{
					Year:      committedDate.Year(),
					Month:     committedDate.Month(),
					Commits:   1,
					Additions: additions,
					Deletions: deletions,
				}
			}
		}

		if !history.PageInfo.HasNextPage {
			break
		}
		cursor = &history.PageInfo.EndCursor
	}

	return repoStats, monthlyData, nil
}

// monthlyMapToSlice converts a map to a sorted slice of MonthlyCommitStats.
func monthlyMapToSlice(m map[string]*MonthlyCommitStats) []MonthlyCommitStats {
	result := make([]MonthlyCommitStats, 0, len(m))

	for _, stats := range m {
		result = append(result, *stats)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Year != result[j].Year {
			return result[i].Year < result[j].Year
		}
		return result[i].Month < result[j].Month
	})

	return result
}
