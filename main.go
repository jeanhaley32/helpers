package main

import "time"

func main() {
	l := StartLogger()
	l.Debug("This is a debug message")
	l.Error("This is an error message")
	l.Warning("This is a warning message")
	l.Info("this is an info message")
	time.Sleep(1 * time.Second)
	l.Critical("This is a critical message")
	time.Sleep(1 * time.Second)
}
