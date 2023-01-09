package push

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

const (
	PUSH_BEGIN_FILE = "C"
	PUSH_BEGIN_FOLDER = "D"
	PUSH_BEGIN_END_FOLDER = "0"
	PUSH_END_FOLDER = "E"
	PUSH_END = "\x00"
)

type Push struct {
	client *ssh.Client
}

func GetPerm(f *os.File) (perm string) {
	fileStat, _ := f.Stat()
	mod := fileStat.Mode()
	if mod > (1 << 9) {
		mod = mod % (1 << 9)
	}
	return fmt.Sprintf("%#o", uint32(mod))
}

func lsDir(w io.WriteCloser, dir string) {
	fi, _ := ioutil.ReadDir(dir)
	for _, f := range fi {
		if f.IsDir() {
			folderSrc, _ := os.Open(path.Join(dir, f.Name()))
			defer folderSrc.Close()
			fmt.Fprintln(w, PUSH_BEGIN_FOLDER + GetPerm(folderSrc), PUSH_BEGIN_END_FOLDER, f.Name())
			lsDir(w, path.Join(dir, f.Name()))
			fmt.Fprintln(w, PUSH_END_FOLDER)
		} else {
			prepareFile(w, path.Join(dir, f.Name()))
		}
	}
}

func prepareFile(w io.WriteCloser, src string) {
	fileSrc, srcErr := os.Open(src)
	defer fileSrc.Close()
	if srcErr != nil {
		log.Fatalln("Failed to open source file: " + srcErr.Error())
	}
	srcStat, statErr := fileSrc.Stat()
	if statErr != nil {
		log.Fatalln("Failed to stat file: " + statErr.Error())
	}
	fmt.Fprintln(w, PUSH_BEGIN_FILE + GetPerm(fileSrc), srcStat.Size(), filepath.Base(src))
	io.Copy(w, fileSrc)
	fmt.Fprint(w, PUSH_END)
}

func NewPush(client *ssh.Client) *Push{
	return &Push{
		client: client,
	}
}

func (push *Push) PushFile(src string, dst string) error {
	session, err := push.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		prepareFile(w, src)
	}()
	if err := session.Run("/usr/bin/scp -rt " + dst); err != nil {
		return err
	}
	return nil
}

func (push *Push) PushDir(src string, dst string) error {
	session, err := push.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		folderSrc, _ := os.Open(src)
		fmt.Fprintln(w, PUSH_BEGIN_FOLDER + GetPerm(folderSrc), PUSH_BEGIN_END_FOLDER, filepath.Base(src))
		lsDir(w, src)
		fmt.Fprintln(w, PUSH_END_FOLDER)
	}()
	if err := session.Run("/usr/bin/scp -qrt " + dst); err != nil {
		return err
	}
	return  nil
}
