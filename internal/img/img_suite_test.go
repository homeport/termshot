// Copyright © 2020 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package img_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	. "github.com/homeport/termshot/internal/img"
)

func TestImg(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Image Creation Suite")
}

func testdata(path ...string) string {
	return filepath.Join(append([]string{"..", "..", "test", "data"}, path...)...)
}

func LookLike(path string) types.GomegaMatcher {
	return &LookLikeMatcher{path}
}

type LookLikeMatcher struct{ path string }

func (m *LookLikeMatcher) Match(actual interface{}) (bool, error) {
	scaffold, ok := actual.(Scaffold)
	if !ok {
		return false, fmt.Errorf("LookLike must be passed a Scaffold. Got\n%T", actual)
	}

	var out bytes.Buffer
	if err := scaffold.WritePNG(&out); err != nil {
		return false, err
	}

	// Uncomment to regenerate expected outputs
	//os.WriteFile(m.path, out.Bytes(), 0666)

	reference, err := os.ReadFile(m.path)
	if err != nil {
		return false, err
	}

	return bytes.Equal(out.Bytes(), reference), nil
}

func (matcher *LookLikeMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected scaffold to look like %s", matcher.path)
}

func (matcher *LookLikeMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected scaffold not to look like %s", matcher.path)
}
