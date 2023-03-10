package run

import (
	"errors"
	"fmt"
	"multissh/config"
	"multissh/logs"
	"multissh/machine"
	"multissh/output"
	"path/filepath"
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

func SinglePull(host string, cu *CommonUser, src, dst string, timeout int) {
	server := machine.NewPullServer(host, cu.port, cu.user, cu.pwd, cu.sshPrivateKeyPath, "pull", src, dst, cu.force, timeout)
	var mode string
	if cu.sshPrivateKeyPath != "" {
		mode = "K"
	} else {
		mode = "P"
	}
	err := server.RunPull(mode)
	output.PrintPullResult(host, src, dst, err)
}

func ServerPull(src, dst string, cu *CommonUser, wt *sync.WaitGroup, crs chan machine.Result, ipFile string, ccons chan struct{}, timeout int) {
	hosts, err := parseIpfile(ipFile, cu)
	if err != nil {
		log.Error("Parse %s error, error=%s", ipFile, err)
		return
	}
	ips := config.GetIps(hosts)
	log.Info("[servers]=%v", ips)
	fmt.Printf("[servers]=%v\n", ips)

	ls := len(hosts)
	go output.PrintResults2(crs, ls, wt, ccons, timeout)

	for _, h := range hosts {
		ip := h.Ip
		localPath := filepath.Join(src, ip)
		ccons <- struct{}{}
		server := machine.NewPullServer(h.Ip, h.Port, h.User, h.Pwd, h.SshPrivateKeyPath, "pull", localPath, dst, cu.force, timeout)
		wt.Add(1)
		var mode string
		if h.SshPrivateKeyPath == "" {
			mode = "P"
		} else {
			mode = "K"
		}
		go server.PRunPullChoose(mode, crs)
	}
}

func SinglePush(host, src, dst string, cu *CommonUser, force bool, timeout int) {
	server := machine.NewPushServer(host, cu.port, cu.user, cu.pwd, cu.sshPrivateKeyPath, "push", src, dst, force, timeout)
	cmd := "push " + server.FileName + " to " + server.Ip + ":" + server.RemotePath

	rs := machine.Result{
		Ip:  server.Ip,
		Cmd: cmd,
	}
	var mode string
	if cu.sshPrivateKeyPath != "" {
		mode = "K"
	} else {
		mode = "P"
	}
	err := server.RunPushDir(mode)
	if err != nil {
		rs.Err = err
	} else {
		rs.Result = cmd + " ok\n"
	}
	output.Print(rs)
}

func ServersPush(src, dst string, cu *CommonUser, wt *sync.WaitGroup, crs chan machine.Result, ipFile string, ccons chan struct{}, timeout int) {
	hosts, err := parseIpfile(ipFile, cu)
	if err != nil {
		log.Error("Parse %s error, error=%s", ipFile, err)
		return
	}

	ips := config.GetIps(hosts)

	log.Info("[servers]=%v", ips)
	fmt.Printf("[servers]=%v\n", ips)

	ls := len(hosts)
	go output.PrintResults2(crs, ls, wt, ccons, timeout)

	for _, h := range hosts {
		ccons <- struct{}{}
		server := machine.NewPushServer(h.Ip, h.Port, h.User, h.Pwd, h.SshPrivateKeyPath, "push", src, dst, cu.force, timeout)
		wt.Add(1)
		var mode string
		if h.SshPrivateKeyPath == "" {
			mode = "P"
		} else {
			mode = "K"
		}
		go server.PRunPushChoose(mode, crs)
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

	// ???????????????ccons==1???????????????????????????
	if cap(ccons) == 1 {
		log.Info("??????????????????ccons?????????1???????????????????????????")
		for _, h := range hosts {
			r := machine.Result{}
			server := machine.NewCmdServer(h.Ip, h.Port, h.User, h.Pwd, h.SshPrivateKeyPath, "cmd", cmd, cu.force, timeout)
			if h.SshPrivateKeyPath == "" {
				r = server.SRunCmd()
			} else {
				r = server.SRunCmd_With_Private_Key()
			}
			if r.Err != nil && safe {
				log.Debug("%s????????????", h.Ip)
				output.Print(r)
				break
			} else {
				output.Print(r)
			}
		}
	} else {
		log.Info("?????????????????????????????????%v\n", cap(ccons))
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
