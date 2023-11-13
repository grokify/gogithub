package main

import (
	"context"
	"fmt"

	"github.com/grokify/gogithub"
	"github.com/grokify/mogo/fmt/fmtutil"
	"github.com/grokify/mogo/log/logutil"
)

func main() {
	c := gogithub.NewClient(nil)
	r1, _, err := c.SearchOpenPullRequests(context.Background(), "grokify", nil)
	logutil.FatalErr(err)
	fmtutil.MustPrintJSON(r1)

	iss := gogithub.Issues(r1.Issues)
	tbl, err := iss.Table()
	logutil.FatalErr(err)

	err = tbl.WriteXLSX("githubissues.xlsx", "issues")
	logutil.FatalErr(err)

	fmt.Println("DONE")
}
