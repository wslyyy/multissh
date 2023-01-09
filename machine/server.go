package machine

import (
	"errors"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"multissh/auth"
	"multissh/logs"
	"multissh/push"
	"multissh/tools"
	"net"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

var (
	PASSWORD_SOURCE = "db"
	//PASSWORD_SOURCE   = "web"

	NO_PASSWORD = "GET PASSWORD ERROR\n"

	log = logs.NewLogger()
)

const (
	NO_EXIST = "0"
	IS_FILE = "1"
	IS_DIR = "2"
)

type Server struct {
	Ip                string
	Port              string
	User              string
	Pwd               string
	sshPrivateKeyPath string
	Action            string
	Cmd               string
	FileName          string
	RemotePath        string
	Force             bool
	Timeout           int
}

type PushConfig struct {
	Src string
	Dst string
}

type Result struct {
	Ip     string
	Cmd    string
	Result string
	Err    error
}

func NewCmdServer(ip, port, user, pwd, sshPrivateKeyPath, action, cmd string, force bool, timeout int) *Server {
	server := &Server{
		Ip:                ip,
		Port:              port,
		User:              user,
		Action:            action,
		Cmd:               cmd,
		Pwd:               pwd,
		sshPrivateKeyPath: sshPrivateKeyPath,
		Force:             force,
		Timeout:           timeout,
	}
	if pwd == "" {
		server.SetPwd()
	}
	return server
}

func NewPushServer(ip, port, user, pwd, sshPrivateKeyPath, action, file, rpath string, force bool, timeout int) *Server {
	rfile := path.Join(rpath, path.Base(file))
	cmd := createShell(rfile)
	server := &Server{
		Ip: ip,
		Port: port,
		User: user,
		Pwd: pwd,
		sshPrivateKeyPath: sshPrivateKeyPath,
		Action: action,
		FileName: file,
		RemotePath: rpath,
		Cmd: cmd,
		Force: force,
		Timeout: timeout,
	}
	if pwd == "" {
		server.SetPwd()
	}
	return server
}

func (s *Server) getSSH(mode string) (client *ssh.Client, err error) {
	switch mode {
	case "P":
		client, err = s.getSshClient()
	case "K":
		client, err = s.getSshClient_With_Private_Key(s.sshPrivateKeyPath)
	default:
		client, err = s.getSshClient()
	}
	return
}

func (s *Server) getSshClient_With_Private_Key(key_path string) (client *ssh.Client, err error) {
	if !tools.FileExists(key_path) {
		return nil, errors.New("ssh-key file not found!")
	}
	key, err := ioutil.ReadFile(key_path)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout:         (time.Duration(s.Timeout)) * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	ip_port := s.Ip + ":" + s.Port
	client, err = ssh.Dial("tcp", ip_port, sshConfig)
	return
}

func (s *Server) getSshClient() (client *ssh.Client, err error) {
	authMethods := []ssh.AuthMethod{}
	keyboardinteractiveChallenge := func(user,
		instruction string,
		questions []string,
		echos []bool,
	) (answers []string, err error) {

		if len(questions) == 0 {
			return []string{}, nil
		}

		for i, question := range questions {
			log.Debug("SSH Question %d: %s", i+1, question)
		}

		answers = make([]string, len(questions))
		for i := range questions {
			yes, _ := regexp.MatchString("*yes*", questions[i])
			if yes {
				answers[i] = "yes"
			} else {
				answers[i] = s.Pwd
			}
		}
		return answers, nil
	}
	authMethods = append(authMethods, ssh.KeyboardInteractive(keyboardinteractiveChallenge))
	authMethods = append(authMethods, ssh.Password(s.Pwd))

	sshConfig := &ssh.ClientConfig{
		User: s.User,
		Auth: authMethods,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: (time.Duration(s.Timeout)) * time.Second,
	}

	ip_port := s.Ip + ":" + s.Port
	client, err = ssh.Dial("tcp", ip_port, sshConfig)
	return
}

func (s *Server) SetPwd() {
	pwd, err := auth.GetPassword(PASSWORD_SOURCE, s.Ip, s.User)
	if err != nil {
		s.Pwd = NO_PASSWORD
		return
	}
	s.Pwd = pwd
}

func (s *Server) SetCmd(cmd string) {
	s.Cmd = cmd
}

func (s *Server) SRunCmd() Result {
	rs := Result{
		Ip:  s.Ip,
		Cmd: s.Cmd,
	}

	if s.Pwd == NO_PASSWORD {
		rs.Err = errors.New(NO_PASSWORD)
		return rs
	}

	client, err := s.getSshClient()
	if err != nil {
		rs.Err = err
		return rs
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		rs.Err = err
		return rs
	}
	defer session.Close()

	cmd := s.Cmd
	bs, err := session.CombinedOutput(cmd)
	if err != nil {
		rs.Err = err
		return rs
	}
	rs.Result = string(bs)
	return rs
}

func (s *Server) SRunCmd_With_Private_Key() Result {
	rs := Result{
		Ip:  s.Ip,
		Cmd: s.Cmd,
	}
	client, err := s.getSshClient_With_Private_Key(s.sshPrivateKeyPath)
	if err != nil {
		rs.Err = err
		return rs
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		rs.Err = err
		return rs
	}
	defer session.Close()

	cmd := s.Cmd
	bs, err := session.CombinedOutput(cmd)
	if err != nil {
		rs.Err = err
		return rs
	}
	rs.Result = string(bs)
	return rs
}

func (s *Server) PRunCmd(crs chan Result) {
	rs := s.SRunCmd()
	crs <- rs
}

func (s *Server) PRunCmd_With_Private_Key(crs chan Result) {
	rs := s.SRunCmd_With_Private_Key()
	crs <- rs
}

func (s *Server) PRunPushChoose(mode string, crs chan Result) {
	cmd := "push " + s.FileName + " to " + s.Ip + ":" + s.RemotePath
	rs := Result{
		Ip: s.Ip,
		Cmd: cmd,
	}
	result := s.RunPushDir(mode)
	if result != nil {
		rs.Err = result
	} else {
		rs.Result = cmd + " ok\n"
	}
	crs <- rs
}

func (s *Server) RunPushDir(mode string) (err error) {
	re := strings.TrimSpace(s.checkRemoteFile(mode))
	log.Debug("server.checkRemoteFile()=%s\n", re)

	// 远程机器存在同名文件
	if re == IS_FILE && s.Force == false {
		errString := "<ERROR>\nRemote Server's " + s.RemotePath + " has the same file " + s.FileName + "\nYou can use `-f` option force to cover the remote file.\n</ERROR>\n"
		return errors.New(errString)
	}

	rfile := s.RemotePath
	cmd := createShell(rfile)
	s.SetCmd(cmd)
	re = strings.TrimSpace(s.checkRemoteFile(mode))
	log.Debug("server.checkRemoteFile()=%s\n", re)

	if re != IS_DIR {
		errString := "[" + s.Ip + ":" + s.RemotePath + "] does not exist or not a dir\n"
		return errors.New(errString)
	}
	client, err := s.getSSH(mode)
	if err != nil {
		return err
	}
	defer client.Close()

	filename := s.FileName
	fi, err := os.Stat(filename)
	if err != nil {
		log.Debug("open source file %s error\n", filename)
		return err
	}
	push := push.NewPush(client)
	if fi.IsDir() {
		err = push.PushDir(filename, s.RemotePath)
		return err
	}
	err = push.PushFile(filename, s.RemotePath)
	return err
}

func (s *Server) checkRemoteFile(mode string) (result string) {
	re := Result{}
	switch mode {
	case "P":
		re = s.SRunCmd()
	case "K":
		re = s.SRunCmd_With_Private_Key()
	default:
		re = s.SRunCmd()
	}
	result = re.Result
	return
}

func createShell(file string) string {
	s1 := "bash << EOF \n"
	s2 := "if [[ -f " + file + " ]];then \n"
	s3 := "echo '1'\n"
	s4 := "elif [[ -d " + file + " ]];then \n"
	s5 := `echo "2"
else
echo "0"
fi
EOF`
	cmd := s1 + s2 + s3 + s4 + s5
	return cmd
}
