package webservise

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/valyala/fasttemplate"
	"github.com/labstack/gommon/log"
	"encoding/json"
)

type (
	Logger struct {
		prefix     string
		level      int
		lv         log.Lvl
		skip       int
		output     io.Writer
		template   *fasttemplate.Template
		levels     []string
		bufferPool sync.Pool
		mutex      sync.Mutex
	}

	Lvl log.Lvl
)

const (
	DEBUG      = iota + 1
	INFO
	WARN
	ERROR
	OFF
	panicLevel
	fatalLevel
)

var (
	defaultHeader = `[${level}]<${prefix}>${time} ${short_file}:${line} `
)

func (l *Logger) initLevels() {
	l.levels = []string{
		"TRACE",
		"DEBUG",
		"INFO",
		"WARN",
		"ERROR",
		"OFF",
		"PANIC",
		"FATAL",
	}
}

func NewLogger(prefix string) (l *Logger) {
	l = &Logger{
		level:    WARN,
		skip:     2,
		prefix:   prefix,
		template: l.newTemplate(defaultHeader),
		bufferPool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 256))
			},
		},
	}

	l.initLevels()

	l.SetOutput(os.Stdout)
	return
}

func (l *Logger) newTemplate(format string) *fasttemplate.Template {
	return fasttemplate.New(format, "${", "}")
}

func (l *Logger) Prefix() string {
	return l.prefix
}

func (l *Logger) SetPrefix(p string) {
	l.prefix = p
}

func (l *Logger) Level() log.Lvl {
	return l.lv
}

func (l *Logger) SetLevel(v log.Lvl) {
	l.lv = v
	l.level = int(v)
}

func (l *Logger) Output() io.Writer {
	return l.output
}

func (l *Logger) SetOutput(w io.Writer) {
	l.output = w
}

func (l *Logger) SetHeader(h string) {
	l.template = l.newTemplate(h)
}

func (l *Logger) Print(i ...interface{}) {
	l.log(0, "", i...)
	// fmt.Fprintln(l.output, i...)
}

func (l *Logger) Printf(format string, args ...interface{}) {
	l.log(0, format, args...)
}

func (l *Logger) Printj(j log.JSON) {
	l.log(0, "json", j)
}

func (l *Logger) Debug(i ...interface{}) {
	l.log(DEBUG, "", i...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Debugj(j log.JSON) {
	l.log(DEBUG, "json", j)
}

func (l *Logger) Info(i ...interface{}) {
	l.log(INFO, "", i...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Infoj(j log.JSON) {
	l.log(INFO, "json", j)
}

func (l *Logger) Warn(i ...interface{}) {
	l.log(WARN, "", i...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Warnj(j log.JSON) {
	l.log(WARN, "json", j)
}

func (l *Logger) Error(i ...interface{}) {
	l.log(ERROR, "", i...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) Errorj(j log.JSON) {
	l.log(ERROR, "json", j)
}

func (l *Logger) Fatal(i ...interface{}) {
	l.log(fatalLevel, "", i...)
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(fatalLevel, format, args...)
	os.Exit(1)
}

func (l *Logger) Fatalj(j log.JSON) {
	l.log(fatalLevel, "json", j)
	os.Exit(1)
}

func (l *Logger) Panic(i ...interface{}) {
	l.log(panicLevel, "", i...)
	panic(fmt.Sprint(i...))
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	l.log(panicLevel, format, args...)
	panic(fmt.Sprintf(format, args...))
}

func (l *Logger) Panicj(j log.JSON) {
	l.log(panicLevel, "json", j)
	panic(j)
}

func (l *Logger) log(v int, format string, args ...interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	buf := l.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer l.bufferPool.Put(buf)
	_, file, line, _ := runtime.Caller(l.skip)

	if v >= l.level || v == 0 {
		message := ""
		if format == "" {
			message = fmt.Sprint(args...)
		} else if format == "json" {
			b, err := json.Marshal(args[0])
			if err != nil {
				panic(err)
			}
			message = string(b)
		} else {
			message = fmt.Sprintf(format, args...)
		}

		_, err := l.template.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
			switch tag {
			case "time":
				return w.Write([]byte(time.Now().Format("2006/01/02 15:04:05")))
			case "level":
				return w.Write([]byte(l.levels[v]))
			case "prefix":
				return w.Write([]byte(l.prefix))
			case "long_file":
				return w.Write([]byte(file))
			case "short_file":
				return w.Write([]byte(path.Base(file)))
			case "line":
				return w.Write([]byte(strconv.Itoa(line)))
			}
			return 0, nil
		})

		if err == nil {
			s := buf.String()
			i := buf.Len() - 1
			if s[i] == '}' {
				// json header
				buf.Truncate(i)
				buf.WriteByte(',')
				if format == "json" {
					buf.WriteString(message[1:])
				} else {
					buf.WriteString(`"message":`)
					buf.WriteString(strconv.Quote(message))
					buf.WriteString(`}`)
				}
			} else {
				// Text header
				buf.WriteByte(' ')
				buf.WriteString(message)
			}
			buf.WriteByte('\n')
			l.output.Write(buf.Bytes())
		}
	}
}
