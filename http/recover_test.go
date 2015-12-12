package http

// This file is Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style license.
//
// Modifications by Kevin Stenerson for Reflexion Health Inc. Copyright 2015

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reflexionhealth/vanilla/router"
	"github.com/reflexionhealth/vanilla/router/routertest"
)

// TestPanicInHandler assert that panic has been recovered.
func TestPanicInHandler(t *testing.T) {
	buffer := new(bytes.Buffer)
	Logger.Global.SetOutput(buffer)

	r := router.New()
	r.Use(Recover)
	r.GET("/recovery", func(_ *router.Context) {
		panic("Oupps, Houston, we have a problem")
	})

	// RUN
	w := routertest.PerformRequest(r, "GET", "/recovery")

	// TEST
	assert.Equal(t, w.Code, 500)
	assert.Contains(t, buffer.String(), "Oupps, Houston, we have a problem")
	assert.Contains(t, buffer.String(), "TestPanicInHandler")
}
