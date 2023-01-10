package main

import (
	"flag"
	"fmt"
	"multissh/enc"
	"multissh/help"
	"multissh/logs"
	"multissh/machine"
	"multissh/run"
	"multissh/tools"
	"path/filepath"
	"sync"
)

const AppVersion = "multissh 0.1"

var (
	// common options
	port              = flag.String("P", "22", "ssh port")
	host              = flag.String("h", "", "ssh ip")
	user              = flag.String("u", "root", "ssh user")
	pwd               = flag.String("p", "", "ssh password")
	sshPrivateKeyPath = flag.String("K", "", "ssh private key path")
	prunType          = flag.String("t", "cmd", "running mode: cmd|push|pull")

	// batch option running options
	ipFile = flag.String("i", "ip.txt", "ip file when batch running mode")
	cons   = flag.Int("c", 30, "the number of concurrency when b")

	// safe options
	encFlag   = flag.Bool("e", false, "password is Encrypted")
	force     = flag.Bool("f", false, "force to run even if it is not safe")
	psafe     = flag.Bool("s", false, "if -s is setting, multissh will exit when error occurs")
	pkey      = flag.String("key", "", "aes key for password decrypt and encryption")
	blackList = []string{"rm", "mkfs", "mkfs.ext3", "make.ext2", "make.ext4", "make2fs", "shutdown", "reboot", "init", "dd"}

	// log options
	plogLevel = flag.String("l", "info", "log level (debug|info|warn|error)")
	plogPath  = flag.String("logpath", "./log/", "logfile path")
	logFile   = "multissh.log"
	log       = logs.NewLogger()

	// version options
	pversion = flag.Bool("version", false, "multissh version")
	ptimeout = flag.Int("timeout", 10, "ssh timeout setting")
)

func main() {
	usage := func() {
		fmt.Println(help.Help)
	}

	flag.Parse()

	// version
	if *pversion {
		fmt.Println(AppVersion)
		return
	}

	if *pkey != "" {
		enc.SetKey([]byte(*pkey))
	}

	if flag.NArg() < 1 || *prunType == "" || flag.Arg(0) == "" {
		usage()
		return
	}

	if err := initLog(); err != nil {
		fmt.Printf("init log error:%s\n", err)
		return
	}

	// 异步日志，需要最后刷新和关掉
	defer func() {
		log.Flush()
		log.Close()
	}()

	log.Debug("parse flag ok, init log setting ok.")

	switch *prunType {
	case "cmd":
		if flag.NArg() != 1 {
			usage()
			return
		}

		cmd := flag.Arg(0)

		if flag := tools.CheckSafe(cmd, blackList); !flag && *force == false {
			fmt.Printf("Dangerous command in %s", cmd)
			fmt.Printf("You can use the -f option to force to excute")
			log.Error("Dangerous command in %s", cmd)
			return
		}

		puser := run.NewUser(*user, *port, *pwd, *sshPrivateKeyPath, *force, *encFlag)
		log.Info("multissh -t=cmd cmd=[%s]", cmd)

		if *host != "" {
			log.Info("[servers]=%s", *host)
			run.SingleRun(*host, cmd, puser, *force, *ptimeout)
		} else {
			cr := make(chan machine.Result)
			ccons := make(chan struct{}, *cons)
			wg := &sync.WaitGroup{}
			run.ServersRun(cmd, puser, wg, cr, *ipFile, ccons, *psafe, *ptimeout)
			wg.Wait()
		}
	case "push":
		if flag.NArg() != 2 {
			usage()
			return
		}

		src := flag.Arg(0)
		dst := flag.Arg(1)
		log.Info("multissh -t=push local-file=%s, remote-path=%s", src, dst)

		puser := run.NewUser(*user, *port, *pwd, *sshPrivateKeyPath, *force, *encFlag)
		if *host != "" {
			log.Info("[servers]=%s", *host)
			run.SinglePush(*host, src, dst, puser, *force, *ptimeout)
		} else {
			cr := make(chan machine.Result)
			ccons := make(chan struct{}, *cons)
			wg := &sync.WaitGroup{}
			run.ServersPush(src, dst, puser, wg, cr, *ipFile, ccons, *ptimeout)
			wg.Wait()
		}
	case "pull":
		if flag.NArg() != 2 {
			usage()
			return
		}

		// 本地目录
		src := flag.Arg(1)
		// 远程文件
		dst := flag.Arg(0)
		log.Info("multissh -t=pull remote-file=%s local-path=%s", dst, src)

		puser := run.NewUser(*user, *port, *pwd, *sshPrivateKeyPath, *force, *encFlag)
		if *host != "" {
			log.Info("[servers]=%s", *host)
			run.SinglePull(*host, puser, src, dst, *ptimeout)
		} else {
			cr := make(chan machine.Result)
			ccons := make(chan struct{}, *cons)
			wg := &sync.WaitGroup{}
			run.ServerPull(src, dst, puser, wg, cr, *ipFile, ccons, *ptimeout)
			wg.Wait()
		}
	default:
		usage()
	}
}

func initLog() error {
	switch *plogLevel {
	case "debug":
		log.SetLevel(logs.LevelDebug)
	case "error":
		log.SetLevel(logs.LevelError)
	case "info":
		log.SetLevel(logs.LevelInfo)
	case "warn":
		log.SetLevel(logs.LevelWarn)
	default:
		log.SetLevel(logs.LevelInfo)
	}

	logpath := *plogPath
	err := tools.MakePath(logpath)
	if err != nil {
		return err
	}

	logname := filepath.Join(logpath, logFile)
	logstring := `{"filename":"` + logname + `"}`

	err = log.SetLogger("file", logstring)
	if err != nil {
		return err
	}

	// 开启日志异步提升性能
	log.Async()
	return nil
}
