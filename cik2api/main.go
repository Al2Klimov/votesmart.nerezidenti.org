package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type httpLogger struct {
	next http.RoundTripper
}

var _ http.RoundTripper = httpLogger{}

func (hl httpLogger) RoundTrip(request *http.Request) (*http.Response, error) {
	fmt.Fprintf(os.Stderr, "%s %s\n", request.Method, request.URL.String())
	return hl.next.RoundTrip(request)
}

type closableReader struct {
	r io.Reader
}

var _ io.ReadCloser = closableReader{}

func (cr closableReader) Read(p []byte) (int, error) {
	return cr.r.Read(p)
}

func (cr closableReader) Close() error {
	return nil
}

var pollingStation = regexp.MustCompile(`(?m)\s*\(.*?\)\s*\z`)

func main() {
	cikCsv := flag.String("data", "", "FILE")
	uRL := flag.String("url", "", "URL")
	user := flag.String("user", "", "USERNAME")
	force := flag.Bool("force", false, "")
	flag.Parse()

	if strings.TrimSpace(*cikCsv) == "" {
		fmt.Fprintln(os.Stderr, "-data missing")
		os.Exit(2)
	}

	if strings.TrimSpace(*uRL) == "" {
		fmt.Fprintln(os.Stderr, "-url missing")
		os.Exit(2)
	}

	if strings.TrimSpace(*user) == "" {
		fmt.Fprintln(os.Stderr, "-user missing")
		os.Exit(2)
	}

	pass := os.Getenv("PASSWORD")
	if strings.TrimSpace(pass) == "" {
		fmt.Fprintln(os.Stderr, "$PASSWORD missing")
		os.Exit(2)
	}

	baseUrl, errPU := url.Parse(*uRL)
	if errPU != nil {
		fmt.Fprintln(os.Stderr, errPU.Error())
		os.Exit(2)
	}

	data, errOp := os.Open(*cikCsv)
	if errOp != nil {
		fmt.Fprintln(os.Stderr, errOp.Error())
		os.Exit(1)
	}

	reader := csv.NewReader(bufio.NewReader(data))
	states := map[string]map[string]struct{}{}

	for {
		row, errRd := reader.Read()
		if errRd != nil {
			if errRd == io.EOF {
				break
			}

			fmt.Fprintln(os.Stderr, errRd.Error())
			os.Exit(1)
		}

		if len(row) > 3 {
			state := strings.TrimSpace(row[3])
			offices, ok := states[state]

			if !ok {
				offices = map[string]struct{}{}
				states[state] = offices
			}

			offices[strings.TrimSpace(pollingStation.ReplaceAllLiteralString(row[4], ""))] = struct{}{}
		}
	}

	_ = data.Close()

	if !*force {
		fmt.Fprintf(os.Stderr, "Would have created %d states\n\n", len(states))

		buf := bufio.NewWriter(os.Stdout)

		for state, offices := range states {
			buf.Write([]byte("- state: "))
			json.NewEncoder(buf).Encode(state)
			buf.Write([]byte("  offices:\n"))

			for office := range offices {
				buf.Write([]byte("  - "))
				json.NewEncoder(buf).Encode(office)
			}
		}

		buf.Flush()
		return
	}

	client := http.Client{Transport: httpLogger{http.DefaultTransport}}
	req := http.Request{Method: "PUT", URL: baseUrl, Header: http.Header{}}

	req.SetBasicAuth(*user, pass)

	for state, offices := range states {
		buf := &bytes.Buffer{}

		{
			errEc := json.NewEncoder(buf).Encode(struct {
				RuName string `json:"ru_name"`
			}{state})
			if errEc != nil {
				fmt.Fprintln(os.Stderr, errEc.Error())
				os.Exit(1)
			}
		}

		baseUrl.Path = "/v1/states"
		req.Body = closableReader{buf}

		resp, errDR := client.Do(&req)
		if errDR != nil {
			fmt.Fprintln(os.Stderr, errDR.Error())
			os.Exit(1)
		}

		if resp.StatusCode != 201 {
			fmt.Fprintf(os.Stderr, "HTTP %d\n", resp.StatusCode)
			os.Exit(1)
		}

		{
			var rb struct {
				Id uuid.UUID `json:"id"`
			}

			if errDc := json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&rb); errDc != nil {
				fmt.Fprintln(os.Stderr, errDc.Error())
				os.Exit(1)
			}

			baseUrl.Path = "/v1/states/" + rb.Id.String() + "/offices"
		}

		resp.Body.Close()

		for office := range offices {
			buf := &bytes.Buffer{}

			{
				errEc := json.NewEncoder(buf).Encode(struct {
					RuName string `json:"ru_name"`
				}{office})
				if errEc != nil {
					fmt.Fprintln(os.Stderr, errEc.Error())
					os.Exit(1)
				}
			}

			req.Body = closableReader{buf}

			resp, errDR := client.Do(&req)
			if errDR != nil {
				fmt.Fprintln(os.Stderr, errDR.Error())
				os.Exit(1)
			}

			if resp.StatusCode != 204 {
				fmt.Fprintf(os.Stderr, "HTTP %d\n", resp.StatusCode)
				os.Exit(1)
			}

			resp.Body.Close()
		}
	}
}
