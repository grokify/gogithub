package main

import (
	"github.com/grokify/gogithub/cliutil"
	"github.com/grokify/mogo/fmt/fmtutil"
	"github.com/grokify/mogo/log/logutil"
)

func main() {
	dir := "/Users/johnwang/go/src/github.com/grokify/grokify.github.io/docs"
	lines, err := cliutil.GitStatusShortLines(dir)
	logutil.FatalErr(err)
	fmtutil.MustPrintJSON(lines)

	rmlines, err := cliutil.GitRmDeletedLines(dir)
	logutil.FatalErr(err)
	fmtutil.MustPrintJSON(rmlines)

	err = cliutil.GitRmDeletedFile("todel.sh", dir)
	logutil.FatalErr(err)
}
