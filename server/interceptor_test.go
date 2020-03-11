package server

import (
	"github.com/echocat/goxr"
	"github.com/echocat/goxr/box/fs"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"testing"
)

func Test_Server_embedding_regular(t *testing.T) {
	instance := &testInterceptor{
		t:                             t,
		toReturnOnBeforeHandleContext: &fasthttp.RequestCtx{},
		toReturnOnBeforeHandleBox:     &fs.Box{},
	}
	s := Server{
		Box:         &fs.Box{},
		Interceptor: instance,
	}

	givenCtx := &fasthttp.RequestCtx{}

	s.Handle(givenCtx)

	assert.Same(t, s.Box, instance.onBeforeHandleBox)
	assert.Same(t, givenCtx, instance.onBeforeHandleContext)
	assert.Same(t, instance.toReturnOnBeforeHandleContext, instance.onAfterHandleContext)
	assert.Same(t, instance.toReturnOnBeforeHandleBox, instance.onAfterHandleBox)
}
func Test_Server_embedding_errors(t *testing.T) {
	instance := &testInterceptor{
		t:                            t,
		toReturnOnHandleErrorContext: &fasthttp.RequestCtx{},
	}
	box, err := fs.OpenBox("../resources/testBase1")
	assert.Nil(t, err)
	assert.NotNil(t, box)
	s := Server{
		Box:         box,
		Interceptor: instance,
	}

	givenBox := &fs.Box{}
	givenCtx := &fasthttp.RequestCtx{}
	givenError := errors.New("foobar")

	s.HandleError(givenBox, givenError, true, givenCtx)

	assert.Same(t, givenBox, instance.onHandleErrorBox)
	assert.Same(t, givenCtx, instance.onHandleErrorContext)
	assert.Same(t, givenError, instance.onHandleErrorError)
}

type testInterceptor struct {
	t *testing.T

	toReturnOnBeforeHandleContext *fasthttp.RequestCtx
	toReturnOnBeforeHandleBox     goxr.Box
	toReturnOnHandleErrorContext  *fasthttp.RequestCtx

	onBeforeHandleContext *fasthttp.RequestCtx
	onBeforeHandleBox     goxr.Box
	onAfterHandleContext  *fasthttp.RequestCtx
	onAfterHandleBox      goxr.Box
	onHandleErrorContext  *fasthttp.RequestCtx
	onHandleErrorBox      goxr.Box
	onHandleErrorError    error
}

func (instance *testInterceptor) OnBeforeHandle(box goxr.Box, ctx *fasthttp.RequestCtx) (handled bool, newBox goxr.Box, newCtx *fasthttp.RequestCtx) {
	assert.NotNil(instance.t, ctx)

	instance.onBeforeHandleBox = box
	instance.onBeforeHandleContext = ctx
	return true, instance.toReturnOnBeforeHandleBox, instance.toReturnOnBeforeHandleContext
}

func (instance *testInterceptor) OnAfterHandle(box goxr.Box, ctx *fasthttp.RequestCtx) {
	assert.NotNil(instance.t, ctx)

	instance.onAfterHandleBox = box
	instance.onAfterHandleContext = ctx
}

func (instance *testInterceptor) OnTargetPathResolved(goxr.Box, string, *fasthttp.RequestCtx) string {
	panic("not implemented")
}

func (instance *testInterceptor) OnHandleError(box goxr.Box, err error, _ bool, ctx *fasthttp.RequestCtx) (handled bool, newErr error, newCtx *fasthttp.RequestCtx) {
	assert.NotNil(instance.t, err)
	assert.NotNil(instance.t, ctx)

	instance.onHandleErrorBox = box
	instance.onHandleErrorContext = ctx
	instance.onHandleErrorError = err
	return true, err, instance.onHandleErrorContext
}

func (instance *testInterceptor) OnAccessLog(goxr.Box, *fasthttp.RequestCtx, *map[string]interface{}) (handled bool) {
	panic("not implemented")
}
