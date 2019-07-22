package main

import (
	"fmt"

	"github.com/hqpko/hutils"

	"flyfish"
	"flyfish/conf"

	//"github.com/sniperHW/kendynet/golog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	hutils.Must(nil, conf.InitConfig(os.Args[1]))
	config := conf.DefConfig

	flyfish.InitLogger()

	if !flyfish.InitTableConfig() {
		fmt.Println("InitTableConfig failed")
		return
	}

	flyfish.InitProcessUnit(config.DBConfig.DbHost, config.DBConfig.DbPort, config.DBConfig.DbDataBase, config.DBConfig.DbUser, config.DBConfig.DbPassword)
	flyfish.RedisInit(config.Redis.RedisHost, config.Redis.RedisPort, config.Redis.RedisPassword)
	flyfish.Recover()

	go func() {
		http.ListenAndServe("0.0.0.0:8899", nil)
	}()

	err := flyfish.StartTcpServer("tcp", fmt.Sprintf("%s:%d", config.ServiceHost, config.ServicePort))
	if nil == err {
		fmt.Println("flyfish start:", fmt.Sprintf("%s:%d", config.ServiceHost, config.ServicePort))
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT) //监听指定信号
		_ = <-c                          //阻塞直至有信号传入
		flyfish.Stop()
		fmt.Println("server stop")
	} else {
		fmt.Println(err)
	}
}
