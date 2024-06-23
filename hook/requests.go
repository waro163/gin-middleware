package hook

import (
	"context"
	"net/http"

	gmw "github.com/waro163/gin-middleware"
	"github.com/waro163/requests"
	"github.com/waro163/requests/hooks"
)

type AddHeadersHook struct {
	hooks.NoOps
}

var _ requests.Hook = (*AddHeadersHook)(nil)

// PrepareRequest handle this request before do requesting
func (h *AddHeadersHook) PrepareRequest(ctx context.Context, req *http.Request) error {
	gmw.InjectContextRequestHeader(ctx, req)
	return nil
}
