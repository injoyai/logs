# logs
日志

### 如何使用

```go

package main

import (
	"github.com/injoyai/logs"
	"time"
)

func main() {

	//打印执行时间
	defer logs.Spend("耗时: ")()

	//默认打印日志
	logs.Debug("Debug")
	logs.Info("Info")
	logs.Error("Error")

	//全局设置打印日志等级,打印全部
	logs.SetLevel(logs.LevelAll)

	//全局设置打印颜色
	logs.SetShowColor()

	//自定义日志,并写入文件 结果: [测试] 2024-02-01 08:02:05 Test
	t := logs.New("测试").WriteToFile(logs.DefaultDir, logs.DefaultLayout)
	t.Println("Test")

	//全局设置自定义打印模板   
	logs.SetFormatter(logs.TimeFormatter)

	//全局添加日志输出方式,输出到TCP客户端
	logs.WriteToTCPClient("127.0.0.1:10086")

	//等效于,输出到TCP客户端
	w, err := logs.NewTCPClient("127.0.0.1:10086")
	//遇到错误打印并panic
	logs.PanicErr(err)
	//全局设置日志输出到io.Writer
	logs.AddWriter(w)

	//设置日志保存时间,使用默认日志文件位置有效,保存到其他位置需要自行处理
	logs.SetSaveTime(time.Hour*24)

}

```


