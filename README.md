# MLOGGER  
About:      module for multi file level logging  
Author:     F.Pessolano  
Licence:    MIT  

NO LONGER MANTAINED

**Description**  
Module for implementing centralised level loggers with multiple files and aggregation of log lines.  
Each log line follows this format:  

    DATE_LINE_CREATION    ID  LEVEL   MESSAGE DATA    DATE_LAST_CHANGE

The mlogger module supports as level LOG, INFO, ERROR, WARNING, RECOVERED amd PANIC. Apart from the traditional message, data can be included in the log line
and it can be accumulated over time (in this case the last update date is also added as of the first accumulation). 
Log files are stored in the folder ./log/   

**Initialisation**  
A logger is initialised with:

    logId, er := mlogger.DeclareLog(name, date) 

Where _name_ is the logfile name. A log file is created everyday and the date appended to the name if _date_ is _true_.  
An id is given in _logId_ (int) and an _er_ error returned if any.  
The logfile is formatted with the methods:  

    mlogger.SetTextLimit(logId, lm, li, ll)
    
Where _lm_, _li_ and _ll_ are the number of maximum characters to be used for the message text, id and level. If 0 is given, no restriction will be used.
The logs can also be set to echo all written liones to the console by togglingthe verbose flag:

    mlogger.Verbose({true|false})


**Usage**
A log line can be stored by using a method associated to a given level (Log, Info, Error, Warning, Recovered amd Panic). For example:  

    mlogger.Log(logId, LoggerData{Id: string, Message: string, Data: []int, Agregate: bool})  
    
_LoggerData_ is a struct containing the log line data.
When _Aggregate_ is true, the data in _Data_ will be summed and the old first written log line will be updated with the new value and the latest modification date.  
The _mlogger.Panic_ level method accept an additional parameter:

    mlogger.Panic(logId, LoggerData{Id: id, Message: message, Data: data, Agregate: aggregate}, quit)  

_Quit_ is a bool. When set, it will force the execution to brutally halt.





