package sysx_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/polarixa/replify/pkg/sysx"
)

// ---------------------------------------------------------------------------
// MimeFromName
// ---------------------------------------------------------------------------

func TestMimeFromName_Cases(t *testing.T) {
	cases := map[string]string{
		"a.csv":     sysx.MimeCSV,
		"b.json":    sysx.MimeJSON,
		"c.pdf":     sysx.MimePDF,
		"d.tar.gz":  sysx.MimeGZIP,
		"e.tgz":     sysx.MimeGZIP,
		"f.HTML":    sysx.MimeHTML,
		"g.unknown": sysx.MimeOctetStream,
		"noext":     sysx.MimeOctetStream,
		"trailing.": sysx.MimeOctetStream,
		"":          sysx.MimeOctetStream,
		"log.LOG":   sysx.MimeText,
		"sql.sql":   sysx.MimeSQL,
		"x.xml":     sysx.MimeXML,
		"z.zip":     sysx.MimeZIP,
	}
	for in, want := range cases {
		if got := sysx.MimeFromName(in); got != want {
			t.Errorf("MimeFromName(%q) = %q, want %q", in, got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// MemBlob
// ---------------------------------------------------------------------------

func TestMemBlob_RoundTrip(t *testing.T) {
	data := []byte("hello world")
	blob := sysx.NewMemBlob(data)

	if blob.Len() != int64(len(data)) {
		t.Fatalf("Len = %d, want %d", blob.Len(), len(data))
	}
	if !bytes.Equal(blob.Bytes(), data) {
		t.Fatalf("Bytes mismatch")
	}

	got, err := io.ReadAll(blob)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("read = %q, want %q", got, data)
	}

	if _, err := blob.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("Seek: %v", err)
	}
	got2, _ := io.ReadAll(blob)
	if !bytes.Equal(got2, data) {
		t.Fatalf("re-read = %q", got2)
	}

	if err := blob.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// Close is idempotent.
	if err := blob.Close(); err != nil {
		t.Fatalf("double Close: %v", err)
	}
}

func TestMemBlob_NilSafe(t *testing.T) {
	var blob *sysx.MemBlob
	if got := blob.Len(); got != 0 {
		t.Fatalf("nil Len = %d", got)
	}
	if got := blob.Bytes(); got != nil {
		t.Fatalf("nil Bytes = %v", got)
	}
	if err := blob.Close(); err != nil {
		t.Fatalf("nil Close = %v", err)
	}
}

// ---------------------------------------------------------------------------
// TempFile
// ---------------------------------------------------------------------------

func TestTempFile_NewAndClose(t *testing.T) {
	tf, err := sysx.NewTempFile()
	if err != nil {
		t.Fatalf("NewTempFile: %v", err)
	}
	if tf.Path() == "" {
		t.Fatalf("Path empty")
	}
	if !tf.RemoveOnClose() {
		t.Fatalf("RemoveOnClose should default to true")
	}

	if _, err := tf.Write([]byte("data")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := tf.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("Seek: %v", err)
	}
	got, _ := io.ReadAll(tf)
	if string(got) != "data" {
		t.Fatalf("read = %q", got)
	}

	path := tf.Path()
	if err := tf.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if !tf.Closed() {
		t.Fatalf("Closed should be true after Close")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("temp file still on disk: %v", err)
	}

	// Idempotent.
	if err := tf.Close(); err != nil {
		t.Fatalf("double Close: %v", err)
	}
}

func TestTempFile_NamedAndKept(t *testing.T) {
	tf, err := sysx.NewTempFilename("sysx-keep-*.dat")
	if err != nil {
		t.Fatalf("NewTempFileNamed: %v", err)
	}
	if !strings.HasSuffix(tf.Path(), ".dat") {
		t.Fatalf("pattern not honoured: %s", tf.Path())
	}

	tf.WithRemoveOnClose(false)
	path := tf.Path()
	if err := tf.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	defer os.Remove(path)

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("WithRemoveOnClose(false) did not preserve file: %v", err)
	}
}

func TestTempFile_AtCustomDir(t *testing.T) {
	dir := t.TempDir()
	tf, err := sysx.NewTempFileAt(dir, "scoped-*.bin")
	if err != nil {
		t.Fatalf("NewTempFileAt: %v", err)
	}
	defer tf.Close()
	if filepath.Dir(tf.Path()) != dir {
		t.Fatalf("temp file in %s, want %s", filepath.Dir(tf.Path()), dir)
	}
	stat, err := tf.Stat()
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if stat.Size() != 0 {
		t.Fatalf("fresh temp file size = %d", stat.Size())
	}
}

// ---------------------------------------------------------------------------
// Resource builder + lifecycle
// ---------------------------------------------------------------------------

func TestResource_FromBytes(t *testing.T) {
	payload := []byte("id,name\n1,john\n")
	res := sysx.NewResource().
		WithName("user-report.csv").
		FromBytes(payload)
	defer res.Close()

	if res.Name() != "user-report.csv" {
		t.Fatalf("Name = %q", res.Name())
	}
	if res.Size() != int64(len(payload)) {
		t.Fatalf("Size = %d", res.Size())
	}
	if res.ContentType() != sysx.MimeCSV {
		t.Fatalf("ContentType = %q", res.ContentType())
	}

	var buf bytes.Buffer
	n, err := res.CopyTo(&buf)
	if err != nil {
		t.Fatalf("CopyTo: %v", err)
	}
	if n != int64(len(payload)) || !bytes.Equal(buf.Bytes(), payload) {
		t.Fatalf("payload mismatch: %q", buf.String())
	}
}

func TestResource_FromString_OverrideMime(t *testing.T) {
	res := sysx.NewResource().
		WithName("note.txt").
		WithContentType(sysx.MimeOctetStream).
		FromString("hi")
	defer res.Close()

	if res.ContentType() != sysx.MimeOctetStream {
		t.Fatalf("explicit ContentType ignored: %q", res.ContentType())
	}
	if res.Size() != 2 {
		t.Fatalf("Size = %d", res.Size())
	}
}

func TestResource_FromTempFile_AutoCleanup(t *testing.T) {
	res, err := sysx.NewResource().
		WithName("audit.txt").
		WithTempPattern("sysx-audit-*.txt").
		FromTempFile(func(w io.Writer) error {
			_, err := io.WriteString(w, "audit-line")
			return err
		})
	if err != nil {
		t.Fatalf("FromTempFile: %v", err)
	}

	tf, ok := res.Content().(*sysx.TempFile)
	if !ok {
		t.Fatalf("expected *TempFile, got %T", res.Content())
	}
	path := tf.Path()
	if res.Size() != int64(len("audit-line")) {
		t.Fatalf("Size = %d", res.Size())
	}

	got, err := io.ReadAll(res.Content())
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(got) != "audit-line" {
		t.Fatalf("payload = %q", got)
	}

	if err := res.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("temp file leaked: %v", err)
	}
}

func TestResource_FromTempFile_KeepDisabled(t *testing.T) {
	res, err := sysx.NewResource().
		WithName("keep.txt").
		WithRemoveOnClose(false).
		FromTempFile(func(w io.Writer) error {
			_, err := io.WriteString(w, "kept")
			return err
		})
	if err != nil {
		t.Fatalf("FromTempFile: %v", err)
	}
	tf := res.Content().(*sysx.TempFile)
	path := tf.Path()
	defer os.Remove(path)

	if err := res.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("WithRemoveOnClose(false) did not preserve file: %v", err)
	}
}

func TestResource_FromTempFile_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	_, err := sysx.NewResource().
		FromTempFile(func(w io.Writer) error { return want })
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
}

func TestResource_FromReader_InMemory(t *testing.T) {
	src := strings.NewReader("streamed")
	res, err := sysx.NewResource().
		WithName("stream.txt").
		FromReader(src)
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	defer res.Close()

	if res.Size() != int64(len("streamed")) {
		t.Fatalf("Size = %d", res.Size())
	}
	if _, ok := res.Content().(*sysx.TempFile); ok {
		t.Fatalf("did not expect spill to disk")
	}
	out, _ := io.ReadAll(res.Content())
	if string(out) != "streamed" {
		t.Fatalf("payload = %q", out)
	}
}

func TestResource_FromReader_SpillsToDisk(t *testing.T) {
	payload := bytes.Repeat([]byte("x"), 64<<10) // 64 KiB
	res, err := sysx.NewResource().
		WithName("big.bin").
		WithSpillThreshold(4 << 10). // 4 KiB
		FromReader(bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	defer res.Close()

	if res.Size() != int64(len(payload)) {
		t.Fatalf("Size = %d", res.Size())
	}
	n, err := res.Drain()
	if err != nil {
		t.Fatalf("Drain: %v", err)
	}
	if n != int64(len(payload)) {
		t.Fatalf("drained = %d", n)
	}
}

func TestResource_FromReader_NilSrc(t *testing.T) {
	_, err := sysx.NewResource().FromReader(nil)
	if !errors.Is(err, sysx.ErrNilResource) {
		t.Fatalf("err = %v", err)
	}
}

func TestResource_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "adopted.txt")
	if err := os.WriteFile(path, []byte("adopted"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	res, err := sysx.NewResource().
		WithRemoveOnClose(false).
		FromFile(f)
	if err != nil {
		t.Fatalf("FromFile: %v", err)
	}
	defer res.Close()

	if res.Name() != "adopted.txt" {
		t.Fatalf("Name = %q", res.Name())
	}
	if res.Size() != 7 {
		t.Fatalf("Size = %d", res.Size())
	}
	if res.ContentType() != sysx.MimeText {
		t.Fatalf("ContentType = %q", res.ContentType())
	}

	got, _ := io.ReadAll(res.Content())
	if string(got) != "adopted" {
		t.Fatalf("payload = %q", got)
	}
}

func TestResource_FromFile_NilSafe(t *testing.T) {
	_, err := sysx.NewResource().FromFile(nil)
	if !errors.Is(err, sysx.ErrNilResource) {
		t.Fatalf("err = %v", err)
	}
}

func TestResource_RewindAndDoubleRead(t *testing.T) {
	res := sysx.NewResource().WithName("a.txt").FromString("abc")
	defer res.Close()

	a, _ := io.ReadAll(res.Content())
	if err := res.Rewind(); err != nil {
		t.Fatalf("Rewind: %v", err)
	}
	b, _ := io.ReadAll(res.Content())
	if !bytes.Equal(a, b) {
		t.Fatalf("rewind failed: %q vs %q", a, b)
	}
}

func TestResource_NilSafe(t *testing.T) {
	var res *sysx.Resource
	if res.Name() != "" {
		t.Fatalf("nil Name not empty")
	}
	if res.Size() != 0 {
		t.Fatalf("nil Size != 0")
	}
	if res.Content() != nil {
		t.Fatalf("nil Content not nil")
	}
	if err := res.Close(); err != nil {
		t.Fatalf("nil Close: %v", err)
	}
	if err := res.Rewind(); !errors.Is(err, sysx.ErrNilResource) {
		t.Fatalf("nil Rewind err = %v", err)
	}
	if _, err := res.CopyTo(io.Discard); !errors.Is(err, sysx.ErrNilResource) {
		t.Fatalf("nil CopyTo err = %v", err)
	}
}

func TestResource_WithContent(t *testing.T) {
	custom := sysx.NewMemBlob([]byte("custom"))
	res := sysx.NewResource().
		WithName("c.bin").
		WithContent(custom).
		WithSize(int64(custom.Len())).
		WithContentType(sysx.MimeOctetStream)

	if res.Content() != custom {
		t.Fatalf("content not attached")
	}
	if res.Size() != 6 {
		t.Fatalf("Size = %d", res.Size())
	}
	if err := res.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestResource_AccessorsReflectBuilder(t *testing.T) {
	res := sysx.NewResource().
		WithSpillThreshold(1 << 12).
		WithTempPattern("sysx-acc-*.bin").
		WithTempDir("/tmp").
		WithRemoveOnClose(false)

	if res.SpillThreshold() != 1<<12 {
		t.Fatalf("SpillThreshold = %d", res.SpillThreshold())
	}
	if res.TempPattern() != "sysx-acc-*.bin" {
		t.Fatalf("TempPattern = %q", res.TempPattern())
	}
	if res.TempDir() != "/tmp" {
		t.Fatalf("TempDir = %q", res.TempDir())
	}
	if res.RemoveOnClose() {
		t.Fatalf("RemoveOnClose should be false")
	}

	// SpillThreshold sanitises non-positive values.
	res.WithSpillThreshold(0)
	if res.SpillThreshold() != sysx.DefaultSpillThreshold {
		t.Fatalf("SpillThreshold reset = %d", res.SpillThreshold())
	}
}

// ---------------------------------------------------------------------------
// End-to-end consumer scenario
// ---------------------------------------------------------------------------

// fakeUploader is a stand-in for an S3 / Telegram client that consumes a
// Resource without knowing how the bytes are stored.
type fakeUploader struct {
	name  string
	size  int64
	mime  string
	bytes []byte
}

func (u *fakeUploader) Put(res *sysx.Resource) error {
	defer res.Close()
	u.name = res.Name()
	u.size = res.Size()
	u.mime = res.ContentType()
	b, err := io.ReadAll(res.Content())
	if err != nil {
		return err
	}
	u.bytes = b
	return nil
}

func TestResource_ConsumerAgnosticToBacking(t *testing.T) {
	payload := []byte("abc,def\n1,2\n")

	memRes := sysx.NewResource().WithName("a.csv").FromBytes(payload)
	tmpRes, err := sysx.NewResource().
		WithName("b.csv").
		FromTempFile(func(w io.Writer) error {
			_, err := w.Write(payload)
			return err
		})
	if err != nil {
		t.Fatalf("FromTempFile: %v", err)
	}
	streamRes, err := sysx.NewResource().
		WithName("c.csv").
		FromReader(bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}

	for _, r := range []*sysx.Resource{memRes, tmpRes, streamRes} {
		u := &fakeUploader{}
		if err := u.Put(r); err != nil {
			t.Fatalf("Put: %v", err)
		}
		if u.size != int64(len(payload)) || !bytes.Equal(u.bytes, payload) {
			t.Fatalf("uploader mismatch for %s: size=%d, payload=%q", u.name, u.size, u.bytes)
		}
		if u.mime != sysx.MimeCSV {
			t.Fatalf("mime = %q", u.mime)
		}
	}
}
