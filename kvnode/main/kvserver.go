package main

import (
	"flag"
	"fmt"
	"github.com/sniperHW/flyfish/conf"
	kvnode "github.com/sniperHW/flyfish/kvnode"
	futil "github.com/sniperHW/flyfish/util"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	go func() {
		http.ListenAndServe("0.0.0.0:8899", nil)
	}()

	cluster := flag.String("cluster", "http://127.0.0.1:12379", "comma separated cluster peers")
	id := flag.Int("id", 1, "node ID")
	config := flag.String("config", "config.toml", "config")

	flag.Parse()

	futil.Must(nil, conf.LoadConfig(*config))

	kvnode.InitLogger()

	node := kvnode.NewKvNode()

	err := node.Start(id, cluster)
	if nil == err {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT) //监听指定信号
		_ = <-c                          //阻塞直至有信号传入
		node.Stop()
		fmt.Println("server stop")
	} else {
		fmt.Println(err)
	}
}