package main

// 程序名：logdel
// 用途：删除带有日期文件，特别是java应用产生的日志，例如tomcat应用日志。

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

var optConfDir string
var optDryRun bool

type Configurations struct {
	Global Global  `yaml:"global"`
	Items  []Items `yaml:"items"`
}

type Global struct {
	RemainedNum int      `yaml:"remained_num"`
	DateFormats []string `yaml:"date_formats"`
	Suffixes    []string `yaml:"suffixes"`
}

type Items struct {
	Name        string   `yaml:"name"`
	Paths       []string `yaml:"paths"`
	RemainedNum int      `yaml:"remained_num"`
	DateFormats []string `yaml:"date_formats"`
	Suffixes    []string `yaml:"suffixes"`
}

func getConfFiles(confDir string) []string {
	confFiles := make([]string, 0)
	files, err := ioutil.ReadDir(confDir)
	if err != nil {
		log.Fatalf("打开配置文件目录\"%s\"失败。", confDir, err)
	}

	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()
			// 配置文件后缀名为`.yaml`或`.yml`。
			if strings.HasSuffix(strings.ToLower(fileName), ".yaml") || strings.HasSuffix(strings.ToLower(fileName), ".yml") {
				confFiles = append(confFiles, optConfDir+"/"+file.Name())
			}
		}
	}
	return confFiles
}

func decodeConfFile(confFile string) (*Configurations, error) {
	conf := new(Configurations)

	f, err := os.Open(confFile)
	if err != nil {
		log.Errorf("打开配置文件目录\"%s\"失败。\n", confFile)
		return conf, err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(conf); err != nil {
		return conf, err
	}

	return conf, nil
}

func parseConf(confFile string, confData *Configurations) []Items {
	instances := make([]Items, 0)

	// 当全局参数没有配置时，使用默认值。
	defaultRemainedNum := 7
	defaultDateFormats := []string{"YYYY-MM-DD", "YYYYMMDD", "YYYY_MM_DD"}
	defaultSuffixes := []string{".log", ".txt"}

	// 当全局参数配置时，覆盖默认值。
	if confData.Global.RemainedNum != 0 {
		defaultRemainedNum = confData.Global.RemainedNum
	}

	if len(confData.Global.DateFormats) != 0 {
		defaultDateFormats = confData.Global.DateFormats
	}

	if len(confData.Global.Suffixes) != 0 {
		defaultSuffixes = confData.Global.Suffixes
	}

	// 当配置文件中，没有填写items配置时，其中name和path均为默认值，则过滤此配置。
	for _, item := range confData.Items {
		if len(item.Name) == 0 {
			log.Warnf("配置文件\"%s\"配置项items子项目\"%#v\"中\"name\"参数为空，忽略此配置项。", confFile, item)
			continue
		}

		if len(item.Paths) == 0 {
			log.Warnf("配置文件\"%s\"配置项items子项目\"%#v\"中\"paths\"参数为空，忽略此配置项。", confFile, item)
			continue
		}

		if item.RemainedNum == 0 {
			item.RemainedNum = defaultRemainedNum
		}

		if len(item.DateFormats) == 0 {
			item.DateFormats = defaultDateFormats
		}

		if len(item.Suffixes) == 0 {
			item.Suffixes = defaultSuffixes
		}

		instances = append(instances, item)
	}
	return instances
}

func delLogFiles(logItems []Items, optDryRun bool) {
	for _, item := range logItems {
		// 先是根据文件后缀名进行分类。
		for _, logSuffix := range item.Suffixes {
			// 创建临时数组，用于存放匹配到的文件。
			logFiles := make([]string, 0)

			for _, dateFormat := range item.DateFormats {
				// 根据配置项目中的date_formats和suffixes进行组合拼一个正则。
				regexpString := ""
				tmpRegexpString := ""

				tmpRegexpString += strings.ToLower(dateFormat) + strings.ToLower(logSuffix) + "$|"
				tmpRegexpString += strings.ToLower(logSuffix) + strings.ToLower(dateFormat) + "$|"

				regexpString = tmpRegexpString[:len(tmpRegexpString)-1]
				regexpString = strings.ReplaceAll(regexpString, "yyyy", "\\d{4}")
				regexpString = strings.ReplaceAll(regexpString, "mm", "\\d{2}")
				regexpString = strings.ReplaceAll(regexpString, "dd", "\\d{2}")

				for _, path := range item.Paths {
					files, err := ioutil.ReadDir(path)
					if err != nil {
						log.Errorf("读取\"%s\"目录出错。错误为：%#v\n", path, err)
						continue
					} else {
						log.Infof("读取\"%s\"目录成功。\n", path)
					}

					log.Infof("当前匹配模式：%#v\n", regexpString)
					for _, file := range files {
						if !file.IsDir() {
							fileName := file.Name()
							if strings.HasSuffix(strings.ToLower(fileName), logSuffix) {
								validString := regexp.MustCompile(regexpString)
								if validString.MatchString(fileName) {
									// 匹配成功的文件则存入在临时的数组中。
									log.Infof("匹配成功：%s\n", fileName)
									logFiles = append(logFiles, path+"/"+fileName)
								} else {
									log.Warnf("匹配失败：%s\n", fileName)
								}
							}
						}
					}

					// 只有匹配到的文件大于需要保留的文件数时，才执行。
					if len(logFiles) > item.RemainedNum {
						logFiles = logFiles[:len(logFiles)-item.RemainedNum+1]

						if optDryRun {
							for _, logFile := range logFiles {
								log.Infof("当前运行在dry-run模式，仅显示被删除日志文件名：%s", logFile)
							}
						} else {
							for _, logFile := range logFiles {
								log.Infof("当前运行在删除模式，即将删除此日志文件：%s", logFile)
								if err := os.Remove(logFile); err != nil {
									log.Infof("文件删除失败：%s，原因：%#v\n", logFile, err)
								} else {
									log.Infof("文件删除成功：%s\n", logFile)
								}
							}
						}
					} else {
						log.Warnf("根据当前规划匹配不到文件或匹配到的文件数量少于预设值。\n")
					}
					logFiles = make([]string, 0)
				}
			}
		}
	}

}

func main() {
	// 获取参数。
	// 默认配置文件目录在/etc/logdel.d，配置文件以`.yaml`或`.yml`结尾，不能放在这个目录下其他目录下。
	flag.StringVar(&optConfDir, "conf-dir", "/etc/logdel.d", "Set configuration directory.")

	// 运行时，默认跑在dry-run下，并且日志打印在终端。
	flag.BoolVar(&optDryRun, "dry-run", true, "Run in dry-run mode.")

	flag.Parse()

	// 获取配置目录下的配置文件。
	confFiles := getConfFiles(optConfDir)

	if len(confFiles) == 0 {
		log.Errorf("配置文件目录\"%s\"下无配置文件。", optConfDir)
		return
	} else {
		log.Infof("配置文件目录\"%s\"下配置文件列表：%v。\n", optConfDir, confFiles)
	}

	for _, confFile := range confFiles {
		if confData, err := decodeConfFile(confFile); err != nil {
			log.Errorf("解析配置文件\"%s\"失败，跳过此配置文件。\n", confFile)
		} else {
			log.Infof("解析配置文件\"%s\"成功。\n", confFile)
			logItems := parseConf(confFile, confData)

			if len(logItems) == 0 {
				log.Warnf("配置文件\"%s\"中无items配置项。\n", confFile)
			} else {
				for _, logItem := range logItems {
					log.Infof("配置文件\"%s\"中items配置项：%#v。\n", confFile, logItem)
					delLogFiles(logItems, optDryRun)
				}
			}
		}
	}
}
