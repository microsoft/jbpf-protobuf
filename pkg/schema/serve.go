// Copyright (c) Microsoft Corporation. All rights reserved.

package schema

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/google/uuid"
)

func readBodyAs[T any](req *http.Request) (out T, err error) {
	buf := new(strings.Builder)
	_, err = io.Copy(buf, req.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(buf.String()), &out)
	return
}

func (s *Server) serveHTTP(ctx context.Context) error {
	http.HandleFunc("/schema", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			body, err := readBodyAs[UpsertSchemaRequest](r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if err := s.UpsertProtoPackage(r.Context(), &body); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				w.WriteHeader(http.StatusOK)
			}

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			body, err := readBodyAs[AddSchemaAssociationRequest](r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if err := s.AddStreamToSchemaAssociation(r.Context(), &body); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				w.WriteHeader(http.StatusOK)
			}

		case http.MethodDelete:
			streamUUIDStr := r.URL.Query().Get("streamUUID")
			bs, err := base64.StdEncoding.DecodeString(streamUUIDStr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			streamUUID, err := uuid.FromBytes(bs)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if err := s.DeleteStreamToSchemaAssociation(r.Context(), streamUUID); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				w.WriteHeader(http.StatusAccepted)
			}

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	if s.opts.jbpf.Enable {
		http.HandleFunc("/control", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				body, err := readBodyAs[SendControlRequest](r)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				if err := s.SendControl(r.Context(), &body); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					w.WriteHeader(http.StatusOK)
				}

			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.opts.control.ip, s.opts.control.port),
		Handler: nil,
	}

	go func() {
		stopper := make(chan os.Signal, 1)
		signal.Notify(stopper, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

		select {
		case <-stopper:
		case <-ctx.Done():
		}
		if err := srv.Close(); err != nil {
			s.logger.WithError(err).Error("failed stopping the server")
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
