package util

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func ReadBodyAsString(r *http.Request) string {

	body, err := ioutil.ReadAll(r.Body)
	Must(err, "failed to read body")

	return string(body)
}

type jsonRespOpt struct {
	statusCode int
}

type jSONRespOptFunc = func(opt *jsonRespOpt)

// WithStatusCode allows modifying the status code of the
// JSON response. The default is 200 OK.
func WithStatusCode(statusCode int) jSONRespOptFunc {
	return func(opt *jsonRespOpt) {
		opt.statusCode = statusCode
	}
}

// JSONResponse marshalls the argument `out` into JSON and response on the provided
// ResponseWriter. Options can be provided to further modidy the response.
func JSONResponse(w http.ResponseWriter, out any, optfuncs ...jSONRespOptFunc) {

	jsonRspO := jsonRespOpt{
		statusCode: 200,
	}
	ApplyOnType(&jsonRspO, optfuncs)

	resp, err := json.Marshal(out)
	if err != nil {
		JSONResponse(
			w,
			"Oops! That's not right! Send me an email and I'll look into it...probably.",
			WithStatusCode(500),
		)
		return
	}

	w.WriteHeader(jsonRspO.statusCode)
	_, err = w.Write(resp)
	LogOnError(err, "failed to write response")
}
