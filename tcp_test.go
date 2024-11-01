package sockfile

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	s server
	c client
)

func init() {
	currentDir, _ = os.Getwd()
	s = server{
		addr:      "0.0.0.0:51122",
		poolSize:  make(chan struct{}, 4),
		fo:        newFileOps(),
		defautDir: filepath.Join(currentDir, "testData", "tcpTest"),
	}

	c = client{
		others: []string{"127.0.0.1:51122"},
		fo:     newFileOps(),
	}
}

func TestSendFile(t *testing.T) {
	defer s.close()
	var e error
	go func() {
		err := s.listen()
		if err != nil {
			e = err
		}
	}()
	if e != nil {
		t.Fatal(e)
	}
	time.Sleep(time.Second * 2)
	e = c.transToOthers(filepath.Join(currentDir, "testData", "tarTest.tar.gz"))
	if e != nil {
		t.Fatal(e)
	}
	time.Sleep(time.Second * 5)
}
