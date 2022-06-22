package api

import (
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	ecr "github.com/awslabs/amazon-ecr-credential-helper/ecr-login"
	"github.com/chrismellard/docker-credential-acr-env/pkg/credhelper"
	"github.com/go-chi/httplog"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/github"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sigstore/cosign/pkg/cosign"
	ociremote "github.com/sigstore/cosign/pkg/oci/remote"
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
			RegistryClientOpts: []ociremote.Option{
				ociremote.WithRemoteOptions(remote.WithContext(api.ctx)),
			},
			RootCerts:   x509.NewCertPool(),
			SigVerifier: key,
		}

		if api.k8sKeychain {
			kc := authn.NewMultiKeychain(
				authn.DefaultKeychain,
				google.Keychain,
				authn.NewKeychainFromHelper(ecr.NewECRHelper(ecr.WithLogger(ioutil.Discard))),
				authn.NewKeychainFromHelper(credhelper.NewACRCredentialsHelper()),
				github.Keychain,
			)
			opts.RegistryClientOpts = append(
				opts.RegistryClientOpts,
				ociremote.WithRemoteOptions(remote.WithAuthFromKeychain(kc)),
			)
		}

		if _, _, err = cosign.VerifyImageSignatures(api.ctx, ref, opts); err != nil {
			http.Error(w, "no valid signature is found", http.StatusBadRequest)
			return
		}

		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte("OK"))
	}
}
