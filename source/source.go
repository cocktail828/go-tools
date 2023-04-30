package source

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
)

type option func(*loadOption)
type sourceF func() string

type loadOption struct {
	defaultVal  string
	dataSources []sourceF
	onExit      func(string) string
}

// 设置默认值
func DefaultValue(v string) option {
	return func(lo *loadOption) { lo.defaultVal = v }
}

// 从环境变量查找
func EnvSource(key string) option {
	return func(lo *loadOption) {
		lo.dataSources = append(lo.dataSources, func() string {
			return os.Getenv(key)
		})
	}
}

// 从启动命令查找
func CmdlineSource(keywords []string) option {
	return func(lo *loadOption) {
		lo.dataSources = append(lo.dataSources, func() string {
			for idx, val := range os.Args {
				cmdkey := val
				cmdval := ""
				slices := strings.Split(val, "=")
				if len(slices) == 2 {
					cmdkey = slices[0]
					cmdval = slices[1]
				}

				if matched := func() bool {
					for _, word := range keywords {
						if cmdkey == word {
							return true
						}
					}
					return false
				}(); !matched {
					continue
				}

				if cmdval != "" {
					return cmdval
				}

				if len(os.Args) > idx {
					return os.Args[idx+1]
				}
			}
			return ""
		})
	}
}

// 从文件中按行查找
func FileSource(key, fname string) option {
	return func(lo *loadOption) {
		lo.dataSources = append(lo.dataSources, func() string {
			file, err := os.Open(fname)
			if err != nil {
				return ""
			}
			defer file.Close()

			reader := bufio.NewReader(file)
			for {
				line, _, err := reader.ReadLine()
				if err == io.EOF {
					break
				}

				slices := strings.Split(string(line), "=")
				if len(slices) != 2 {
					continue
				}

				if slices[0] == key {
					return slices[1]
				}
			}

			return ""
		})
	}
}

// 运行指定命令, 获取输出
func ExecSource(name string, args ...string) option {
	return func(lo *loadOption) {
		lo.dataSources = append(lo.dataSources, func() string {
			cmd := exec.Command(name, args...)
			bf := &bytes.Buffer{}
			cmd.Stdout = bf
			// cmd.Stderr = os.DevNull
			if err := cmd.Run(); err != nil {
				return ""
			}
			return bf.String()
		})
	}
}

// 返回前调用
func OnReturn(f func(val string) string) option {
	return func(lo *loadOption) { lo.onExit = f }
}

func ArgFromSource(options ...option) string {
	o := loadOption{}
	for _, f := range options {
		f(&o)
	}

	for idx := range o.dataSources {
		o.defaultVal = o.dataSources[idx]()
		if o.defaultVal != "" {
			break
		}
	}

	if o.onExit != nil {
		o.defaultVal = o.onExit(o.defaultVal)
	}

	return o.defaultVal
}
