package repo

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
)

const retryAttempts = 5

func assertURLContent(t *testing.T, url string, content []byte) {
	t.Helper()
	client := retryablehttp.NewClient()
	client.RetryWaitMin = time.Microsecond
	client.RetryWaitMax = time.Microsecond
	client.RetryMax = retryAttempts

	t.Logf("Getter.Get %v", url)
	resp, err := Getter(&getterImpl{client: client}).Get(url)
	if err != nil {
		t.Fatalf("Get %v: %v", url, err)
	}
	defer resp.Body.Close()
	recv, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Read response failed: %v", err)
	}
	if !bytes.Equal(recv, content) {
		t.Fatalf("Read wrong content, %q instead of %q", string(recv), string(content))
	}
}

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
			assertURLContent(t, tc.url, content)
		})
	}
}

func TestGetterRetry(t *testing.T) {
	count := 0
	retries := retryAttempts
	content := []byte("my file contents\n")
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		count += 1

		if count < retries {
			t.Log("sending error response", http.StatusInternalServerError)
			rw.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(rw, "hopefully this is just temporary")
		} else {
			n, err := rw.Write(content)
			if err != nil {
				t.Fatal("write content: ", err)
			}
			if n != len(content) {
				t.Fatalf("short write, %v instead of %v", n, len(content))
			}
		}
	}))
	defer s.Close()

	assertURLContent(t, s.URL, content)
	if count != retries {
		t.Fatalf("expected to attempt request %v times instead of %v", retries, count)
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

func newPassthroughClient() *retryablehttp.Client {
	c := retryablehttp.NewClient()
	c.RetryMax = 0
	c.ErrorHandler = retryablehttp.PassthroughErrorHandler
	return c
}

func hijackConnection(t *testing.T, rw http.ResponseWriter) net.Conn {
	t.Helper()
	hj, ok := rw.(http.Hijacker)
	if !ok {
		t.Fatal("server doesn't support hijacking")
	}
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		t.Fatalf("hijack: %v", err)
	}
	bufrw.WriteString("HTTP/1.1 200 OK\r\nContent-Length: 10000\r\n\r\npartial")
	bufrw.Flush()
	return conn
}

func newTestFetcher(serverURL, cacheDir string) (*RepoFetcherImpl, *url.URL) {
	mirror, _ := url.Parse(serverURL + "/repo/")
	return &RepoFetcherImpl{
		Getter:      &getterImpl{client: newPassthroughClient()},
		CacheHelper: NewCacheHelper(cacheDir),
	}, mirror
}

func TestFetchFileRetry(t *testing.T) {
	content := []byte("test repository content")
	sum := sha256.Sum256(content)
	checksum := hex.EncodeToString(sum[:])

	repomd := func() *api.Repomd {
		d := api.Data{}
		d.Type = api.PrimaryFileType
		d.Checksum.Text = checksum
		d.Checksum.Type = "sha256"
		d.Location.Href = "repodata/primary.xml.gz"
		return &api.Repomd{Data: []api.Data{d}}
	}()

	repo := &bazeldnf.Repository{Name: "test-repo"}

	transferErrors := []struct {
		error     string
		closeConn func(net.Conn)
	}{
		{"unexpected EOF", func(c net.Conn) { c.Close() }},
		{"connection reset", func(c net.Conn) { c.(*net.TCPConn).SetLinger(0); c.Close() }},
	}

	for _, tc := range transferErrors {
		t.Run("retries on "+tc.error, func(t *testing.T) {
			calls := 0
			s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				calls++
				if calls == 1 {
					tc.closeConn(hijackConnection(t, rw))
					return
				}
				rw.Write(content)
			}))
			defer s.Close()

			fetcher, mirror := newTestFetcher(s.URL, t.TempDir())
			if err := fetcher.fetchFile(api.PrimaryFileType, repo, repomd, mirror); err != nil {
				t.Fatalf("expected success after retry, got: %v", err)
			}
			if calls != 2 {
				t.Fatalf("expected 2 calls, got %d", calls)
			}
		})
	}

	for _, tc := range transferErrors {
		t.Run("exhausted retries on "+tc.error, func(t *testing.T) {
			calls := 0
			s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				calls++
				tc.closeConn(hijackConnection(t, rw))
			}))
			defer s.Close()

			fetcher, mirror := newTestFetcher(s.URL, t.TempDir())
			err := fetcher.fetchFile(api.PrimaryFileType, repo, repomd, mirror)
			if err == nil {
				t.Fatal("expected error after exhausted retries, got nil")
			}
			if calls != transferRetries+1 {
				t.Fatalf("expected %d calls, got %d", transferRetries+1, calls)
			}
			if !strings.Contains(err.Error(), tc.error) {
				t.Fatalf("expected error to mention %q, got: %v", tc.error, err)
			}
		})
	}

	t.Run("no retry on 4xx", func(t *testing.T) {
		calls := 0
		s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			calls++
			rw.WriteHeader(http.StatusNotFound)
		}))
		defer s.Close()

		fetcher, mirror := newTestFetcher(s.URL, t.TempDir())
		if err := fetcher.fetchFile(api.PrimaryFileType, repo, repomd, mirror); err == nil {
			t.Fatal("expected error, got nil")
		}
		if calls != 1 {
			t.Fatalf("expected 1 call (no retry), got %d", calls)
		}
	})

	t.Run("no retry on filesystem error", func(t *testing.T) {
		calls := 0
		s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			calls++
			rw.Write(content)
		}))
		defer s.Close()

		fetcher, mirror := newTestFetcher(s.URL, "/dev/null/impossible")
		if err := fetcher.fetchFile(api.PrimaryFileType, repo, repomd, mirror); err == nil {
			t.Fatal("expected error, got nil")
		}
		if calls != 1 {
			t.Fatalf("expected 1 call (no retry), got %d", calls)
		}
	})
}
