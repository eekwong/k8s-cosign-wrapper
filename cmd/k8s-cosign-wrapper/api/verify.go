package api

import (
	"crypto/x509"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/httplog"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/sigstore/cosign/pkg/cosign"
	"github.com/sigstore/cosign/pkg/signature"
)

func (api *api) verify() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := struct {
			Image string `json:"image"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		request.Image = strings.TrimSpace(request.Image)
		if request.Image == "" {
			http.Error(w, "image cannot be empty", http.StatusBadRequest)
			return
		}

		oplog := httplog.LogEntry(r.Context())
		oplog.Info().Msgf("image: %s", request.Image)

		api.mux.Lock()
		defer api.mux.Unlock()

		key, err := signature.LoadPublicKey(api.ctx, api.key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ref, err := name.ParseReference(request.Image)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		opts := &cosign.CheckOpts{
			RootCerts:   x509.NewCertPool(),
			SigVerifier: key,
		}

		if _, _, err = cosign.VerifyImageSignatures(api.ctx, ref, opts); err != nil {
			http.Error(w, "no valid signature is found", http.StatusBadRequest)
			return
		}

		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte("OK"))
	}
}
