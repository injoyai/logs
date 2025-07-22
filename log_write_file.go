package logs

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
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
}

func (this *File) addIndex(filename string) string {
	if this.fileIndex == 0 {
		return filename
	}
	name := filepath.Base(filename)
	ext := filepath.Ext(filename)
	return filepath.Join(
		filepath.Dir(filename),
		name[:len(name)-len(ext)]+"-"+strconv.Itoa(this.fileIndex)+filepath.Ext(filename),
	)
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
		return this.file.Close()
	}
	return nil
}

func (this *File) Write(p []byte) (int, error) {

	//生成文件名, todo 这里是否耗时严重?
	originFilename := time.Now().Format(this.Filename)
	filename := this.addIndex(originFilename)

	//判断设置的文件地址是否有效
	if len(filename) == 0 {
		return 0, errors.New("无效文件地址")
	}

	//判断文件夹是否存在,不存在则新建
	os.MkdirAll(filepath.Dir(filename), 0666)

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
			filename = this.addIndex(originFilename)
			//获取文件信息
			_, err := os.Stat(filename)
			if err != nil && !os.IsNotExist(err) {
				return 0, err
			} else if os.IsNotExist(err) {
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
		return 0, err
	}

	//记录文件大小
	this.filesize += int64(n)
	return n, nil
}
