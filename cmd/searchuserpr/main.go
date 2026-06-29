package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v88/github"
	"github.com/grokify/gogithub/search"
	flags "github.com/jessevdk/go-flags"
)

type Options struct {
	Accounts []string `short:"a" long:"accounts" description:"GitHub accounts to search" required:"true"`
	Outfile  string   `short:"o" long:"outfile" description:"Output file (format from extension: .xlsx, .md, .csv)" required:"false"`
}

func main() {
	opts := Options{}
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	if len(opts.Accounts) == 0 {
		fmt.Println("No accounts specified")
		os.Exit(0)
	}

	opts.Outfile = strings.TrimSpace(opts.Outfile)

	// Only print status messages when writing to file (not stdout)
	if opts.Outfile != "" {
		fmt.Fprintf(os.Stderr, "Loading public pull requests for (%s)\n", strings.Join(opts.Accounts, ", "))
	}

	ghClient, err := github.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	c := search.NewClient(ghClient)

	ii := search.Issues{}

	for _, acct := range opts.Accounts {
		qry := search.NewQuery().User(acct).StateOpen().IsPR().Build()
		iss, err := c.SearchIssuesAll(context.Background(), qry, nil)
		if err != nil {
			log.Fatalf("search issues for %s: %v", acct, err)
		}
		ii = append(ii, iss...)
	}

	// Output to stdout if no outfile specified
	if opts.Outfile == "" {
		tbl, err := ii.Table("Pull Requests")
		if err != nil {
			log.Fatalf("create table: %v", err)
		}
		if err := tbl.Text(os.Stdout); err != nil {
			log.Fatalf("write table: %v", err)
		}
		return
	}

	// Write to file based on extension
	ext := strings.ToLower(filepath.Ext(opts.Outfile))
	switch ext {
	case ".xlsx":
		ts, err := ii.TableSet()
		if err != nil {
			log.Fatalf("create table: %v", err)
		}
		if err := ts.WriteXLSX(opts.Outfile); err != nil {
			log.Fatalf("write xlsx: %v", err)
		}
	case ".md":
		tbl, err := ii.Table("Pull Requests")
		if err != nil {
			log.Fatalf("create table: %v", err)
		}
		if err := tbl.WriteMarkdown(opts.Outfile, 0644, "\n", true); err != nil {
			log.Fatalf("write markdown: %v", err)
		}
	case ".csv":
		tbl, err := ii.Table("Pull Requests")
		if err != nil {
			log.Fatalf("create table: %v", err)
		}
		if err := tbl.WriteCSV(opts.Outfile); err != nil {
			log.Fatalf("write csv: %v", err)
		}
	default:
		log.Fatalf("unsupported file extension %q (use .xlsx, .md, or .csv)", ext)
	}

	fmt.Fprintf(os.Stderr, "Wrote %s\n", opts.Outfile)
}
