 
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strings"
	//"github.com/alecthomas/chroma/quick"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/", sourceCodeHandler)
	mux.HandleFunc("/panic/", panicDemo)
	mux.HandleFunc("/panic-after/", panicAfterDemo)
	mux.HandleFunc("/", hello)
	log.Fatal(http.ListenAndServe(":3022", devMw(mux)))
}

func sourceCodeHandler(w http.ResponseWriter, r *http.Request) {
	//path := "/home/hemant/Desktop/goprogram/RECOVER_CHROMA/main.go"
	
	path := r.FormValue("path")  // returns the first value for the named components of the query. 

	file, err := os.Open(path) // For read Access
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	/*b := bytes.NewBuffer(nil)
	_, err = io.Copy(b, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}*/

	//_ = quick.Highlight(w, b.String(), "go", "html", "monokai") //  used for highlighting syntax in browsers.

	io.Copy(w, file)
}

func devMw(app http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
				stack := debug.Stack()
				log.Println(string(stack))  // used to print the stacktrace for the current stack trace.
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "<h1>panic: %v</h1><pre>%s</pre>", err, makeLinks(string(stack)))
			}
		}()
		app.ServeHTTP(w, r)
	}
}

func panicDemo(w http.ResponseWriter, r *http.Request) {
	funcThatPanics()
}

func panicAfterDemo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello!</h1>")
	funcThatPanics()
}

func funcThatPanics() {
	panic("Oh no!")
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>Hello!</h1>")
}

/*
goroutine 31 [running]:
runtime/debug.Stack(0xc0000bf870, 0x1, 0x1)
	/usr/local/go/src/runtime/debug/stack.go:24 +0x9f
main.devMw.func1.1(0x6f7350, 0xc0000fca80)
	/home/hemant/Desktop/goprogram/RECOVER_CHROMA/main.go:48 +0xaa
panic(0x650bc0, 0x6ef3f0)
	/usr/local/go/src/runtime/panic.go:965 +0x1b9
main.funcThatPanics(...)
	/home/hemant/Desktop/goprogram/RECOVER_CHROMA/main.go:68
main.panicDemo(0x6f7350, 0xc0000fca80, 0xc000119300)
	/home/hemant/Desktop/goprogram/RECOVER_CHROMA/main.go:59 +0x39
net/http.HandlerFunc.ServeHTTP(0x6b7aa0, 0x6f7350, 0xc0000fca80, 0xc000119300)
	/usr/local/go/src/net/http/server.go:2069 +0x44
net/http.(*ServeMux).ServeHTTP(0xc0000da1c0, 0x6f7350, 0xc0000fca80, 0xc000119300)
	/usr/local/go/src/net/http/server.go:2448 +0x1ad
main.devMw.func1(0x6f7350, 0xc0000fca80, 0xc000119300)
	/home/hemant/Desktop/goprogram/RECOVER_CHROMA/main.go:54 +0x7e
net/http.HandlerFunc.ServeHTTP(0xc0000a80a8, 0x6f7350, 0xc0000fca80, 0xc000119300)
	/usr/local/go/src/net/http/server.go:2069 +0x44
net/http.serverHandler.ServeHTTP(0xc0000fc000, 0x6f7350, 0xc0000fca80, 0xc000119300)
	/usr/local/go/src/net/http/server.go:2887 +0xa3
net/http.(*conn).serve(0xc0000b8e60, 0x6f7760, 0xc0000da7c0)
	/usr/local/go/src/net/http/server.go:1952 +0x8cd
created by net/http.(*Server).Serve
	/usr/local/go/src/net/http/server.go:3013 +0x39b
*/

// makeLinks function actually used for make links or  we can say that parse the stack trace or create  links

func makeLinks(stack string) string {
	lines := strings.Split(stack, "\n") // split string when next "\n" occur
	for li, line := range lines {
		if len(line) == 0 || line[0] != '\t' {
			continue
		}
		file := ""
		for i, ch := range line {
			if ch == ':' {
				file = line[1:i]
				break
			}
		}
		//	/usr/local/go/src/runtime/debug/stack.go:24 +0x9f
		var lineStr strings.Builder
		for i := len(file) + 2; i < len(line); i++ {
			if line[i] < '0' || line[i] > '9' {
				break
			}
			lineStr.WriteByte(line[i])
		}

		v := url.Values{} // used to map a string key to a list of values.
		v.Set("path", file)
		lines[li] = "\t<a href=\"/debug/?" + v.Encode() + "\">" + file + ":" + lineStr.String() + "</a>" + line[len(file)+2+len(lineStr.String()):]
	}
	return strings.Join(lines, "\n")
}
