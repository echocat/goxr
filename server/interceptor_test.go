package server

import (
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
	}
	box, err := fs.OpenBox("../resources/testBase1")
	assert.Nil(t, err)
	assert.NotNil(t, box)
	s := Server{
		Box:         box,
		Interceptor: instance,
	}

	givenCtx := &fasthttp.RequestCtx{}

	s.Handle(givenCtx)

	assert.Same(t, givenCtx, instance.onBeforeHandleContext)
	assert.Same(t, instance.toReturnOnBeforeHandleContext, instance.onAfterHandleContext)
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

	givenCtx := &fasthttp.RequestCtx{}
	givenError := errors.New("foobar")

	s.HandleError(givenError, true, givenCtx)

	assert.Same(t, givenCtx, instance.onHandleErrorContext)
	assert.Same(t, givenError, instance.onHandleErrorError)
}

type testInterceptor struct {
	t *testing.T

	toReturnOnBeforeHandleContext *fasthttp.RequestCtx
	toReturnOnHandleErrorContext  *fasthttp.RequestCtx

	onBeforeHandleContext *fasthttp.RequestCtx
	onAfterHandleContext  *fasthttp.RequestCtx
	onHandleErrorContext  *fasthttp.RequestCtx
	onHandleErrorError    error
}

func (instance *testInterceptor) OnBeforeHandle(ctx *fasthttp.RequestCtx) (handled bool, newCtx *fasthttp.RequestCtx) {
	assert.NotNil(instance.t, ctx)

	instance.onBeforeHandleContext = ctx
	return true, instance.toReturnOnBeforeHandleContext
}

func (instance *testInterceptor) OnAfterHandle(ctx *fasthttp.RequestCtx) {
	assert.NotNil(instance.t, ctx)

	instance.onAfterHandleContext = ctx
}

func (instance *testInterceptor) OnTargetPathResolved(path string, ctx *fasthttp.RequestCtx) (newPath string) {
	panic("not implemented")
}

func (instance *testInterceptor) OnHandleError(err error, interceptAllowed bool, ctx *fasthttp.RequestCtx) (handled bool, newErr error, newCtx *fasthttp.RequestCtx) {
	assert.NotNil(instance.t, err)
	assert.NotNil(instance.t, ctx)

	instance.onHandleErrorContext = ctx
	instance.onHandleErrorError = err
	return true, err, instance.onHandleErrorContext
}
