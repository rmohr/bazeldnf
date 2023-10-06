package repo

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
)

func TestGetter(t *testing.T) {
	content := []byte("my file contents\n")
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("wrong method, %v instead of %v", r.Method, http.MethodGet)
		}
		n, err := rw.Write(content)
		if err != nil {
			t.Fatal("write content: ", err)
		}
		if n != len(content) {
			t.Fatalf("short write, %v instead of %v", n, len(content))
		}
	}))
	defer s.Close()

	localFile := path.Join(t.TempDir(), "contentfile")
	if err := os.WriteFile(localFile, content, os.ModePerm); err != nil {
		t.Fatalf("WriteFile %v failed: %v", localFile, err)
	}

	for _, tc := range []struct {
		name string
		url  string
	}{
		{
			name: "HTTP",
			url:  s.URL,
		},
		{
			name: "local",
			url:  "file://" + localFile,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Getter.Get %v", tc.url)
			resp, err := Getter(&getterImpl{}).Get(tc.url)
			if err != nil {
				t.Fatalf("Get %v: %v", tc.url, err)
			}
			defer resp.Body.Close()
			recv, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Read response failed: %v", err)
			}
			if !bytes.Equal(recv, content) {
				t.Fatalf("Read wrong content, %v instead of %v", recv, content)
			}
		})
	}
}
