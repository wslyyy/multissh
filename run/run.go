package run

import (
	"errors"
	"fmt"
	"multissh/config"
	"multissh/logs"
	"multissh/machine"
	"multissh/output"
	"sync"
)

var (
	log = logs.NewLogger()
)

type CommonUser struct {
	user              string
	port              string
	pwd               string
	sshPrivateKeyPath string
	force             bool
	encflag           bool
}

func NewUser(user, port, pwd, sshPrivateKeyPath string, force, encflag bool) *CommonUser {
	return &CommonUser{
		user:              user,
		port:              port,
		pwd:               pwd,
		sshPrivateKeyPath: sshPrivateKeyPath,
		force:             force,
		encflag:           encflag,
	}
}

func SingleRun(host, cmd string, cu *CommonUser, force bool, timeout int) {
	r := machine.Result{}
	server := machine.NewCmdServer(host, cu.port, cu.user, cu.pwd, cu.sshPrivateKeyPath, "cmd", cmd, force, timeout)
	if cu.sshPrivateKeyPath == "" {
		r = server.SRunCmd()
	} else {
		r = server.SRunCmd_With_Private_Key()
	}
	output.Print(r)
}

func ServersRun(cmd string, cu *CommonUser, wt *sync.WaitGroup, crs chan machine.Result, ipFile string, ccons chan struct{}, safe bool, timeout int) {
	hosts, err := parseIpfile(ipFile, cu)
	if err != nil {
		log.Error("Parse %s error, error=%s", ipFile, err)
		return
	}

	ips := config.GetIps(hosts)

	log.Info("[servers]=%v", ips)
	fmt.Printf("[servers]=%v\n", ips)

	ls := len(hosts)

	// 并发度如果ccons==1串行执行，可以暂停
	if cap(ccons) == 1 {
		log.Info("由于并发参数ccons设置为1，已退化为串行执行")
		for _, h := range hosts {
			r := machine.Result{}
			server := machine.NewCmdServer(h.Ip, h.Port, h.User, h.Pwd, h.SshPrivateKeyPath, "cmd", cmd, cu.force, timeout)
			if h.SshPrivateKeyPath == "" {
				r = server.SRunCmd()
			} else {
				r = server.SRunCmd_With_Private_Key()
			}
			if r.Err != nil && safe {
				log.Debug("%s执行出错", h.Ip)
				output.Print(r)
				break
			} else {
				output.Print(r)
			}
		}
	} else {
		log.Info("并行执行，整体并发度为%v\n", cap(ccons))
		go output.PrintResults2(crs, ls, wt, ccons, timeout)

		for _, h := range hosts {
			ccons <- struct{}{}
			server := machine.NewCmdServer(h.Ip, h.Port, h.User, h.Pwd, h.SshPrivateKeyPath, "cmd", cmd, cu.force, timeout)
			wt.Add(1)
			if h.SshPrivateKeyPath == "" {
				go server.PRunCmd(crs)
			} else {
				go server.PRunCmd_With_Private_Key(crs)
			}
		}
	}
}

func parseIpfile(ipFile string, cu *CommonUser) ([]config.Host, error) {
	hosts, err := config.ParseIps(ipFile, cu.encflag)
	if err != nil {
		log.Error("parse Ip File %s error, %s\n", ipFile, err)
		return hosts, err
	}

	if len(hosts) == 0 {
		return hosts, errors.New(ipFile + " is null")
	}
	hosts = config.PaddingHosts(hosts, cu.port, cu.user, cu.pwd, cu.sshPrivateKeyPath)
	return hosts, nil
}
