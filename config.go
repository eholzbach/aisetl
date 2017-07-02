package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"runtime"
	"strings"
)

func config() (string, bool, string, string) {

	var fwd string
	var lsn string
	var rds string

	user, err := user.Current()
	if err != nil {
		fmt.Println(err)
	}

	home := user.HomeDir
	const conf string = "/.aisetl.conf"
	d := home + conf

	var e string
	if _, err := os.Stat(d); os.IsNotExist(err) {
		if runtime.GOOS == "freebsd" {
			e = "/usr/local/etc/aisetl.conf"
		} else {
			e = "/etc/aisetl.conf"
		}
	} else {
		e = d
	}

	if _, err := os.Stat(e); err == nil {

		f, err := ioutil.ReadFile(e)

		if err != nil {
			panic(err)
		}

		g := strings.Split(string(f), "\n")

		for _, v := range g {
			t := strings.TrimSpace(v)
			if strings.HasPrefix(t, "forward ") {
				fwd = strings.TrimPrefix(t, "forward ")
			}
			if strings.HasPrefix(t, "listen ") {
				lsn = strings.TrimPrefix(t, "listen ")
			}
			if strings.HasPrefix(t, "redis ") {
				rds = strings.TrimPrefix(t, "redis ")
			}
		}
	} else {
		fmt.Println("config not found, exiting")
		os.Exit(0)
	}
	if len(lsn) <= 1 {
		lsn = "127.0.0.1:10110"
	}
	if len(rds) <= 1 {
		fmt.Println("redis server config not found, exiting")
		os.Exit(0)
	}

	if len(fwd) > 0 {
		return rds, true, fwd, lsn
	} else {
		return rds, false, fwd, lsn
	}
}
