/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2017-2021 Kusakabe Si. All Rights Reserved.
 */

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"syscall"
	"time"

	nonSecureRand "math/rand"
	"net/http"
	_ "net/http/pprof"

	"github.com/bingbaga/AnyWhereEtherNet/gencfg"
	"github.com/bingbaga/AnyWhereEtherNet/path"
	"github.com/bingbaga/AnyWhereEtherNet/tap"
)

const (
	ExitSetupSuccess = 0
	ExitSetupFailed  = 1
)

var (
	tconfig      = flag.String("config", "", "Config path for the interface.")
	mode         = flag.String("mode", "", "Running mode. [super|edge|solve|gencfg]")
	printExample = flag.Bool("example", false, "Print example config")
	cfgmode      = flag.String("cfgmode", "", "Running mode for generated config. [none|super|p2p]")
	bind         = flag.String("bind", "linux", "UDP socket bind mode. [linux|std]\nYou may need std mode if you want to run Etherguard under WSL.")
	pprofaddr    = flag.String("pprof", "", "pprof listing address")
	version      = flag.Bool("version", false, "Show version")
	help         = flag.Bool("help", false, "Show this help")
)

func main() {
	flag.Parse()
	if *version {
		fmt.Printf("etherguard-go %s\n%s-%s\n%s\n\nA full mesh layer 2 VPN powered by Floyd Warshall algorithm.\nInformation available at https://github.com/bingbaga/AnyWhereEtherNet.\nCopyright (C) Kusakabe Si <si@kskb.eu.org>.\n", Version, runtime.GOOS, runtime.GOARCH, tap.VPP_SUPPORT)
		return
	}
	if *help {
		flag.Usage()
		return
	}

	nonSecureRand.Seed(time.Now().UnixNano())
	if *pprofaddr != "" {
		go func() {
			//内网可访问的pprof地址
			err := http.ListenAndServe(*pprofaddr, nil)
			if err != nil {
				panic(fmt.Errorf("pprof error: %v", err))
			}
		}()
	}

	var err error
	switch *mode {
	case "edge":
		err = Edge(*tconfig, *printExample, *bind)
	case "super":
		err = Super(*tconfig, *printExample, *bind)
	case "solve":
		err = path.Solve(*tconfig, *printExample)
	case "gencfg":
		switch *cfgmode {
		case "super":
			err = gencfg.GenSuperCfg(*tconfig, *printExample)
		case "static":
			err = gencfg.GenNMCfg(*tconfig, false, *printExample)
		case "p2p":
			err = gencfg.GenNMCfg(*tconfig, true, *printExample)
		default:
			err = fmt.Errorf("gencfg: generate config for %v mode are not implement", *cfgmode)
		}
	default:
		flag.Usage()
	}
	if err != nil {
		switch err := err.(type) {
		case syscall.Errno:
			os.Exit(int(err))
		default:
			fmt.Fprintf(os.Stderr, "Error :%v\n", err)
			os.Exit(1)
		}
	}
}
