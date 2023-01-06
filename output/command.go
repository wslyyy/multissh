package output

import (
	"fmt"
	"multissh/machine"
	"sync"
	"time"
)

const (
	TIMEOUT = 4500
)

func Print(res machine.Result) {
	fmt.Printf("ip=%s\n", res.Ip)
	fmt.Printf("command=%s\n", res.Cmd)
	if res.Err != nil {
		fmt.Printf("return=1\n")
		fmt.Printf("%s\n", res.Err)
	} else {
		fmt.Printf("return 0\n")
		fmt.Printf("%s\n", res.Result)
	}
	fmt.Println("--------------------------------------------------")
}

func PrintResults2(crs chan machine.Result, ls int, wt *sync.WaitGroup, ccons chan struct{}, timeout int) {
	if timeout == 0 {
		timeout = TIMEOUT
	}

	for i := 0; i < ls; i++ {
		select {
		case rs := <-crs:
			//PrintResult(rs.Ip, rs.Cmd, rs.Result)
			Print(rs)
		case <-time.After(time.Second * time.Duration(timeout)):
			fmt.Printf("getSSHClient error,SSH-Read-TimeOut,Timeout=%ds", timeout)
		}
		wt.Done()
		<-ccons
	}
}
