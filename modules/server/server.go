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
	Default_server modules.DefaultServerConfig
	Routes         map[string]modules.RouteConfig
}

func New(config modules.Config) Server {
	return Server{
		host:           config.Server.Host,
		ConnectedPort:  math.MaxUint,
		all_ports:      config.Server.Ports,
		Default_server: config.DefaultServer,
		Routes:         config.Routes,
	}
}

// This method gets the port with the lowest value digits. For instance, if the server has two ports of [9859, 5065], this function returns 5065.
func (s Server) min_port() uint {
	minPort := s.all_ports[0]
	for _, port := range s.all_ports {
		if port < minPort {
			minPort = port
		}
	}
	return minPort
}

// Dynamically connects to an available port. Start with priority of order of ports specified in the configuration file, if all specified ports the
func (s *Server) connectToAvailablePort() (net.Listener, error) {
	// if no port specified, default to 5000
	if len(s.all_ports) == 0 {
		s.all_ports = append(s.all_ports, 5000)
	}

	checkedAllPorts := false
	var outOfPlacePort uint = 0

	for {
		// checking all ports specified in server config
		for _, port := range s.all_ports {
			address := fmt.Sprintf("%v:%v", s.host, port)
			if listener, err := net.Listen("tcp", address); err == nil {
				s.ConnectedPort = port
				return listener, nil
			} else {
				log.Println(err)
			}
		}

		if !checkedAllPorts {
			checkedAllPorts = true
			outOfPlacePort = s.min_port() + 1
		} else {
			outOfPlacePort++
		}

		if outOfPlacePort > 9999 {
			return nil, errors.New("no available ports to bind to")
		}

		// trying to bind to the next available dynamic port
		address := fmt.Sprintf("%v:%v", s.host, outOfPlacePort)
		if listener, err := net.Listen("tcp", address); err == nil {
			s.ConnectedPort = outOfPlacePort
			return listener, nil
		} else {
			log.Println(err)
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

		err = s.handleConnection(conn)
		if err != nil {
			log.Println("error handling connection: ", err)
			break
		}
	}
	return nil
}

func (s Server) handleConnection(conn net.Conn) error {
	request, err := request.From(conn)
	if err != nil {
		return err
	}
	fmt.Printf("Request: %+v\n", request)

	// loading up the 404 page, just in case.
	path404 := ""
	if p, ok := s.Default_server.ErrorPages["404"]; ok {
		path404 = p
	} else {
		path404 = "static/404.html"
	}

	filename := ""
	if routeconf, ok := s.Routes[request.Path]; ok {
		filename = routeconf.DefaultFile
	} else {
		filename = path404
	}

	statusLine := "HTTP/1.1 200 OK\r\n"
	filename = fmt.Sprintf("static/%s", filename)
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	response := statusLine + "Content-Length: " + strconv.Itoa(len(content)) + "\r\n\r\n" + string(content)
	conn.Write([]byte(response))
	return nil
}
