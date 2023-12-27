package logs

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//==============================WriteFile==============================

// File 写入文件,自动打开关闭文件
type File struct {
	Name   string //实例名称
	Dir    string //路径
	Layout string //文件格式

	filename string         //现在打开的文件名称
	file     io.WriteCloser //文件流
}

func (this *File) Write(p []byte) (int, error) {

	//生成文件名
	layout := strings.ReplaceAll(this.Layout, "{name}", this.Name)
	filename := this.Dir + time.Now().Format(layout)

	//判断设置的文件地址是否有效
	if len(filename) == 0 {
		return 0, errors.New("无效文件地址")
	}

	//是否是新的文件,或文件流是否存在
	if filename != this.filename || this.file == nil {
		//获取文件不匹配
		//则关闭之前的文件,并重新生成文件
		if this.file != nil {
			if err := this.file.Close(); err != nil {
				return 0, err
			}
		}

		//判断文件是否存在,不存在则新建
		os.MkdirAll(filepath.Dir(filename), 0666)

		//新建文件(如果不存在),添加至文件最后
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeAppend|os.ModePerm)
		if err != nil {
			log.Println(err)
			return 0, err
		}

		//赋值文件名和文件流,方便下次操作
		this.filename = filename
		this.file = file

	}

	//写入数据
	return this.file.Write(p)
}

// NewFile 写入文件
func NewFile(name, dir, layout string) io.Writer {
	f := &File{Name: name, Dir: dir, Layout: layout}
	return f
}

var DefaultRemoveFile = NewRemoveFile(0)

type RemoveFile struct {
	SaveTime   time.Duration       //保存时间
	RemoveFunc func(r *RemoveFile) //删除文件函数
}

func (this *RemoveFile) SetSaveTime(saveTime time.Duration) *RemoveFile {
	this.SaveTime = saveTime
	return this
}

func (this *RemoveFile) SetFunc(f func(r *RemoveFile)) *RemoveFile {
	this.RemoveFunc = f
	return this
}

func NewRemoveFile(saveTime time.Duration) *RemoveFile {
	r := &RemoveFile{
		SaveTime: saveTime,
		RemoveFunc: func(r *RemoveFile) {
			date := time.Now().Add(-r.SaveTime).Format("2006-01-02")
			files, _ := ioutil.ReadDir(DefaultDir)
			for _, v := range files {
				if len(v.Name()) == 10 && v.Name() <= date {
					_ = os.RemoveAll(DefaultDir + "/" + v.Name())
				}
			}
		},
	}
	go func(r *RemoveFile) {
		for {
			time.Sleep(time.Second * 5)
			if r.SaveTime > 0 && r.RemoveFunc != nil {
				r.RemoveFunc(r)
			}
			time.Sleep(time.Hour)
		}
	}(r)
	return r
}
