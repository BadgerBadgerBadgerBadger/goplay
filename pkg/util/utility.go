package util

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
	"golang.org/x/exp/constraints"
)

type Numeric interface {
	constraints.Signed | constraints.Float
}

func Sign[T Numeric](inp T) T {

	if inp >= 0 {
		return 1
	}

	return -1
}

func RandNegOneOne() float64 {
	return (rand.Float64() * 2) - 1
}

func Map[T Numeric](input, inputStart, inputEnd, outputStart, outputEnd T) T {
	return outputStart + ((outputEnd-outputStart)/(inputEnd-inputStart))*(input-inputStart)

}

func Line(x float64) float64 {

	// y = mx + b
	return (0.3 * x) + 0.2
}

func Sigmoid(x float64) float64 {
	return 1 / (1 + math.Exp(-x))
}

func IsNonNilPointer(v interface{}) bool {

	if v == nil {
		return false
	}

	rv := reflect.ValueOf(v)

	return rv.Kind() != reflect.Ptr || rv.IsNil()
}

// OptimisticAtoi is the, well, optimistic version of the stdlib's
// Atoi function. It assumes the conversion will succeed for sure.
// Don't use in production.
func OptimisticAtoi(inp string) int {

	out, err := strconv.Atoi(inp)
	Must(err, FTWSTHf("%s was supposed to be a number 😐", inp))

	return out
}

// FTWSTHf adds a message at the end of the given string,
// indicating the developer's frustration. It takes a format
// string as its main input so go nuts.
func FTWSTHf(formatString string, vals ...interface{}) string {

	formatString += "; fuck, that wasn't supposed to happen!"

	return fmt.Sprintf(formatString, vals...)
}

func NewImageReaderFromUrl(endpoint string) (io.ReadCloser, error) {

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp.Body, nil
}

func Sample[T any](slice []T) T {

	l := len(slice)
	index := rand.Intn(l)

	return slice[index]
}

// ApplyOnType takes the address of type `t` and an array of functions
// that do something with this type, and applies the functions to the type.
func ApplyOnType[T any](t *T, tFuncs []func(*T)) {
	for _, optFunc := range tFuncs {
		optFunc(t)
	}
}
