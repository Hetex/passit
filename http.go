package main

import (
	"fmt"
	"net/http"

	"github.com/alexedwards/flow"
	"github.com/rs/zerolog/log"
)

type HTTPHandler struct {
	tm *TransferManager
}

func (h *HTTPHandler) routes() http.Handler {
	mux := flow.New()
	mux.HandleFunc("/dl/:id", h.handleDownload, http.MethodGet)

	return mux
}

func (s *HTTPHandler) handleDownload(w http.ResponseWriter, r *http.Request) {
	tID := flow.Param(r.Context(), "id")

	transfer := s.tm.GetTransfer(tID)
	if transfer == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("invalid download link"))
		log.Warn().Msgf("%s: attempted download %s", tID, r.RemoteAddr)
		return
	}

	log.Info().Msgf("%s: download started %s", tID, r.RemoteAddr)

	w.Header().Add("Content-type", "application/octet-stream")
	w.Header().Add("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, transfer.fileName))

	for data := range transfer.channel {
		w.Write(data)
	}

	log.Info().Msgf("%s: transfer complete", tID)
}
