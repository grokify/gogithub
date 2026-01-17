package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/grokify/gogithub/search"
	"github.com/grokify/mogo/log/logutil"
	flags "github.com/jessevdk/go-flags"
)

type Options struct {
	Accounts []string `short:"a" long:"accounts" description:"Accounts" required:"true"`
	Outfile  string   `short:"o" long:"outfile" description:"Output File" required:"false"`
}

func main() {
	opts := Options{}
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("loading public pull requests for (%s)\n", strings.Join(opts.Accounts, ", "))
	opts.Outfile = strings.TrimSpace(opts.Outfile)
	if opts.Outfile == "" {
		opts.Outfile = "githubissues.xlsx"
	}

	if len(opts.Accounts) == 0 {
		fmt.Println("DONE")
		os.Exit(0)
	}

	c := search.NewClientHTTP(nil)

	ii := search.Issues{}

	for _, acct := range opts.Accounts {
		iss2, err := c.SearchIssuesAll(context.Background(), search.Query{
			search.ParamUser:  acct,
			search.ParamState: search.ParamStateValueOpen,
			search.ParamIs:    search.ParamIsValuePR,
		}, nil)
		logutil.FatalErr(err)
		ii = append(ii, iss2...)
	}

	ts, err := ii.TableSet()
	logutil.FatalErr(err)

	err = ts.WriteXLSX(opts.Outfile)
	logutil.FatalErr(err)
	fmt.Printf("WROTE (%s)\n", opts.Outfile)

	fmt.Println("DONE")
}
