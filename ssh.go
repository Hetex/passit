package main

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/rs/zerolog/log"
)

type SSHHandler struct {
	tm *TransferManager
}

func (h *SSHHandler) handleTransfer(httpAddr string) func(s ssh.Session) {
	return func(s ssh.Session) {
		fileName := "transfer.changeme"

		for _, command := range s.Command() {
			if s, ok := strings.CutPrefix(command, "file="); ok {
				fileName = s
			}
		}

		tID, transfer := h.tm.NewTransfer(fileName)

		log.Info().Msgf("%s: transfer created", tID)
		fmt.Fprintf(s, "download link: http://%s/dl/%s\n", httpAddr, tID)

		transferedBytes := 0

		reader := bufio.NewReader(s)
		inputBuf := make([]byte, 1024)
		for {
			n, err := reader.Read(inputBuf)
			if err != nil {
				break
			}

			outputBuf := make([]byte, n)
			copy(outputBuf, inputBuf[:n])

			transfer.channel <- outputBuf

			transferedBytes += len(outputBuf)
			fmt.Fprintf(s, "sent: %d MB \033[0K\r", transferedBytes/1000000)
		}

		fmt.Fprint(s, "transfer recieved")
		s.Exit(0)

		close(transfer.channel)
	}
}
