package config

import (
	"bufio"
	"errors"
	"io"
	"multissh/enc"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var (
	rzkey = enc.GetKey()
)

type Host struct {
	Ip   string
	Port string
	User string
	Pwd               string
	SshPrivateKeyPath string
}

func GetIps(h []Host) []string {
	ips := make([]string, 0)
	for _, v := range h {
		ips = append(ips, v.Ip)
	}
	return ips
}

func PaddingHosts(h []Host, port, user, pwd, sshPrivateKeyPath string) []Host {
	hosts := make([]Host, 0)
	for _, v := range h {
		if v.Port == "" {
			v.Port = port
		}
		if v.User == "" {
			v.User = user
		}
		if v.Pwd == "" {
			v.Pwd = pwd
		}
		if v.SshPrivateKeyPath == "" {
			v.SshPrivateKeyPath = sshPrivateKeyPath
		}

		hosts = append(hosts, v)
	}
	return hosts
}
func ParseIps(ipfile string, eflag bool) ([]Host, error) {
	hosts := make([]Host, 0)

	AppPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}

	configfile := ""

	if filepath.IsAbs(ipfile) {
		configfile = ipfile
	} else {
		configfile = filepath.Join(AppPath, ipfile)
	}

	f, err := os.Open(configfile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := bufio.NewReader(f)

	for {
		s, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF && s != "" {
				goto Lable
			} else {
				return hosts, nil
			}
			return hosts, err
		}
	Lable:
		line := strings.TrimSpace(s)
		if line == "" || line[0] == '#' {
			continue
		}
		h, err := parseLine(line, eflag)
		if err != nil {
			continue
		}
		hosts = append(hosts, h)
	}
	return hosts, err
}

func parseLine(s string, eflag bool) (Host, error) {
	host := Host{}
	line := strings.TrimSpace(s)

	if line[0] == '#' {
		return host, errors.New("comment line")
	}
	if line == "" {
		return host, errors.New("null line")
	}

	fields := strings.Split(line, "|")
	mode := strings.TrimSpace(fields[0])
	hname := strings.TrimSpace(fields[1])
	_, err := net.LookupHost(hname)
	if err != nil {
		return host, errors.New("ill ip")
	}
	lens := len(fields)

	// password
	switch mode {
	case "p":
		switch lens {
		case 2:
			host.Ip = hname
		case 3:
			host.Ip = hname
			host.Port = strings.TrimSpace(fields[2])
		case 4:
			host.Ip = hname
			host.Port = strings.TrimSpace(fields[2])
			host.User = strings.TrimSpace(fields[3])
		case 5:
			host.Ip = hname
			host.Port = strings.TrimSpace(fields[2])
			host.User = strings.TrimSpace(fields[3])
			pass := strings.TrimSpace(fields[4])
			if eflag && pass != "" {
				text, err := decrypt(pass, rzkey)
				if err != nil {
					return host, errors.New("decrypt the password error")
				}
				host.Pwd = string(text)
			} else {
				host.Pwd = pass
			}
		default:
			return host, errors.New("format err")
		}
	case "k":
		switch lens {
		case 2:
			host.Ip = hname
		case 3:
			host.Ip = hname
			host.Port = strings.TrimSpace(fields[2])
		case 4:
			host.Ip = hname
			host.Port = strings.TrimSpace(fields[2])
			host.User = strings.TrimSpace(fields[3])
		case 5:
			host.Ip = hname
			host.Port = strings.TrimSpace(fields[2])
			host.User = strings.TrimSpace(fields[3])
			host.SshPrivateKeyPath = strings.TrimSpace(fields[4])

		default:
			return host, errors.New("format err")
		}
	}
	return host, nil
}

func decrypt(pass string, key []byte) ([]byte, error) {
	skey := key[:16]
	return enc.AesDecEncode(pass, skey)
}
