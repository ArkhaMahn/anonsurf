package command

import (
	"fmt"
	"io"
	"strconv"
	"sync"
)

var verboseOutput struct {
	sync.RWMutex
	writer io.Writer
}

func SetVerboseOutput(writer io.Writer) {
	verboseOutput.Lock()
	defer verboseOutput.Unlock()
	verboseOutput.writer = writer
}

func writeVerboseInvocation(command string, args []string, stdout, stderr []byte) {
	verboseOutput.RLock()
	defer verboseOutput.RUnlock()
	if verboseOutput.writer == nil {
		return
	}

	fmt.Fprintf(verboseOutput.writer, "[CMD ] %s", command)
	for _, arg := range args {
		fmt.Fprintf(verboseOutput.writer, " %s", strconv.Quote(arg))
	}
	fmt.Fprintln(verboseOutput.writer)
	writeVerboseBlock(verboseOutput.writer, "stdout", stdout)
	writeVerboseBlock(verboseOutput.writer, "stderr", stderr)
}

func writeVerboseBlock(writer io.Writer, name string, data []byte) {
	if len(data) == 0 {
		return
	}
	fmt.Fprintf(writer, "%s:\n", name)
	_, _ = writer.Write(data)
	if data[len(data)-1] != '\n' {
		fmt.Fprintln(writer)
	}
}
