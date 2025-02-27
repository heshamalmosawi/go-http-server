package server

import (
	"errors"
	"fmt"
	"httpserver/modules"
	"httpserver/modules/request"
	"log"
	"math"
	"net"
	"os"
	"strconv"
)

type Server struct {
	host           string
	ConnectedPort  uint
	all_ports      []uint
	default_server modules.DefaultServerConfig
	routes         map[string]modules.RouteConfig
}

func New(config modules.Config) Server {
	return Server{
		host:           config.Server.Host,
		ConnectedPort:  math.MaxUint,
		all_ports:      config.Server.Ports,
		default_server: config.DefaultServer,
		routes:         config.Routes,
	}
}

func (s Server) min_port() uint {
	minPort := s.all_ports[0]
	for _, port := range s.all_ports {
		if port < minPort {
			minPort = port
		}
	}
	return minPort
}

func (s *Server) connectToAvailablePort() (net.Listener, error) {
	checkedAllPorts := false
	portIndex := 0

	for {
		address := fmt.Sprintf("%v:%v", s.host, s.all_ports[portIndex])
		if listener, err := net.Listen("tcp", address); err == nil {
			s.ConnectedPort = s.all_ports[portIndex]
			return listener, nil
		} else {
			log.Println(err)
			portIndex++

			var nextPort uint = 0
			if !checkedAllPorts && portIndex >= len(s.all_ports) {
				checkedAllPorts = true
				nextPort = s.min_port() + 1
			} else if portIndex >= len(s.all_ports) {
				nextPort = s.all_ports[len(s.all_ports)-1] + 1
			}

			if nextPort > 9999 {
				return nil, errors.New("no available ports to bind to")
			} else if nextPort != 0 {
				s.all_ports = append(s.all_ports, nextPort)
			}
		}
	}
}

func (s Server) Run() error {
	listener, err := s.connectToAvailablePort()
	if err != nil {
		return err
	}
	fmt.Printf("You can access the server at http://%s:%v/\n", s.host, s.ConnectedPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("error accepting connection: ", err)
			break
		}
		defer conn.Close()

		err = handleConnection(conn)
		if err != nil {
			log.Println("error handling connection: ", err)
			break
		}
	}
	return nil
}

func handleConnection(conn net.Conn) error {
	request, err := request.From(conn)
	if err != nil {
		return err
	}
	fmt.Printf("Request: %+v\n", request)

	statusLine := "HTTP/1.1 200 OK\r\n"
	content, err := os.ReadFile("static/index.html")
	if err != nil {
		return err
	}
	response := statusLine+"Content-Length: "+strconv.Itoa(len(content))+"\r\n\r\n"+string(content)
	conn.Write([]byte(response))
	return nil
}
