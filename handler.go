package ueloghandler

type LogHandler interface {
	HandleLog(log Log) error
}

func NewLogHandler(function func(log Log) error) LogHandler {
	return &funcLogHanlder{function: function}
}

type funcLogHanlder struct {
	function func(log Log) error
}

func (l *funcLogHanlder) HandleLog(log Log) error {
	return l.function(log)
}
