package flow

import (
	"bufio"
	"cork/gcp"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/gookit/color"
)

type Log struct {
	Trigger  string
	Message  string
	LogUrl   string
	Progress string
}

type WaitInput struct {
	Trigger         string
	Message         string
	LogUrl          string
	ResponseChannel chan bool
}

var (
	inputWaitChannel chan WaitInput
	logChannel       chan Log
)

const (
	SKIP     = "SKIP"
	ALL_DONE = "ALL_DONE"
)

var (
	urlLink           = color.Gray.Render
	successLabel      = color.Green.Render
	cancelledLabel    = color.Yellow.Render
	errorLabel        = color.Red.Render
	runningLabel      = color.Blue.Render
	skipLabel         = color.Yellow.Render
	waitingInputLabel = color.Magenta.Render
	contextText       = color.White.Render
)

type logMessageFunc func(trigger string, message string, url string)

func errorMessage(trigger string, message string, url string) {
	fmt.Printf("%s %s %s %s\n", errorLabel("[   ERROR   ]"), contextText("["+trigger+"]"), message, urlLink(url))
}

func successMessage(trigger string, message string, url string) {
	fmt.Printf("%s %s %s %s\n", successLabel("[  SUCCESS  ]"), contextText("["+trigger+"]"), message, urlLink(url))
}

func progressMessage(trigger string, message string, url string) {
	fmt.Printf("%s %s %s %s\n", runningLabel("[  RUNNING  ]"), contextText("["+trigger+"]"), message, urlLink(url))
}

func cancelledMessage(trigger string, message string, url string) {
	fmt.Printf("%s %s %s %s\n", cancelledLabel("[ CANCELLED ]"), contextText("["+trigger+"]"), message, urlLink(url))
}

func skipAppMessage(message string) {
	fmt.Printf("%s %s\n", skipLabel("[   SKIP    ]"), message)
}

func waitForInput(trigger string, message string, url string) bool {
	var s string

	fmt.Printf("%s %s %s %s (y/N):", waitingInputLabel("[  WAITING  ]"), contextText("["+trigger+"]"), message, urlLink(url))
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		s = scanner.Text()
	}

	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if s == "y" || s == "yes" {
		return true
	}
	return false
}

func flowLog(log Log) {
	logChannel <- log
}

func flowInputWait(waitInput WaitInput) {
	inputWaitChannel <- waitInput
}

func startLogger(waitGroup *sync.WaitGroup) {
	inputWaitChannel = make(chan WaitInput)
	logChannel = make(chan Log)
	go func() {
		defer close(inputWaitChannel)
		defer close(logChannel)
		cloudBuildLoggerFunctions := map[string]logMessageFunc{
			gcp.SUCCESS:   successMessage,
			gcp.FAILURE:   errorMessage,
			gcp.RUNNING:   progressMessage,
			gcp.CANCELLED: cancelledMessage,
		}
		for {
			select {
			case waitInput := <-inputWaitChannel:
				waitInput.ResponseChannel <- waitForInput(waitInput.Trigger, waitInput.Message, waitInput.LogUrl)

			case log := <-logChannel:
				if log.Progress == ALL_DONE {
					waitGroup.Done()
					continue
				}
				if log.Progress == SKIP {
					skipAppMessage(log.Message)
					continue
				}
				if _, ok := cloudBuildLoggerFunctions[log.Progress]; ok {
					(cloudBuildLoggerFunctions[log.Progress])(log.Trigger, log.Message, log.LogUrl)
				}
			}
		}
	}()
}
