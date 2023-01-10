package pull

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

type Pull struct {
	client *ssh.Client
}

func NewPull(client *ssh.Client) *Pull {
	return &Pull{
		client: client,
	}
}

func (pull *Pull) PullFile(srcPath, targetFile string) error {
	session, err := pull.client.NewSession()
	if err != nil {
		log.Fatalln("Failed to create session: " + err.Error())
		return err
	}
	defer session.Close()

	go func() {
		iw, err := session.StdinPipe()
		if err != nil {
			log.Fatalln("Failed to create input pipe: ", err.Error())
		}
		or, err := session.StdoutPipe()
		if err != nil {
			log.Fatalln("Failed to create output pipe: ", err.Error())
		}
		fmt.Fprint(iw, "\x00")

		sr := bufio.NewReader(or)
		localFile := path.Join(srcPath, path.Base(targetFile))
		src, srcErr := os.Create(localFile)
		if srcErr != nil {
			log.Fatalln("Failed to create source file: " + srcErr.Error())
		}
		if controlString, ok := sr.ReadString('\n'); ok == nil && strings.HasPrefix(controlString, "C") {
			fmt.Fprint(iw, "\x00")
			controlParts := strings.Split(controlString, " ")
			size, _ := strconv.ParseInt(controlParts[1], 10, 64)
			if n, ok := io.CopyN(src, sr, size); ok != nil || n < size {
				fmt.Fprint(iw, "\x02")
				return
			}
			sr.Read(make([]byte, 1))
		}
		fmt.Fprint(iw, "\x00")
	}()

	if err := session.Run(fmt.Sprintf("scp -f %s", targetFile)); err != nil {
		log.Fatalln("Failed to run: " + err.Error())
		return err
	}
	return nil
}
