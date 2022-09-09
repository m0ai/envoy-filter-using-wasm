package main

import (
	"fmt"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func main() {
	proxywasm.SetVMContext(&vmContext{})
}

type vmContext struct {
	// Embed the default VM context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultVMContext
}

// Override types.DefaultVMContext.
func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{}
}

type pluginContext struct {
	// Embed the default plugin context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultPluginContext
	shouldEchoBody bool
}

// Override types.DefaultPluginContext.
func (ctx *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	if ctx.shouldEchoBody {
		return &echoBodyContext{}
	}

	return &customHttpContext{}
}

// Override types.DefaultPluginContext.
func (ctx *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	data, err := proxywasm.GetPluginConfiguration()
	if err != nil {
		proxywasm.LogCriticalf("error reading plugin configuration: %v", err)
	}
	ctx.shouldEchoBody = string(data) == "echo"
	return types.OnPluginStartStatusOK
}

type customHttpContext struct {
	// Embed the default root http context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultHttpContext
	totalRequestBodySize  int
	totalResponseBodySize int
	bufferOperation       string
	username              string
}

// Override types.DefaultHttpContext.
func (ctx *customHttpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	username, err := proxywasm.GetHttpRequestHeader("x-username")

	// Fallback if 'x-username' headers is empty
	if err == types.ErrorStatusNotFound {
		username = "Anonymous"
	}

	ctx.username = username
	return types.ActionContinue
}

// Override types.DefaultHttpContext.
func (ctx *customHttpContext) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
	// Remove Content-Length in order to prevent severs from crashing if we set different body.
	if err := proxywasm.RemoveHttpResponseHeader("content-length"); err != nil {
		panic(err)
	}

	return types.ActionContinue
}

// Override types.DefaultHttpContext.
func (ctx *customHttpContext) OnHttpResponseBody(bodySize int, endOfStream bool) types.Action {
	ctx.totalResponseBodySize += bodySize
	if !endOfStream {
		// Wait until we see the entire body to replace.
		return types.ActionPause
	}

	originalBody, err := proxywasm.GetHttpResponseBody(0, ctx.totalResponseBodySize)
	if err != nil {
		proxywasm.LogErrorf("failed to get response body: %v", err)
		return types.ActionContinue
	}
	proxywasm.LogInfof("original response %s body: %s", ctx.username, string(originalBody))

	err = proxywasm.AppendHttpResponseBody([]byte(
		fmt.Sprintf(", To %s", ctx.username)))

	if err != nil {
		proxywasm.LogErrorf("failed to %s response body: %v", ctx.bufferOperation, err)
		return types.ActionContinue
	}
	return types.ActionContinue
}

type echoBodyContext struct {
	// mbed the default plugin context
	// so that you don't need to reimplement all the methods by yourself.
	types.DefaultHttpContext
	totalRequestBodySize int
}

// Override types.DefaultHttpContext.
func (ctx *echoBodyContext) OnHttpRequestBody(bodySize int, endOfStream bool) types.Action {
	ctx.totalRequestBodySize += bodySize
	if !endOfStream {
		// Wait until we see the entire body to replace.
		return types.ActionPause
	}

	// Send the request body as the response body.
	body, _ := proxywasm.GetHttpRequestBody(0, ctx.totalRequestBodySize)
	if err := proxywasm.SendHttpResponse(200, nil, body, -1); err != nil {
		panic(err)
	}
	return types.ActionPause
}
