package main

import (
	"os"
	"fmt"
	"log"
	"time"
	"bytes"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/samuel/go-zookeeper/zk"
)
var serverName string = ""
func main() {

	serverName = os.Getenv("servername")
	servers := make([]string, 1)
        servers[0] = "zookeeper:2181"

	conn, _, err := zk.Connect(servers, time.Second)
	showErrorLog(err)
	for conn.State() != zk.StateHasSession {
		fmt.Printf("Zookeeper is loading .....\n")
		time.Sleep(30)
	}
	
	fmt.Printf(" %s is connected with Zookeeper\n", serverName)
	gserve, err := conn.Create("/grproxy/"+serverName, []byte(serverName+":9091"), int32(zk.FlagEphemeral), zk.WorldACL(zk.PermAll))
	showErrorLog(err)
	fmt.Printf("Ephemeral node: %+v\n", gserve)

	http.HandleFunc("/library", handler)
	log.Fatal(http.ListenAndServe(":9091", nil))
	conn.Close()
}

func showErrorLog(err error) {
	if err != nil {
		fmt.Println("Error Log: %+v\n", err)
	}
}

func handler(writer http.ResponseWriter, req *http.Request) {

	if req.Method == "POST" || req.Method == "PUT" {
		encodedJsonByte, err := ioutil.ReadAll(req.Body)
		showErrorLog(err)

		encodedJSONData := encoder(encodedJsonByte)
		fmt.Println("encodedJSON Data : ", string(encodedJSONData))

		req.Header.Set("Content-type", "application/json")
		requestUrl := "http://hbase:8080/se2:library/fakerow"
		postResponse, postError := http.Post(requestUrl, "application/json", bytes.NewBuffer([]byte(encodedJSONData)))

		if postError != nil {
			fmt.Println("Error from post response: %+v", postError)
			return
		}

		fmt.Println("Post response: ", postResponse.Status)
		defer postResponse.Body.Close()
		fmt.Fprintf(writer, "an %s\n", "POST")

	} else if req.Method == "GET" {
		req.Header.Set("Accept", "application/json")
		requestUrl := "http://hbase:8080/se2:library/*"

		request, _ := http.NewRequest("GET", requestUrl, nil)
		request.Header.Set("Accept", "application/json")
		client := &http.Client{}
		getResponse, getError := client.Do(request)
		showErrorLog(getError)

		fmt.Println("Get response: ", getResponse.Status)
		encodedJsonByte, err := ioutil.ReadAll(getResponse.Body)
		showErrorLog(err)

		decodedJSON := decoder(encodedJsonByte)
		defer getResponse.Body.Close()
		fmt.Fprintf(writer, "HBase response:\n\n %s\n", string(decodedJSON))

	} else {
		fmt.Fprintf(writer, "Invalid client request")
	}
	fmt.Fprintf(writer, "proudly served by %s", serverName)
}

func encoder(unencodedJSON []byte) string {
	var unencodedRows RowsType
	json.Unmarshal(unencodedJSON, &unencodedRows)
	encodedRows := unencodedRows.encode()
	encodedJSON, _ := json.Marshal(encodedRows)
	return string(encodedJSON)
}

func decoder(encodedJSON []byte) string {
	var encodedRows EncRowsType
	json.Unmarshal(encodedJSON, &encodedRows)
	decodedRows, err := encodedRows.decode()
	showErrorLog(err);
	deCodedJSON, _ := json.Marshal(decodedRows)
	return string(deCodedJSON)
}
