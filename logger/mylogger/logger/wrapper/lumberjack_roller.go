package wrapper

import "gopkg.in/natefinch/lumberjack.v2"

type LmberjackRoller struct {
	roller *lumberjack.Logger
}

func NewLmberjackRoller(fileName string, maxSize int, maxBackup, maxAge int, compress bool) *LmberjackRoller {
	l := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    maxSize, // megabytes
		MaxBackups: maxBackup,
		MaxAge:     maxAge,   //days
		Compress:   compress, // disabled by default
	}
	return &LmberjackRoller{
		roller: l,
	}
}

func (l *LmberjackRoller) Write(p []byte) (n int, err error) {
	return l.roller.Write(p)
}

func (l *LmberjackRoller) Close() error {
	return l.roller.Close()
}

func (l *LmberjackRoller) Rotate() error {
	return l.roller.Rotate()
}

func (l *LmberjackRoller) Sync() error {
	return nil
}

func (l *LmberjackRoller) GetPath() string {
	return l.roller.Filename
}
