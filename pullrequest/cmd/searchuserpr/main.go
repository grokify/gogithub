package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/grokify/gocharts/v2/data/table"
	"github.com/grokify/gogithub"
	"github.com/grokify/mogo/log/logutil"
	flags "github.com/jessevdk/go-flags"
)

type Options struct {
	Accounts []string `short:"a" long:"accounts" description:"Accounts" required:"true"`
}

func main() {
	opts := Options{}
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("loading public pull requests for (%s)\n", strings.Join(opts.Accounts, ", "))

	if len(opts.Accounts) == 0 {
		fmt.Println("DONE")
		os.Exit(0)
	}

	c := gogithub.NewClient(nil)

	var tbl *table.Table

	for i, acct := range opts.Accounts {
		iss2, err := c.SearchIssuesAll(context.Background(), gogithub.Query{
			"user":  acct,
			"state": gogithub.IssueStateOpen,
			"is":    gogithub.IssueIsPR,
		}, nil)
		logutil.FatalErr(err)

		tbl2, err := iss2.Table()
		logutil.FatalErr(err)

		if i == 0 {
			tbl = tbl2
		} else {
			tbl.Rows = append(tbl.Rows, tbl2.Rows...)
		}
	}

	err = tbl.WriteXLSX("githubissues.xlsx", "issues")
	logutil.FatalErr(err)

	fmt.Println("DONE")
}
