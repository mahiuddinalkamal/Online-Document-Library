package main

import (
	"fmt"
	"log"
	"time"
        "math/rand"
	"net/http"
	"net/http/httputil"
	"github.com/samuel/go-zookeeper/zk"
)

var urlList []string

func main() {

  servers := make([]string, 1)
  servers[0] = "zookeeper:2181"
  conn, _, err := zk.Connect(servers, time.Second)
  showErrorLog(err)
  defer conn.Close()

  for conn.State() != zk.StateHasSession {
    fmt.Printf("Zookeeper is loading .....\n")
    time.Sleep(5)
  }

  exists, stat, err := conn.Exists("/grproxy")
  showErrorLog(err)
  fmt.Printf("Exists: %+v %+v\n", exists, stat)

  if !exists {
	grproxy, err := conn.Create("/grproxy", []byte("grproxy:80"), int32(0), zk.WorldACL(zk.PermAll))
	showErrorLog(err)
	fmt.Printf("Create: %+v\n", grproxy)
  }

  childchn := make(chan []string)
  errors := make(chan error)
  go func() {  
    for {
	children, _, events, err := conn.ChildrenW("/grproxy")
	if err != nil {
		errors <- err
		return
	}
	childchn <- children
	evnt := <-events
	if evnt.Err != nil {
		errors <- evnt.Err
		return
	}
    }
  }()

  go func() {
    for {
	select {
	    case children := <-childchn:
		  fmt.Printf("%+v \n", children)
		  var temp []string
		  for _, child := range children {
			gserveUrlList, _, err := conn.Get("/grproxy/" + child)
			temp = append(temp, string(gserveUrlList))
			if err != nil {
				fmt.Printf("Errors from child: %+v\n", err)
			}
		}
		urlList = temp
		fmt.Printf("%+v \n", urlList)
	    case err := <-errors:
		fmt.Printf("Other errors: %+v\n", err)
	}
    }
  }()

  proxy := NewMultipleHostReverseProxy()
  log.Fatal(http.ListenAndServe(":8080", proxy))
}

func showErrorLog(err error) {
	if err != nil {
		fmt.Printf("Error Log: %+v\n", err)
	}
}

func NewMultipleHostReverseProxy() *httputil.ReverseProxy {

	director := func(req *http.Request) {

		if req.URL.Path == "/library" {
			fmt.Println("This is handled by gserver")
			hostName := urlList[rand.Int()%len(urlList)]
			req.URL.Scheme = "http"
			req.URL.Host = hostName

		} else {
			fmt.Println("This is handled by nginx")
			req.URL.Scheme = "http"
			req.URL.Host = "nginx"
		}

	}
	return &httputil.ReverseProxy{Director: director}
}
