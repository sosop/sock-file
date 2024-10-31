package sockfile

import "net"

type server struct {
	port     string
	poolSize uint8
}

func (s *server) listen() error {
	l, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go s.handle(conn)
	}
}

func (s *server) handle(c net.Conn) {
	defer c.Close()
}

type client struct {
}
