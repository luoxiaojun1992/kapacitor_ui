package controllers

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
}

type Tick struct {
	Description string   `form:"description"`
	AlertName   string   `form:"alert_name"`
	Measurement string   `form:"measurement"`
	GroupBy     string   `form:"group_by"`
	Where       string   `form:"where"`
	Period      string   `form:"period"`
	Every       string   `form:"every"`
	Sum         string   `form:"sum"`
	Max         string   `form:"max"`
	Min         string   `form:"min"`
	Mean        string   `form:"mean"`
	Crit        string   `form:"crit"`
	Email       []string `form:"email"`
	Phone       string   `form:"phone"`
	Tick        string   `form:"tick"`
}

const TICK_FROM_TPL = `
	|from()
		.measurement('%s')`

const TICK_WHERE_TPL = `
	|where(lambda: %s)`

const TICK_WINDOW_TPL = `
	|window()
		.period(%s)
		.every(%s)`

const TICK_SUM_TPL = `
	|sum('%s')`

const TICK_MIN_TPL = `
	|min('%s')`

const TICK_MAX_TPL = `
	|max('%s')`

const TICK_MEAN_TPL = `
	|mean('%s')`

const TICK_ALERT_TPL = `
	|alert()
		.id('%s')
		.message('报警规则:{{ .ID }},报警级别:{{ .Level }}')
		.details('''
<h1><span style="color: {{ if eq .Level "OK" }}green{{ else }}red{{ end }};">●</span>{{ .ID }}</h1>
<b>{{ .Message }}</b>
监控值: {{ .Fields }}
''')
		.crit(lambda: %s)
		.log('/tmp/alert/%s.log')`

const TICK_EMAIL_TPL = `
		.email('%s')`

const TICK_GROUP_BY_TPL = `
		.groupBy('%s')`

const TICK_SMS_TPL = `
		.exec('./alert/send_sms.sh', '%s')`

const TICK_INFLUX_OUT_TPL = `
	|influxDBOut()
		.create()
		.database('kapacitor_ui')
		.retentionPolicy('autogen')
		.measurement('alerts')
		.tag('alertName', '%s')`

const NEW_LINE = `
`

func (c *MainController) Get() {
	c.Data["Website"] = "kapacitor.aocs.com.cn"
	c.Data["Email"] = "luoxiaojun1992@sina.cn"
	c.TplName = "index.tpl"
}

/**
 * 生成kapacitor tick
 */
func (c *MainController) GenerateTick() {
	tick := &Tick{}
	if err := c.ParseForm(tick); err != nil {
		panic(err)
	}

	writeTick(tick)

	c.Redirect("/", 302)
}

func writeTick(tick *Tick) {
	tick_script := ""

	if tick.Tick != "" {
		tick_script = tick.Tick
	} else {
		tick_script = "//" + tick.Description

		tick_script += `
stream`

		tick_script += fmt.Sprintf(TICK_FROM_TPL, tick.Measurement)

		if tick.GroupBy != "" {
			tick_script += fmt.Sprintf(TICK_GROUP_BY_TPL, tick.GroupBy)
		}

		if tick.Where != "" {
			tick_script += fmt.Sprintf(TICK_WHERE_TPL, tick.Where)
		}

		tick_script += fmt.Sprintf(TICK_WINDOW_TPL, tick.Period, tick.Every)

		if tick.Sum != "" {
			tick_script += fmt.Sprintf(TICK_SUM_TPL, tick.Sum)
		}

		if tick.Min != "" {
			tick_script += fmt.Sprintf(TICK_MIN_TPL, tick.Min)
		}

		if tick.Max != "" {
			tick_script += fmt.Sprintf(TICK_MAX_TPL, tick.Max)
		}

		if tick.Mean != "" {
			tick_script += fmt.Sprintf(TICK_MEAN_TPL, tick.Mean)
		}

		tick_script += fmt.Sprintf(TICK_ALERT_TPL, tick.Description, tick.Crit, tick.AlertName)

		for _, email := range tick.Email {
			if email != "" {
				tick_script += fmt.Sprintf(TICK_EMAIL_TPL, email)
			}
		}

		if tick.Phone != "" {
			tick_script += fmt.Sprintf(TICK_SMS_TPL, tick.Phone)
		}

		tick_script += fmt.Sprintf(TICK_INFLUX_OUT_TPL, tick.AlertName)

		tick_script += NEW_LINE
	}

	// Generate Tick File
	tick_file_name := "./alert/" + tick.AlertName + ".tick"
	removeFile(tick_file_name)
	writeFile(tick_file_name, tick_script)
	go startTask(tick.AlertName, tick_file_name)
}

func startTask(alert_name string, tick_file_name string) {
	updateModifyScript(alert_name, tick_file_name)

	cmd := exec.Command("./alert/modify.sh")
	if _, err := cmd.Output(); err != nil {
		panic(err)
	}
}

func updateModifyScript(alert_name string, tick_file_name string) {
	modify_script_file_name := "./alert/modify.sh"
	start_command := "kapacitor define " + alert_name + " -type stream -dbrp telegraf.default -tick " + tick_file_name + " && kapacitor enable " + alert_name

	f, err1 := os.Open(modify_script_file_name)
	if err1 != nil {
		panic(err1)
	}
	buf := bufio.NewReader(f)
	for {
		line, err2 := buf.ReadString('\n')
		if err2 != nil {
			if err2 == io.EOF {
				break
			} else {
				panic(err2)
			}
		}
		line = strings.TrimSpace(line)

		if line == start_command {
			return
		}
	}

	appendFile(modify_script_file_name, start_command+NEW_LINE)
}

func appendFile(file_name string, content string) {
	f, err1 := os.OpenFile(file_name, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err1 != nil {
		panic(err1)
	}

	defer f.Close()
	_, err2 := io.WriteString(f, content)
	if err2 != nil {
		panic(err2)
	}
}

func writeFile(file_name string, content string) {
	f, err1 := os.Create(file_name)
	if err1 != nil {
		panic(err1)
	}

	defer f.Close()
	_, err2 := io.WriteString(f, content)
	if err2 != nil {
		panic(err2)
	}
}

func removeFile(file_name string) {
	_, err := os.Stat(file_name)
	if err != nil && os.IsNotExist(err) {
		os.Remove(file_name)
	}
}
