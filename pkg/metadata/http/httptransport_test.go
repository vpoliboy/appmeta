/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package http

import (
	"bytes"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/vpoliboy/appmeta/pkg/metadata"
	"gopkg.in/yaml.v2"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	nopMiddleware = mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return next
	})
)

func TestIndexAndSearch(t *testing.T) {

	logger := logrus.New()
	service := metadata.NewService(logger)
	handler := MakeHttpHandler("", mux.NewRouter(), nopMiddleware, service, logger)

	server := httptest.NewServer(handler)
	defer server.Close()

	m1 := []byte(`title: Valid App 2
version: 1.0.1
maintainers:
- name: Vijay Poliboyina
  email: apptwo@hotmail.com
company: Upbound Inc.
website: https://upbound.io
source: https://github.com/upbound/repo
license: Apache-2.0
description: |
 ### Why app 2 is the best
 Because it simply is...`)

	req, err := http.NewRequest(http.MethodPost, server.URL + "/metadata", bytes.NewReader(m1))
	assert.Nil(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	res.Body.Close()

	res, err = http.Get(server.URL + "/metadata/_search?name=vijayx")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	var hits []metadata.MetadataWithID

	err = yaml.NewDecoder(res.Body).Decode(&hits)
	assert.Nil(t, err)
	assert.Len(t, hits, 0)
	res.Body.Close()


	res, err = http.Get(server.URL + "/metadata/_search?name=vijay")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	err = yaml.NewDecoder(res.Body).Decode(&hits)
	assert.Nil(t, err)
	assert.Len(t, hits, 1)
	res.Body.Close()
}
