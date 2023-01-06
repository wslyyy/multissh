package tools

import (
	"errors"
	"os"
	"strings"
)

func FileExists(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	} else {
		return !f.IsDir()
	}
}

func PathExists(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	} else {
		return f.IsDir()
	}
}

func MakePath(path string) error {
	if FileExists(path) {
		return errors.New(path + " is a normal file, not a dir")
	}
	if !PathExists(path) {
		return os.MkdirAll(path, os.ModePerm)
	} else {
		return nil
	}
}

func CheckSafe(cmd string, blacks []string) bool {
	lcmd := strings.ToLower(cmd)
	cmds := strings.Split(lcmd, " ")
	for _, ds := range cmds {
		for _, bk := range blacks {
			if ds == bk {
				return false
			}
		}
	}
	return true
}
