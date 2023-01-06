package machine

import (
	"errors"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"multissh/auth"
	"multissh/logs"
	"multissh/tools"
	"net"
	"regexp"
	"time"
)

var (
	PASSWORD_SOURCE = "db"
	//PASSWORD_SOURCE   = "web"

	NO_PASSWORD = "GET PASSWORD ERROR\n"

	log = logs.NewLogger()
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
