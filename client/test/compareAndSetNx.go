package main

import (
	kclient "flyfish/client"
	"flyfish/errcode"
	"fmt"
	"github.com/sniperHW/kendynet"
	"github.com/sniperHW/kendynet/golog"
)

func CompareAndSetNx(c *kclient.Client) {
	set := c.CompareAndSetNx("counter", "test_counter3", "c", 10, 10)
	set.Exec(func(ret *kclient.SliceResult) {

		if ret.ErrCode != errcode.ERR_OK {
			fmt.Println(errcode.GetErrorStr(ret.ErrCode), ret)
		} else {
			fmt.Println("set ok")
		}
	})
}

func main() {

	kclient.InitLogger(golog.NewOutputLogger("log", "flyfish client", 1024*1024*50), "error")

	services := []string{"127.0.0.1:10012"}
	c := kclient.OpenClient(services) //eventQueue)

	CompareAndSetNx(c)

	//eventQueue.Run()

	sigStop := make(chan bool)
	_, _ = <-sigStop
}
