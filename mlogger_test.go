package mlogger

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func Test_DLogger(t *testing.T) {
	if logId, er := DeclareLog("test", true); er == nil {
		if e := SetTextLimit(logId, 20, 10, 12); e != nil {
			fmt.Println(e)
			os.Exit(0)
		}
		Log(logId, LoggerData{"test1", "testing message", []int{2}, true})
		Error(logId, LoggerData{"test1", "testing message", []int{2}, true})
		Info(logId, LoggerData{"test1", "testing message", []int{2}, false})
		Warning(logId, LoggerData{"test1", "testing message", []int{2}, true})
		Recovered(logId, LoggerData{"test1", "testing message", []int{2}, true})
		Panic(logId, LoggerData{"test1", "testing message", []int{}, true}, false)
		time.Sleep(5 * time.Second)
	} else {
		t.Error(er)
	}
}
