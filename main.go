package main

import "github.com/phamdaiminhquan/vibe-devops/cmd"

var (
	version = "dev"
	commit  = ""
	date    = ""
)

func main() {
	cmd.SetVersionInfo(version, commit, date)
	cmd.Execute()
}
