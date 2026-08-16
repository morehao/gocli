package httpbingo

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/morehao/go-ark-template/demo/config"
	"github.com/morehao/go-ark-template/pkg/testsetup"
	"github.com/morehao/golib/gutil"
	"github.com/stretchr/testify/assert"
)

func findConfigPath(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..", "config", "config.yaml")
}

func TestGet(t *testing.T) {
	testsetup.Initialize(testsetup.AppNameDemo)
	config.LoadConfig(findConfigPath(t))

	ctx := testsetup.NewContext()
	res, err := Get(ctx, &GetRequest{
		ID: 1,
	})
	assert.Nil(t, err)
	t.Log(gutil.ToJsonString(res))
}
