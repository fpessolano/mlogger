package mlogger

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LoggerData is the data sent to the logger from a client
type LoggerData struct {
	Id        string // an unique identifier is necessary for aggregating messages
	Message   string // possible description
	Data      []int  // effective data
	Aggregate bool   // true if the log entry is be cumulative on its data
}

// logMessage is an internal type for the logging message
type logMessage struct {
	id    int        // logger identifier
	level string     // unique message level
	msg   LoggerData // message for the logger
}

//
type logfile struct {
	idLength      int
	levelLength   int
	messageLength int
	filename      string
	file          *os.File
}

// Exported variables
var BufDepth = 50 // buffer depth for the channel to the logger thread

// Internal variables
var declaredLogs map[int]logfile
var lock sync.RWMutex
var loggerChan chan logMessage
var once sync.Once
var consoleLog = false
var index = 0
var verbose = false
var unrolled = false

// DeclareLog is used to declare a new log of name 'fn'. If 'dt' is true the name will become fn_current date.logfile.
// Furthermore, a new logfile will be created everyday. Otherwise it will be fn.logfile and will not be created everyday.
// It returns the log identifier (int) and the eventual error
func DeclareLog(fn string, dt bool) (int, error) {
	lock.Lock()
	defer lock.Unlock()
	SetUpLogger(consoleLog)
	pwd, err := os.Getwd()
	if err != nil {
		return -1, err
	}
	err = os.MkdirAll("log", os.ModePerm)
	if err != nil {
		return -1, err
	}
	if dt {
		ct := time.Now().Local()
		declaredLogs[index] = logfile{filename: filepath.Join(pwd, "log", fn+"_"+ct.Format("2006-01-02")+".logfile")}
		//availableLogs = append(availableLogs, filepath.Join(pwd, "log", fn+"_"+ct.Format("2006-01-02")+".logfile"))
	} else {
		declaredLogs[index] = logfile{filename: filepath.Join(pwd, "log", fn+".logfile")}
		//availableLogs = append(availableLogs, filepath.Join(pwd, "log", fn+".logfile"))
	}
	index += 1
	return index - 1, nil
}

// Close close all open files
func Close() (e error) {
	lock.Lock()
	for _, file := range declaredLogs {
		if er := file.file.Close(); er != nil && e == nil {
			e = errors.New("error inc closing logfiles")
		}
	}
	lock.Unlock()
	return
}

// Enables verbose (affects all logs)
func Verbose(v bool) {
	verbose = v
}

// Enables unroll (affects all logs)
func Unroll(v bool, fn string) {
	if v {
		file, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}

		log.SetOutput(file)
	}
	unrolled = v
}

// SetTextLimit sets formatting limits for message (lm), id (li) and level (ll) in number of characters for logfile tag
func SetTextLimit(tag, lm, li, ll int) error {
	lock.Lock()
	defer lock.Unlock()
	if c, ok := declaredLogs[tag]; ok {
		c.levelLength = ll
		c.idLength = li
		c.messageLength = lm
		declaredLogs[tag] = c
		return nil
	} else {
		return errors.New("cannot set parameters of a not declared logfile")
	}
}

// SetUpLogger set-up the loggerChan. a sync.Once is used to avoid issues with multiple modules
// cf controls the console flag that determines if the application writes logging errors to the consoles or ignores them
func SetUpLogger(cf bool) {
	once.Do(func() {
		//lock.Lock()
		//defer lock.Unlock()
		consoleLog = cf
		declaredLogs = make(map[int]logfile)
		loggerChan = make(chan logMessage, BufDepth)
		go logger(loggerChan)
	})
}

// Log writes a generic comment to the logfile lg
func Log(lg int, dt LoggerData) { loggerChan <- logMessage{lg, "LOG", dt} }

// Error writes an error level line to the logfile lg
func Error(lg int, dt LoggerData) { loggerChan <- logMessage{lg, "ERROR", dt} }

// Info writes an info level line to the logfile lg
func Info(lg int, dt LoggerData) { loggerChan <- logMessage{lg, "INFO", dt} }

// Warning writes a warning level line to the logfile lg
func Warning(lg int, dt LoggerData) { loggerChan <- logMessage{lg, "WARNING", dt} }

// Recovered writes a panic recovery level line to the logfile lg
func Recovered(lg int, dt LoggerData) { loggerChan <- logMessage{lg, "RECOVERED", dt} }

// Panic writes a panic level line to the logfile lg and stop execution (if quit is true)
func Panic(lg int, dt LoggerData, quit bool) {
	loggerChan <- logMessage{lg, "PANIC", dt}
	if quit {
		time.Sleep(5 * time.Second)
		lock.RLock()
		file := strings.Split(declaredLogs[lg].filename, "/")
		lock.RUnlock()
		fmt.Println("Panic error, execution terminated. See log", file[len(file)-1])
		os.Exit(0)
	}
}

// SetError, depending if the error is not nil, sets the error or not
func SetError(lg int, es string, id string, e error, data []int, aggregate bool) bool {
	if id == "" {
		id = "n/a"
	}
	if e != nil {
		Error(lg, LoggerData{id, es, data, aggregate})
		return false
	}
	return true
}

// logger is the core thread handling all the writings to log files
func logger(data chan logMessage) {

	logEntryGenerator := func(file logfile, olddate string, d logMessage, dt ...[]int) (msg string) {
		date := time.Now().Format("Mon Jan:_2 15:04 2006")
		if file.messageLength != 0 {
			if len(d.msg.Message) > file.messageLength {
				d.msg.Message = d.msg.Message[0:file.messageLength]
			} else {
				for len(d.msg.Message) < file.messageLength {
					d.msg.Message += " "
				}
			}
		}
		if file.idLength != 0 {
			if len(d.msg.Id) > file.idLength {
				d.msg.Id = d.msg.Id[0:file.idLength]
			} else {
				for len(d.msg.Id) < file.idLength {
					d.msg.Id += " "
				}
			}
		}
		if file.levelLength != 0 {
			if len(d.level) > file.levelLength {
				d.level = d.level[0:file.levelLength]
			} else {
				for len(d.level) < file.levelLength {
					d.level += " "
				}
			}
		}
		msg = d.msg.Id + "\t\t" + d.level + "\t\t" + d.msg.Message + ""
		if len(dt) == 0 {
			msg = date + "\t\t" + msg
			if len(d.msg.Data) != 0 {
				msg += "\t\t["
				for _, v := range d.msg.Data {
					msg += strconv.Itoa(v) + ","
				}
				msg = strings.Trim(msg, ",")
				msg += "]"
			}
		} else {
			msg += "\t\t["
			if len(dt[0]) != len(d.msg.Data) {
				return ""
			}
			for i, v := range d.msg.Data {
				msg += strconv.Itoa(v+dt[0][i]) + ","
			}
			msg = strings.Trim(msg, ",")
			msg += "]"
			msg = olddate + "\t\t" + msg + "\t\t" + date
		}
		msg = strings.Trim(msg, " ")
		return
	}

	consoleEntryGenerator := func(d logMessage, skipDate bool) (msg string) {
		date := time.Now().Format("Mon Jan:_2 15:04 2006")
		msg = d.level + " -- " + d.msg.Id + ": " + d.msg.Message + ""
		if !skipDate {
			msg = date + " -- " + msg
		}
		msg = strings.Trim(msg, " ")
		return
	}

	defer func() {
		if e := recover(); e != nil {
			if consoleLog {
				fmt.Printf("support.logger: recovering for crash, %v\n ", e)
			}
			go logger(data)
		}
	}()

	for {
		d := <-data
		lock.RLock()
		if d.id < index {
			file := declaredLogs[d.id]
			lock.RUnlock()
			if verbose {
				fmt.Println(consoleEntryGenerator(d, false))
			}
			if unrolled {
				log.Println(consoleEntryGenerator(d, true))
			}
			if input, err := ioutil.ReadFile(file.filename); err != nil {
				if fn, err := os.Create(file.filename); err != nil {
					if consoleLog {
						fmt.Println("support.logger: error creating log: ", err)
					}
				} else {
					if _, err := fn.WriteString(logEntryGenerator(file, "", d) + "\n"); err != nil {
						if consoleLog {
							fmt.Println("support.logger: error creating log: ", err)
						}
					}
					//noinspection GoUnhandledErrorResult
					fn.Close()

				}
			} else {
				// read file and add or replace level
				newC := ""
				adFile := true
				for _, v := range strings.Split(strings.Trim(string(input), " "), "\n") {
					spv := strings.Split(v, "\t\t")
					if len(spv) >= 4 {
						if strings.Trim(spv[1], " ") == d.msg.Id && strings.Trim(spv[2], " ") == d.level && d.msg.Aggregate {
							var nd []int
							skip := false
							data := strings.Split(spv[4][1:len(spv[4])-1], ",")
						mainloop:
							for _, dt := range data {
								if val, e := strconv.Atoi(strings.Trim(dt, " ")); e == nil {
									nd = append(nd, val)
								} else {
									cd := logMessage{d.id, "System Warning", LoggerData{"logger", "error converting accruing data from log " + d.msg.Id,
										[]int{}, false}}
									newC += logEntryGenerator(file, "", cd) + "\n"
									newC += logEntryGenerator(file, "", d) + "\n"
									skip = true
									adFile = false
									break mainloop
								}
							}
							if !skip {
								newC += logEntryGenerator(file, spv[0], d, nd) + "\n"
								adFile = false
							}
						} else {
							if tmp := strings.Trim(v, " "); tmp != "" {
								newC += tmp + "\n"
							}

						}
					}
				}
				if adFile {
					newC += logEntryGenerator(file, "", d) + "\n"
				}
				if err = ioutil.WriteFile(file.filename, []byte(newC), 0644); err != nil {
					log.Println("support.logger: error writing log: ", err)
				}
			}
		} else {
			lock.RUnlock()
		}
	}
}
