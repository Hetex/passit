package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/gliderlabs/ssh"
	okrun "github.com/oklog/run"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("init passit")

	err := run()
	log.Err(err).Msg("exit passit")
}

func run() error {
	httpAddr := flag.String("http_addr", "localhost:3000", "addr to serve downloads")
	sshAddr := flag.String("ssh_addr", "localhost:2022", "addr to listen for ssh connections")
	keyPath := flag.String("ssrt_addr", "./keys/passit", "path to host key file")

	tm := NewTransferManager()

	httpH := &HTTPHandler{tm: tm}
	httpS := http.Server{
		Addr:    *httpAddr,
		Handler: httpH.routes(),
	}

	sshH := SSHHandler{tm: tm}
	sshS := ssh.Server{
		Addr:    *sshAddr,
		Handler: sshH.handleTransfer(*httpAddr),
	}
	sshS.SetOption(ssh.HostKeyFile(*keyPath))

	ctx := context.Background()
	var g okrun.Group
	{
		g.Add(okrun.SignalHandler(ctx, os.Interrupt))
	}
	{
		log.Info().Msgf("ssh server listening on %s", *sshAddr)
		g.Add(sshS.ListenAndServe, func(_ error) {
			ctx, cancel := context.WithTimeout(ctx, time.Second*30)
			if err := sshS.Shutdown(ctx); err != nil {
				log.Err(err).Msg("ssh server shutdown failed")
			}
			cancel()
		})
	}
	{
		log.Info().Msgf("http server listening on %s", *httpAddr)
		g.Add(httpS.ListenAndServe, func(_ error) {
			ctx, cancel := context.WithTimeout(ctx, time.Second*30)
			if err := httpS.Shutdown(ctx); err != nil {
				log.Err(err).Msg("http server shutdown failed")
			}
			cancel()
		})
	}

	if err := g.Run(); err != nil {
		return err
	}

	return nil
}
