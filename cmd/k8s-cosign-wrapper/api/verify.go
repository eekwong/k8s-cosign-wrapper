package api

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	ecr "github.com/awslabs/amazon-ecr-credential-helper/ecr-login"
	"github.com/chrismellard/docker-credential-acr-env/pkg/credhelper"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/github"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	alibabaacr "github.com/mozillazg/docker-credential-acr-helper/pkg/credhelper"
	"github.com/rs/zerolog/log"
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

		log.Info().Str("image", request.Image).Msg("received image")

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

		var kc authn.Keychain
		if api.k8sKeychain {
			kc = authn.NewMultiKeychain(
				authn.DefaultKeychain,
				google.Keychain,
				authn.NewKeychainFromHelper(ecr.NewECRHelper(ecr.WithLogger(io.Discard))),
				authn.NewKeychainFromHelper(credhelper.NewACRCredentialsHelper()),
				authn.NewKeychainFromHelper(alibabaacr.NewACRHelper().WithLoggerOut(io.Discard)),
				github.Keychain,
			)
		} else {
			kc = authn.DefaultKeychain
		}
		opts.RegistryClientOpts = append(
			opts.RegistryClientOpts,
			ociremote.WithRemoteOptions(remote.WithAuthFromKeychain(kc)),
		)

		sigs, bundledVerified, err := cosign.VerifyImageSignatures(api.ctx, ref, opts)
		if err != nil {
			log.Error().Err(err).Msg("error from cosign.VerifyImageSignatures")
			http.Error(w, fmt.Sprintf("error: %+v", err), http.StatusBadRequest)
			return
		}

		signatures := make([]string, 0)
		for s := range sigs {
			signatures = append(signatures, fmt.Sprintf("%+v", s))
		}
		log.Info().
			Strs("sigs", signatures).
			Bool("bundledVerified", bundledVerified).
			Msg("return from cosign.VerifyImageSignatures")

		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte("OK"))
	}
}
