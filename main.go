package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/aphistic/golf"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/kr/logfmt"
)

var (
	app     = "LogRouter"
	version string
	build   string
)

type logMessage struct {
	pairs map[string]interface{}
}

func (m *logMessage) HandleLogfmt(key, value []byte) error {
	m.pairs[string(key)] = string(value)
	return nil
}

func main() {
	var (
		versionFlg    = flag.Bool("version", false, "Display application version")
		inputFmtFlg   = flag.String("input-format", "", "Specify the input format; e.g. logfmt, json or unknown")
		outputFlg     = flag.String("output", "", "Output to?")
		addressFlg    = flag.String("graylog-address", "", "IP address or hostname to connect to Graylog")
		portFlg       = flag.String("graylog-port", "", "UDP GELF port")
		attributesFlg = flag.String("graylog-attributes", "", "Comma separated list of attributes to pass to Graylog in the form name:value")
		debug         = flag.Bool("debug", false, "Enable debugging?")
		logger        log.Logger
	)

	flag.Parse()

	if *versionFlg {
		fmt.Println(app + " v" + version + " build " + build)
		os.Exit(0)
	}

	if *inputFmtFlg == "" {
		fmt.Println("You must specify the input format.  Currently only 'logfmt', 'json', or 'unknown' is supported.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *outputFlg == "" {
		fmt.Println("You must specify the output destination.  Currently only 'graylog' is supported.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *addressFlg == "" {
		fmt.Println("You must specify the IP address or hostname to connect to Graylog.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *portFlg == "" {
		fmt.Println("You must specify the EDP GELF input port for Graylog.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	logger = log.NewLogfmtLogger(os.Stdout)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller, "app", app)

	if *debug {
		logger = level.NewFilter(logger, level.AllowDebug())
	} else {
		logger = level.NewFilter(logger, level.AllowInfo())
	}

	if len(os.Args) == 1 {
		level.Error(logger).Log("msg", "log destination not set")
		os.Exit(1)
	}

	fi, err := os.Stdin.Stat()
	if err != nil {
		level.Error(logger).Log("msg", err)
		os.Exit(1)
	}

	if fi.Mode()&os.ModeNamedPipe == 0 {
		level.Error(logger).Log("msg", "No pipe found to consume data from")
		os.Exit(1)
	}

	c, err := golf.NewClient()
	if err != nil {
		level.Error(logger).Log("msg", err)
		os.Exit(1)
	}
	defer c.Close()

	err = c.Dial("udp://" + *addressFlg + ":" + *portFlg)
	if err != nil {
		level.Error(logger).Log("msg", err)
		os.Exit(1)
	}

	l, err := c.NewLogger()
	if err != nil {
		level.Error(logger).Log("msg", err)
		os.Exit(1)
	}

	if strings.ToLower(*outputFlg) == "graylog" && *attributesFlg != "" {
		attributes := strings.Split(*attributesFlg, ",")

		for _, v := range attributes {
			v = strings.TrimSpace(v)
			idx := strings.Index(v, ":")
			name := v[:idx]
			val := v[idx+1:]

			name = strings.TrimSpace(name)
			val = strings.TrimSpace(val)

			l.SetAttr(name, val)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		var msg interface{}

		message := &logMessage{
			pairs: make(map[string]interface{}),
		}

		switch strings.ToLower(*inputFmtFlg) {
		case "logfmt":
			err := logfmt.Unmarshal([]byte(scanner.Text()), message)
			if err != nil {
				level.Error(logger).Log("msg", err)
			}

			msg = message.pairs["msg"]
			delete(message.pairs, "msg")
		case "json":
			err := json.Unmarshal([]byte(scanner.Text()), &message.pairs)
			if err != nil {
				level.Error(logger).Log("msg", err)
			}

			msg = message.pairs["msg"]
			delete(message.pairs, "msg")
		case "unknown":
			msg = scanner.Text()
			message.pairs["level"] = "notice"
		}

		switch message.pairs["level"] {
		case "debug":
			err = l.Dbgm(message.pairs, "%v", msg)
		case "info":
			err = l.Infom(message.pairs, "%v", msg)
		case "warn":
			err = l.Warnm(message.pairs, "%v", msg)
		case "error":
			err = l.Errm(message.pairs, "%v", msg)
		case "alert":
			err = l.Alertm(message.pairs, "%v", msg)
		case "crit":
			err = l.Critm(message.pairs, "%v", msg)
		case "emergency":
			err = l.Emergm(message.pairs, "%v", msg)
		case "notice":
			err = l.Noticem(message.pairs, "%v", msg)
		}

		if err != nil {
			level.Error(logger).Log("msg", err)
		}

		runtime.GC()
	}
}
