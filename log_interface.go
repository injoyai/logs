package logs

type Logger struct{}

func (*Logger) Trace(v ...interface{}) { Trace(v...) }

func (*Logger) Tracef(format string, v ...interface{}) { Tracef(format, v...) }

func (*Logger) Read(v ...interface{}) { Read(v...) }

func (*Logger) Readf(format string, v ...interface{}) { Readf(format, v...) }

func (*Logger) Write(v ...interface{}) { Write(v...) }

func (*Logger) Writef(format string, v ...interface{}) { Writef(format, v...) }

func (*Logger) Info(v ...interface{}) { Info(v...) }

func (*Logger) Infof(format string, v ...interface{}) { Infof(format, v...) }

func (*Logger) Debug(v ...interface{}) { Debug(v...) }

func (*Logger) Debugf(format string, v ...interface{}) { Debugf(format, v...) }

func (*Logger) Warn(v ...interface{}) { Warn(v...) }

func (*Logger) Warnf(format string, v ...interface{}) { Warnf(format, v...) }

func (*Logger) Error(v ...interface{}) { Error(v...) }

func (*Logger) Errorf(format string, v ...interface{}) { Errorf(format, v...) }
