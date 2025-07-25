package logs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

//==============================WriteFile==============================

// NewFile 写入文件
func NewFile(filename string, maxSize ...int64) io.Writer {
	f := &File{
		Filename: filename,
	}
	if len(maxSize) > 0 && maxSize[0] > 0 {
		f.MaxSize = maxSize[0]
	}
	return f
}

// File 写入文件,自动打开关闭文件
type File struct {
	Filename string //日志名称
	MaxSize  int64  //最大文件大小

	filename  string         //现在打开的文件名称
	file      io.WriteCloser //文件流
	filesize  int64          //当前文件大小
	fileIndex int            //文件分片序号

	lastOriginFilename string     //缓存上一次的文件名称
	lastTime           time.Time  //缓存上一次的时间
	mu                 sync.Mutex //并发锁
}

func (this *File) addIndex(filename string, index int) string {
	if index == 0 {
		return filename
	}
	name := filepath.Base(filename)
	ext := filepath.Ext(filename)
	return filepath.Join(filepath.Dir(filename), name[:len(name)-len(ext)]+"-"+strconv.Itoa(index)+ext)
}

func (this *File) getOriginFilename() string {
	now := time.Now()
	if now.Minute() == this.lastTime.Minute() &&
		now.Hour() == this.lastTime.Hour() &&
		now.Day() == this.lastTime.Day() {
		return this.lastOriginFilename
	}
	this.lastTime = now
	this.lastOriginFilename = now.Format(this.Filename)
	return this.lastOriginFilename
}

func (this *File) open(filename string) error {
	//新建文件(如果不存在),添加至文件最后
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeAppend|os.ModePerm)
	if err != nil {
		return err
	}
	info, err := file.Stat()
	if err != nil {
		return err
	}

	//赋值文件名和文件流,方便下次操作
	this.filename = filename
	this.file = file
	this.filesize = info.Size()
	return nil
}

func (this *File) close() error {
	if this.file != nil {
		err := this.file.Close()
		this.file = nil
		return err
	}
	return nil
}

func (this *File) Write(p []byte) (int, error) {
	this.mu.Lock()
	defer this.mu.Unlock()

	//生成文件名
	originFilename := this.getOriginFilename()
	filename := this.addIndex(originFilename, this.fileIndex)

	//判断设置的文件地址是否有效
	if len(filename) == 0 {
		return 0, errors.New("无效文件地址")
	}

	//判断文件夹是否存在,不存在则新建
	os.MkdirAll(filepath.Dir(filename), 0755)

	if this.file == nil || filename != this.filename ||
		this.MaxSize > 0 && this.filesize+int64(len(p)) > this.MaxSize {

		//文件名称发生变化,例如时间变化,或单个文件大小超出,新建文件
		//则关闭之前的文件,并重新生成文件
		this.close()

		//重置文件序号
		if this.file == nil || filename != this.filename {
			this.fileIndex = 0
		}

		//生成带后缀的文件名称
		for ; ; this.fileIndex++ {
			filename = this.addIndex(originFilename, this.fileIndex)
			//获取文件信息
			info, err := os.Stat(filename)
			if err != nil && !os.IsNotExist(err) {
				return 0, err
			} else if os.IsNotExist(err) {
				break
			}

			//判断下一个序号是否存在,不存在则说明当前这个是最后(最新)一个文件,打开判断大小
			_, err = os.Stat(this.addIndex(originFilename, this.fileIndex+1))
			if err != nil && !os.IsNotExist(err) {
				return 0, err
			} else if err != nil && (info.Size() < this.MaxSize || this.MaxSize <= 0) {
				break
			}
		}

		if err := this.open(filename); err != nil {
			return 0, err
		}
	}

	//写入数据
	n, err := this.file.Write(p)
	if err != nil {
		return 0, fmt.Errorf("写入文件 %s 失败: %w", this.filename, err)
	}

	//记录文件大小
	this.filesize += int64(n)
	return n, nil
}
