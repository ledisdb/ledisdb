package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/siddontang/ledisdb/ledis"
	"os"
	"time"
)

var TimeFormat = "2006-01-02 15:04:05"

var startDateTime = flag.String("start-datetime", "",
	"Start reading the binary log at the first event having a timestamp equal to or later than the datetime argument.")
var stopDateTime = flag.String("stop-datetime", "",
	"Stop reading the binary log at the first event having a timestamp equal to or earlier than the datetime argument.")

var startTime uint32 = 0
var stopTime uint32 = 0xFFFFFFFF

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s [options] log_file\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	logFile := flag.Arg(0)
	f, err := os.Open(logFile)
	if err != nil {
		println(err.Error())
		return
	}
	defer f.Close()

	var t time.Time

	if len(*startDateTime) > 0 {
		if t, err = time.Parse(TimeFormat, *startDateTime); err != nil {
			println("parse start-datetime error: ", err.Error())
			return
		}

		startTime = uint32(t.Unix())
	}

	if len(*stopDateTime) > 0 {
		if t, err = time.Parse(TimeFormat, *stopDateTime); err != nil {
			println("parse stop-datetime error: ", err.Error())
			return
		}

		stopTime = uint32(t.Unix())
	}

	rb := bufio.NewReaderSize(f, 4096)
	err = ledis.ReadEventFromReader(rb, printEvent)
	if err != nil {
		println("read event error: ", err.Error())
		return
	}
}

func printEvent(createTime uint32, event []byte) error {
	if createTime < startTime || createTime > stopTime {
		return nil
	}

	t := time.Unix(int64(createTime), 0)

	fmt.Printf("%s ", t.Format(TimeFormat))

	s, err := ledis.FormatBinLogEvent(event)
	if err != nil {
		fmt.Printf("%s", err.Error())
	} else {
		fmt.Printf(s)
	}

	fmt.Printf("\n")

	return nil
}
