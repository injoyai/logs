package main

import (
	"github.com/injoyai/logs"
	"log"
	"time"
)

func main() {

	logs.Red("red")
	logs.Green("green")
	logs.Blue("blue")
	logs.Yellow("yellow")
	logs.Magenta("magenta")
	logs.Cyan("cyan")
	//logs.SetFormatter(logs.TimeFormatter)
	//logs.Red("red")
	//logs.Green("green")
	//logs.Blue("blue")
	//logs.Yellow("yellow")
	//logs.Magenta("magenta")
	//logs.Cyan("cyan")

	logs.Info("测试xxln,xxf,write")
	logs.Debugf("不换行")
	logs.Debugf("换行\n")
	logs.Debugf("换行*2\n\n")
	logs.Debug("456")
	logs.Debug("789")
	logs.DefaultDebug.Write([]byte("123456\n"))
	logs.DefaultDebug.Println()
	logs.New("测试").Println(666)

	//===================测试TCP Client===================

	logs.Info("测试TCP Client")
	w, err := logs.NewTCPClient(":10089")
	logs.Err(err)
	w2, err := logs.NewTCPServer(10086)
	logs.Err(err)
	logs.AddWriter(logs.NewWriteColor(w), w2)
	<-time.After(time.Second * 5)

	logs.Stdout.EnableFilter()

	<-time.After(time.Second * 5)

	//===================测试Color===================

	logs.Info("测试Color")
	logs.Trace("trace")
	logs.Write("write")
	logs.Read("read")
	logs.Debug("Debug")
	logs.Err("Err")
	logs.Warn("Warn")
	logs.SetShowColor(false)
	logs.Debug("Debug")
	logs.Err("Err")
	logs.Warn("Warn")
	logs.SetShowColor(true)

	//===================测试Level===================

	logs.Info("测试Level")
	logs.Debug("Level Debug Before")
	logs.Info("Level Info Before")
	logs.Err("Level Err Before")
	logs.SetLevel(logs.LevelError)
	logs.Debug("Level Debug After")
	logs.Info("Level Info After")
	logs.Err("Level Err After")
	logs.SetLevel(logs.LevelAll)

	//===================测试Flag===================

	logs.Info("测试Flag")
	logs.DefaultFormatter.SetFlag(log.Ltime)
	logs.Read("read")
	logs.Debug("debug")
	logs.DefaultFormatter.SetFlag(log.Ldate | log.Ltime | log.Lshortfile)

	//===================测试Formatter===================

	logs.Info("测试Formatter")
	logs.SetFormatter(new(_format))
	logs.Debug("Format Debug")
	logs.Info("Format Info")
	logs.Err("Format Err")
	logs.SetFormatterWithTime()
	logs.Debug("Format Debug")
	logs.Info("Format Info")
	logs.Err("Format Err")
	logs.SetFormatter(logs.DefaultFormatter)

	//===================测试Panic和Fatal===================

	logs.Info("测试Panic和Fatal")
	func() {
		defer logs.Spend("总", "耗时")()
		defer logs.Spend()()

		<-time.After(time.Second * 5)
	}()

	testPanic()
	logs.Fatal("Fatal")
	logs.Err("结束")

}

func testPanic() {
	defer func() { recover() }()
	logs.Panic("Panic")
}

type _format struct{}

func (_format) Formatter(e *logs.Entity, msg string) string {
	return "[Format] " + msg
}
