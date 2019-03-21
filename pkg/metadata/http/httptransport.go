/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/vpoliboy/appmeta/pkg/metadata"
	"gopkg.in/yaml.v2"
	"net/http"
	"strings"
)

const (
	ContentTypeYaml        = "application/x-yaml"
	ContentTypeJson        = "application/json"
	NoContentType          = ""
	ctxKeyMetadataEncoding = "mime"
	jsonEncoding           = "json"
)

var (
	errUnsupportedMimeType  = newError(http.StatusUnsupportedMediaType).WithMessage("application/x-yaml is the only supported content-type")
	errInvalidPayloadFormat = newError(http.StatusBadRequest).WithMessage("content does not match metadata schema")
	errInvalidUUIDinPath    = newError(http.StatusBadRequest).WithMessage("missing or invalid uuid in the request")
)

func MakeHttpHandler(base string, router *mux.Router, middleware mux.MiddlewareFunc, svc metadata.Service, _ *logrus.Logger) http.Handler {

	options := []kithttp.ServerOption{

		//
		kithttp.ServerBefore(kithttp.RequestFunc(func(ctx context.Context, r *http.Request) context.Context {
			encodingRequested := strings.ToLower(r.Header.Get("Accept"))
			if encodingRequested == "" {
				return ctx
			}
			if strings.Contains(encodingRequested, jsonEncoding) {
				return context.WithValue(ctx, ctxKeyMetadataEncoding, jsonEncoding)
			}
			return ctx

		})),

		// All the errors are handled in this configuration.
		//  This method handlers the status codes and error messages.
		kithttp.ServerErrorEncoder(
			func(ctx context.Context, err error, w http.ResponseWriter) {
				if metadata.IsNotFoundError(err) {
					kithttp.DefaultErrorEncoder(ctx, newError(http.StatusNotFound).WithMessage("resource not found"), w)
					return
				}
				switch verr := err.(type) {
				case validation.InternalError, validation.Errors:
					kithttp.DefaultErrorEncoder(ctx, newError(http.StatusBadRequest).WithMessage(verr.Error()), w)
				default:
					kithttp.DefaultErrorEncoder(ctx, err, w)
				}
			}),
	}

	indexHandler := kithttp.NewServer(
		endpoint.Endpoint(func(ctc context.Context, v interface{}) (interface{}, error) {
			metadata := v.(*metadata.Metadata)
			return svc.Insert(metadata)
		}),
		decodeMetadataFromRequest,
		encodeIndexResponseWrapper(base+"/metadata"),
		options...,
	)

	searchHandler := kithttp.NewServer(
		endpoint.Endpoint(func(ctx context.Context, v interface{}) (interface{}, error) {
			filters := v.(metadata.Query)
			if len(filters) == 0 {
				return svc.GetAll(ctx)
			}
			return svc.Search(ctx, filters)
		}),
		decodeSearchFiltersFromRequest,
		encodeMetadataResponse,
		options...,
	)

	getAllHandler := kithttp.NewServer(
		endpoint.Endpoint(func(ctx context.Context, _ interface{}) (interface{}, error) {
			return svc.GetAll(ctx)
		}),
		kithttp.NopRequestDecoder,
		encodeMetadataResponse,
		options...,
	)

	getHandler := kithttp.NewServer(
		endpoint.Endpoint(func(ctx context.Context, v interface{}) (interface{}, error) {
			id := v.(uuid.UUID)
			return svc.Get(ctx, id)
		}),
		decodeUUIDFromRequestPath,
		encodeMetadataResponse,
		options...,
	)

	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		version := svc.Version()
		healthStatus := "green"
		if err := svc.Health(); err != nil {
			healthStatus = "red"
		}

		w.Header().Set("Content-Type", ContentTypeJson)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			Version string `json:"version"`
			Health  string `json:"health"`
		}{version, healthStatus})
	})

	subRouter := router.PathPrefix(base).Subrouter()

	// The order of the calls are important for the uuid match
	subRouter.Handle("/metadata", middleware(indexHandler)).Methods(http.MethodPost)
	subRouter.Handle("/metadata", middleware(getAllHandler)).Methods(http.MethodGet)
	subRouter.Handle("/metadata/_search", middleware(searchHandler)).Methods(http.MethodGet)
	subRouter.Handle("/metadata/_health", middleware(healthHandler)).Methods(http.MethodGet)
	subRouter.Handle("/metadata/{uuid}", middleware(getHandler)).Methods(http.MethodGet)

	subRouter.NotFoundHandler = http.NotFoundHandler()
	subRouter.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	return router

}

func decodeUUIDFromRequestPath(_ context.Context, r *http.Request) (interface{}, error) {
	var (
		id  uuid.UUID
		err error
	)
	idStr := mux.Vars(r)["uuid"]
	if id, err = uuid.Parse(idStr); err != nil {
		return nil, errInvalidUUIDinPath
	}
	return id, nil
}

func decodeMetadataFromRequest(_ context.Context, r *http.Request) (interface{}, error) {

	var (
		metadata = &metadata.Metadata{}
		err      error
	)

	contentType := strings.ToLower(r.Header.Get("content-type"))
	switch contentType {
	case NoContentType, ContentTypeYaml:
		if err = yaml.NewDecoder(r.Body).Decode(metadata); err != nil {
			return nil, errInvalidPayloadFormat
		}
	case ContentTypeJson:
		if err = json.NewDecoder(r.Body).Decode(metadata); err != nil {
			return nil, errInvalidPayloadFormat
		}
	default:
		return nil, errUnsupportedMimeType
	}
	return metadata, nil
}

func encodeIndexResponseWrapper(resourceBase string) kithttp.EncodeResponseFunc {
	return kithttp.EncodeResponseFunc(func(_ context.Context, w http.ResponseWriter, v interface{}) error {
		id := v.(uuid.UUID)
		w.Header().Set("Location", fmt.Sprintf("%s/%s", resourceBase, id.String()))
		w.WriteHeader(http.StatusCreated)
		return nil
	})
}

func decodeSearchFiltersFromRequest(_ context.Context, r *http.Request) (interface{}, error) {
	query := metadata.Query{}

	queryParams := r.URL.Query()
	for k, v := range queryParams {
		if len(v) > 0 {
			query[metadata.SearchField(k)] = v[0]
		}
	}
	return query, nil
}

func encodeMetadataResponse(ctx context.Context, w http.ResponseWriter, v interface{}) error {

	encodingRequested, _ := ctx.Value(ctxKeyMetadataEncoding).(string)

	switch encodingRequested {
	case jsonEncoding:
		w.Header().Set("Content-Type", ContentTypeJson)
		w.WriteHeader(http.StatusOK)
		return json.NewEncoder(w).Encode(v)
	default:
		w.Header().Set("Content-Type", ContentTypeYaml)
		w.WriteHeader(http.StatusOK)
		return yaml.NewEncoder(w).Encode(v)
	}
}
