// +build !short

package eventgrid

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Azure/buffalo-azure/sdk/eventgrid"
	"github.com/gobuffalo/buffalo/meta"
)

func TestGenerator_Run(t *testing.T) {
	const bufCmd, depCmd = "buffalo", "dep"
	requiredTools := []string{bufCmd, depCmd, "go", "node"}

	const appName = "gentest"
	for _, tool := range requiredTools {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not found on system", tool)
			return
		}
	}

	subject := Generator{}

	testLoc := path.Join(os.Getenv("GOPATH"), "src")

	loc, err := ioutil.TempDir(testLoc, "buffalo-azure_eventgrid_test")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(loc)
	t.Log("Output Location: ", loc)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Minute)
	defer cancel()

	var outHandle, errHandle io.Writer
	outHandle, err = os.Create(path.Join(loc, "buffalo_stdout.txt"))
	if err != nil {
		t.Logf("not able to harness %s stdout", bufCmd)
		outHandle = ioutil.Discard
	}
	errHandle, err = os.Create(path.Join(loc, "buffalo_stderr.txt"))
	if err != nil {
		t.Logf("not able to harness %s stderr", bufCmd)
		errHandle = ioutil.Discard
	}

	bufCreater := exec.CommandContext(ctx, bufCmd, "new", appName, "--with-dep")
	bufCreater.Dir = loc
	bufCreater.Stdout = outHandle
	bufCreater.Stderr = errHandle
	if err := bufCreater.Run(); err != nil {
		t.Error(err)
		return
	}

	fakeApp := meta.App{
		Root:       filepath.Join(loc, appName),
		ActionsPkg: "github.com/marstr/musicvotes/actions",
	}

	faux := eventgrid.SubscriptionValidationRequest{}

	if err = subject.Run(fakeApp, "ingress", map[string]reflect.Type{
		"Microsoft.EventGrid.SubscriptionValidation": reflect.TypeOf(faux),
	}); err != nil {
		t.Error(err)
		return
	}
}
