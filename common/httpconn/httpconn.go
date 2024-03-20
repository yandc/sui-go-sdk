package httpconn

import (
	"bytes"
	"context"
	"strings"
	"encoding/json"
	"github.com/block-vision/sui-go-sdk/models"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const defaultTimeout = time.Second * 5

type HttpConn struct {
	c       *http.Client
	rl      *rate.Limiter
	rpcUrl  string
	timeout time.Duration
}

func newDefaultRateLimiter() *rate.Limiter {
	rateLimiter := rate.NewLimiter(rate.Every(1*time.Second), 10000) // 10000 request every 1 seconds
	return rateLimiter
}

func NewHttpConn(rpcUrl string) *HttpConn {
	return &HttpConn{
		c:       &http.Client{},
		rpcUrl:  rpcUrl,
		timeout: defaultTimeout,
	}
}

func NewCustomHttpConn(rpcUrl string, cli *http.Client) *HttpConn {
	return &HttpConn{
		c:       cli,
		rpcUrl:  rpcUrl,
		timeout: defaultTimeout,
	}
}

func (h *HttpConn) Request(ctx context.Context, op Operation) ([]byte, error) {
	method := op.Method
	method = strings.Replace(method, "sui_", "bfc_", 1)
	method = strings.Replace(method, "suix_", "bfcx_", 1)
	jsonRPCReq := models.JsonRPCRequest{
		JsonRPC: "2.0",
		ID:      time.Now().UnixMilli(),
		Method:  method,
		Params:  op.Params,
	}
	reqBytes, err := json.Marshal(jsonRPCReq)
	if err != nil {
		return []byte{}, err
	}

	request, err := http.NewRequest("POST", h.rpcUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return []byte{}, err
	}
	request = request.WithContext(ctx)
	request.Header.Add("Content-Type", "application/json")
	rsp, err := h.c.Do(request.WithContext(ctx))
	if err != nil {
		return []byte{}, err
	}
	defer rsp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return []byte{}, err
	}
	return bodyBytes, nil
}
