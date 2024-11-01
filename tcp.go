package sockfile

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"errors"
)

type server struct {
	addr      string
	poolSize  chan struct{}
	fo        *FileOps
	defautDir string
	l         net.Listener
}

func (s *server) listen() error {
	var err error
	s.l, err = net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	fmt.Println("starting server...")
	for {
		conn, err := s.l.Accept()
		if err != nil {
			continue
		}

		go s.handle(conn)
	}
}

func (s *server) close() error {
	return s.l.Close()
}

func (s *server) ack(conn net.Conn, err error) {
	if err != nil {
		conn.Write([]byte("fail: " + err.Error()))
	}
	conn.Write([]byte("success"))
}

func (s *server) handle(c net.Conn) {
	defer func() {
		time.Sleep(time.Second * 3)
		c.Close()
	}()

	headers, _ := bufio.NewReader(c).ReadString('\n')

	info := strings.Split(headers[:len(headers)-1], ",")

	filename := info[0]
	size, _ := strconv.Atoi(info[1])

	fullPath := filepath.Join(s.defautDir, filename)
	s.fo.removeIfExist(fullPath)

	// var size int64
	// binary.Read(c, binary.LittleEndian, &size)
	// fmt.Printf("[server] filename: %s, size: %d\n", filename, size)

	file, err := os.Create(filepath.Join(s.defautDir, filename))
	if err != nil {
		fmt.Println("create file error: ", err)
		s.ack(c, err)
		return
	}
	defer file.Close()
	// n, err := io.Copy(file, c)
	n, err := io.CopyN(file, c, int64(size))
	if err != nil {
		fmt.Println("copy file fron socket error: ", err)
		s.ack(c, err)
		return
	}
	fmt.Printf("[server] filename: %s, receive size: %d\n", filename, n)
	s.ack(c, nil)
}

type client struct {
	others []string
	fo     *FileOps
}

func (c *client) ack(conn net.Conn) error {
	ret := make([]byte, 128)
	times := 0

	for n, err := conn.Read(ret); err == nil && n == 0; n, err = conn.Read(ret) {
		time.Sleep(time.Second)
		times += 1
		if times == 10 {
			return errors.New("read time out")
		}
	}
	retMsg := string(ret)
	if strings.HasPrefix(retMsg, "fail") {
		return errors.New(retMsg)
	}
	return nil
}

func (c *client) transToOthers(fullFilepath string) error {
	name, size, err := c.fo.fileInfo(fullFilepath)
	if err != nil {
		return err
	}
	fmt.Printf("[client] filename: %s, size: %d\n", name, size)
	file, err := os.Open(fullFilepath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, s := range c.others {
		conn, err := net.Dial("tcp", s)
		if err != nil {
			return err
		}
		defer conn.Close()
		fmt.Fprintf(conn, "%s,%d\n", name, size)

		// err = binary.Write(conn, binary.LittleEndian, size)
		// if err != nil {
		//	 return err
		// }

		n, err := io.Copy(conn, file)
		if err != nil {
			return err
		}
		fmt.Printf("[client] filename: %s, trans size: %d\n", name, n)
		err = c.ack(conn)
		if err != nil {
			return err
		}
	}
	return nil
}
