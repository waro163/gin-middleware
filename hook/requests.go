package hook

import (
	"context"
	"net/http"

	gmw "github.com/waro163/gin-middleware"
	"github.com/waro163/requests"
)

type AddHeadersHook struct{}

var _ requests.Hook = (*AddHeadersHook)(nil)

// PrepareRequest handle this request before do requesting
func (h *AddHeadersHook) PrepareRequest(ctx context.Context, req *http.Request) error {
	gmw.InjectContextRequestHeader(ctx, req)
	return nil
}

// OnRequestError is for when PrepareRequest return error, it will handle this error, if this function also return error, all the work will stop
func (h *AddHeadersHook) OnRequestError(context.Context, *http.Request, error) error {
	return nil
}

// ProcessResponse handle this request and response after do requesting
func (h *AddHeadersHook) ProcessResponse(context.Context, *http.Request, *http.Response) error {
	return nil
}

// OnResponseError is for when ProcessResponse return error, it will handle this error, if this function also return error, all the work will stop
func (h *AddHeadersHook) OnResponseError(context.Context, *http.Request, *http.Response, error) error {
	return nil
}
