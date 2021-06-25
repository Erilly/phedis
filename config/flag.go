package config

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

const APPNAME = "phedis"

var(
	COMMITID string

	conf	string
	Logfile	string

	rev	bool
)

func initFlag(){
	flag.StringVar(&conf, "conf", "./conf/phedis_config.toml", "common file")
	flag.StringVar(&Logfile, "log", "", "log file")
	flag.BoolVar(&rev, "rev", false, "reversion")

	flag.Parse()

	if rev {
		fmt.Sprintf("\"%s build on %s [%s, %s] commit_id(%s)\n", APPNAME, runtime.Version(),runtime.GOOS,runtime.GOARCH,COMMITID)
		os.Exit(0);
	}
}
