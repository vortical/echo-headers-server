package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis"
)

func NewRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		//Addr:     "redist:6379",
		Addr: "redis:6379",
		//  Password: "", // no password set
		DB: 0, // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	return client
}

type Metrics struct {
	Success, Errors, Requests int
	Health                    float32 `json:"health,string"`
}

func (m *Metrics) AddError() {
	m.Errors++
	m.Requests++
	m.Health = m.calcHealth()
}

func (m *Metrics) AddSuccess() {
	m.Success++
	m.Requests++
	m.Health = m.calcHealth()
}

func (m *Metrics) calcHealth() float32 {
	if m.Requests == 0 {
		return 1.0
	}
	return 1.0 - float32(m.Errors)/float32(m.Requests)
}

func main() {

	metrics := &Metrics{}
	metrics.Health = metrics.calcHealth()

	redisClient := NewRedisClient()

	port := flag.Int("port", 8001, "port to run on (defaults to 8001)")
	flag.Parse()

	url := fmt.Sprintf("0.0.0.0:%d", *port)
	fmt.Printf("starting %s ....\n", url)

	http.HandleFunc("/crash", func(writer http.ResponseWriter, req *http.Request) {
		if err := redisClient.Incr("crash").Err(); err != nil {
			fmt.Fprintf(writer, "Could not increment crash count: %q\n", err.Error())
		}
		fmt.Fprintf(writer, "Crashing in 3s")

		go func() {
			time.Sleep(3 * time.Second)
			os.Exit(1)
		}()
	})

	http.HandleFunc("/error", func(writer http.ResponseWriter, req *http.Request) {
		metrics.AddError()
		fmt.Println("S,E,R", metrics.Success, metrics.Errors, metrics.Requests, metrics.Health)
		writer.WriteHeader(http.StatusInternalServerError)
	})

	http.HandleFunc("/success", func(writer http.ResponseWriter, req *http.Request) {
		metrics.AddSuccess()
		fmt.Println("S,E,R", metrics.Success, metrics.Errors, metrics.Requests, metrics.Health)
		writer.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/metrics", func(writer http.ResponseWriter, req *http.Request) {
		writer.WriteHeader(http.StatusOK)
		body, _ := json.Marshal(metrics)
		fmt.Fprintf(writer, string(body))
	})

	http.HandleFunc("/metrics/health", func(writer http.ResponseWriter, req *http.Request) {

		if metrics.Requests > 0 && metrics.Health < 0.75 {
			writer.WriteHeader(http.StatusServiceUnavailable)
			// if this happens, we have no way to recover
			fmt.Printf("Health call failed with result %f\n", metrics.Health)
			go func() {
				for {
					time.Sleep(1 * time.Second)
					// heal it back to threshold slowly
					metrics.AddSuccess()
					if metrics.Health >= 0.75 {
						return
					}
				}
			}()
		} else {
			fmt.Printf("Health call succeeded with result %f\n", metrics.Health)
			writer.WriteHeader(http.StatusOK)
		}
		body, _ := json.Marshal(metrics.Health)
		fmt.Fprintf(writer, string(body))
	})

	http.HandleFunc("/headers", func(writer http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(writer, "Method[%s] Url[%s] Protocol[%s]\n", req.Method, req.URL.String(), req.Proto)

		fmt.Fprint(writer, "v4\n")

		for key, value := range req.Header {
			fmt.Fprintf(writer, "Header[%q] = %q\n", key, value)
		}
		fmt.Fprintf(writer, "Host[%s]\n", req.Host)
		fmt.Fprintf(writer, "RemoteAddr[%s]\n", req.RemoteAddr)

		if err := req.ParseForm(); err != nil {
			log.Print(err)
		}

		for key, value := range req.Form {
			fmt.Fprintf(writer, "Form[%q] = %q\n", key, value)
		}

		fmt.Fprintf(writer, "URL path: %q\n", req.URL.Path)

		hostname, _ := os.Hostname()
		fmt.Fprintf(writer, "Hostname: %q\n", hostname)

		cmd := redisClient.Incr("key")
		if cmd.Err() != nil {
			fmt.Fprintf(writer, "Could not incr redis value : %q\n", cmd.Err())
			if err := redisClient.Set("key", 1, 0); err != nil {
				fmt.Fprintf(writer, "Could not set initial redis value : %q\n", err)
				return
			}
		}

		fmt.Fprintf(writer, "NbRequests: %d\n", cmd.Val())

		sCmd := redisClient.Get("crash")
		if sCmd.Err() != nil {
			fmt.Fprintf(writer, "Could not ready any crash counters: %q\n", sCmd.Err().Error())
			return
		}

		fmt.Fprintf(writer, "Crashes: : %s\n", sCmd.String())

	})

	log.Fatal(http.ListenAndServe(url, nil))
}
