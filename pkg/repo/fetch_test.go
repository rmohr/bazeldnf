package repo

import (
	"bytes"
	"encoding/base64"
	"fmt"
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

func TestNetrc(t *testing.T) {
	user := "user"
	password := "secret_whispers"
	authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, password))))
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		t.Logf("Auth header: %v", r.Header.Get("Authorization"))
		if r.Header.Get("Authorization") != authHeader {
			rw.WriteHeader(http.StatusUnauthorized)
		} else {
			rw.WriteHeader(http.StatusOK)
		}
	}))
	defer s.Close()

	// Do an unauthenticated get and confirm the server denies that.
	resp, err := Getter(&getterImpl{}).Get(s.URL)
	if err != nil {
		t.Fatalf("Get %v: %v", s.URL, err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("Close response failed: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("We shouldn't be supplying credentials, so the server should reply with 401 but got %d", resp.StatusCode)
	}

	// Set the netrc file and confirm the header gets set.
	netrcFile := path.Join(t.TempDir(), ".netrc")
	if err := os.Setenv("NETRC", netrcFile); err != nil {
		t.Fatalf("Setting NETRC env var failed: %v", err)
	}
	netrcContent := fmt.Sprintf(`machine 127.0.0.1
login %s
password %s`, user, password)
	if err := os.WriteFile(netrcFile, []byte(netrcContent), os.ModePerm); err != nil {
		t.Fatalf("writing netrc contents to %s failed: %v", netrcFile, err)
	}

	if err := os.Setenv("NETRC", netrcFile); err != nil {
		t.Fatalf("Setting NETRC env var failed: %v", err)
	}
	resp, err = Getter(&getterImpl{}).Get(s.URL)
	if err != nil {
		t.Fatalf("Get %v: %v", s.URL, err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("Close response failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("We've set NETRC so the server should reply with 200 but got %d", resp.StatusCode)
	}
}
