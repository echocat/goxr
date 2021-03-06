package server

import (
	"net/http"
	"os"
	"time"
)

var (
	noHttpTime = HttpTime(time.Time{})
)

func ErrorToHttpResponse(err error) JsonResponse {
	r := JsonResponse{
		Code: http.StatusInternalServerError,
	}
	if os.IsNotExist(err) {
		r.Code = http.StatusNotFound
	}
	if os.IsPermission(err) {
		r.Code = http.StatusForbidden
	}
	return r
}
