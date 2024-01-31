# Logger
A Custom project agnostic logger, intended for use with most of my projects. 

## Why?
This is a two-day effort to combine some of the common logging functionality I've implemented in previous projects. 
I "Bit the bullet" and combined them into a single package. 

## Objective
This version is intended to create pathways to log to console messages of different severities
`INFO,` `DEBUG,` `ERROR,` `WARNING,` `INTERRUPT SIGNAL`, and `CRITICAL.`

Message logged with severities print out as time-stamped, and color-coded messages to console. 
Critical and signal interrupt messages call a "generic exit sequence" that prints the amount of time
the service has run for and exits the application. 

## How to use
The below example starts a logger with a generic log.Default() base logger interface.

```golang
 // Create Logger, passing in log.Default() as the base logger satisfying the log.Logger interface.
 l := logger.StartLogger(log.Default())

 // Example function that returns an error
 _, err := foo()
 if err != nil {
 // Choose the severity of this error, if it's just a warning, this will print a warning message.
 // critical will close the application.
  l.Warning(err)
 }
```

## Example
![](logger.gif)


## Caveats
I want to make this more customizable in the future and work to create paths for logging to multiple sources. 
There are still more concepts I need to learn when setting up logging this way, and I need to dive deeper to implement them. 

There's also a decent chance I'm re-inventing the wheel with a lot of this, but I find a lot of personal value in recreating things that
possibly already exist, at least for my own personal projects.

> The design for this may also be flawed in some conceptual ways.
> 1. Creating multiple separate channels for different error severities was a fun idea. Still, in practice, this non-linear approach disadvantages situations where Time series sequencing is key, which is usually the case with logs.
> 2.  A solution may be to pass the generated message to a queue, where every message is sorted by time and printed off one at a time. The generic exit sequence can close all receiving channels, and wait for the message queue to be printed.
>  3. If I can program all go routines to listen on an exit channel and program them to die when that channel is closed, we can gently shut down the application and clear the queue.
> 4. Once the queue is cleared, we can shut down the logger and fully exit the application.

> This above idea makes me think of a few other implementations. If we take in logs like this and feed them into a sorted view, we can then
> Cut out a slice of that sort view and show that to the user as a "page."
> If we save our logs to a file, and use rolling appenders, we can provide the user with a final paginated display. This may be a better
> idea for a separate project for viewing logs within the terminal.
>
> There's still plenty of room for improvement with this project. I will leave it as is for now, and add some utilitarian functionality as needed.
