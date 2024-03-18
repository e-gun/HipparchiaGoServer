package gen

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// JSONresponse - send the JSON; this function lets one test and document different strategies; jsr should be a json-ready struct
func JSONresponse(c echo.Context, jsr any) error {

	return c.JSON(http.StatusOK, jsr)

	// note that JSONPretty will end up strikingly prominent on the profiler: a waste of memory and cycles unless
	// you are debugging and want to be able to inspect the json manually

	// [1] "vanilla"; and it turns out there is nothing wrong with vanilla; seems like the best choice
	//opt1 := func() error { return c.JSON(http.StatusOK, jsr) }

	// [2] "costs a lot of RAM in return for what?"
	//opt2 := func() error { return c.JSONPretty(http.StatusOK, jsr, JSONINDENT) }

	// [3] "maybe streaming makes sense..." but this uses slightly more memory than [a] and is slightly slower?
	//opt3 := func() error {
	//	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	//	c.Response().WriteHeader(http.StatusOK)
	//	return json.NewEncoder(c.Response()).Encode(jsr)
	//}

	// [4] jsoniter: purportedly faster json, but we are one-big and not many-small...
	// requires: import jsoniter "github.com/json-iterator/go"
	// nb: not fully "ConfigCompatibleWithStandardLibrary" as it cannot do "JSONPretty"
	//

	//opt4 := func() error {
	//	b, e := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(&jsr)
	//	chke(e)
	//	return c.JSONBlob(http.StatusOK, b)
	//}

	//const (
	//	RESPONDER = 1
	//)
	//
	//switch RESPONDER {
	//case 1:
	//	return opt1()
	//case 2:
	//	return opt2()
	//case 3:
	//	return opt3()
	//case 4:
	//	return opt4()
	//default:
	//	return opt1()
	//}

	//return opt1()
}
