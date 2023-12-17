package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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

	ii := gogithub.Issues{}

	for _, acct := range opts.Accounts {
		iss2, err := c.SearchIssuesAll(context.Background(), gogithub.Query{
			gogithub.ParamUser:  acct,
			gogithub.ParamState: gogithub.ParamStateValueOpen,
			gogithub.ParamIs:    gogithub.ParamIsValuePR,
		}, nil)
		logutil.FatalErr(err)
		ii = append(ii, iss2...)
	}

	ts, err := ii.TableSet()
	logutil.FatalErr(err)

	err = ts.WriteXLSX("githubissues.xlsx")
	logutil.FatalErr(err)

	fmt.Println("DONE")
}
