package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type config struct {
	bind  string
	sh    string
	args  []string
	repos map[string]string
}

var (
	g *config
)

func parseConfig(reader *bufio.Reader) *config {
	c := new(config)
	c.bind = ":8001"
	c.sh = "sh"
	c.args = make([]string, 0)
	c.repos = make(map[string]string)

	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			break
		} else if len(s) == 0 || s[0] == '#' {
			continue
		}

		e := strings.Index(s, "=")
		if e <= 0 {
			continue
		}

		k := strings.TrimSpace(s[:e])
		v := strings.TrimSpace(s[e+1:])
		if len(v) == 0 {
			continue
		}

		switch k {
		case "bind":
			c.bind = v
		case "sh":
			c.sh = v
		case "arg":
			c.args = append(c.args, v)
		default:
			repo := "repo."
			if strings.Index(k, repo) == 0 {
				c.repos[k[len(repo):]] = v
			}
		}
	}

	return c
}

func readSampleConfig() *config {
	conf := `
bind=:8001
sh=d:\\software\\cygwin\\bin\\bash
arg=--login
arg=/cygdrive/d/go/src/github.com/benbearchen/gittool/ancestor/ancestor.sh
repo.intro=/cygdrive/d/work/git/intro
`

	return parseConfig(bufio.NewReader(bytes.NewBuffer([]byte(conf))))
}

func readConfig() *config {
	conf, err := os.Open("config.ini")
	if err != nil {
		return nil
	}

	return parseConfig(bufio.NewReader(conf))
}

func checkAncestor(repo, now, eld string) string {
	d, ok := g.repos[repo]
	if !ok {
		return "\ninvalid repo: " + repo
	}

	arg := g.args[:]
	arg = append(arg, d, now, eld)
	b, err := exec.Command(g.sh, arg...).Output()
	if err != nil {
		return string(b) + "\n" + err.Error()
	} else {
		return string(b) + "\n"
	}
}

func isAncestor(repo, now, eld string) (bool, error) {
	result := checkAncestor(repo, now, eld)
	if len(result) >= 2 {
		if result[:2] == "1\n" {
			return true, nil
		} else if result[:2] == "0\n" {
			return false, nil
		}
	}

	return false, fmt.Errorf("error: " + result)
}

func checkArg(arg string) bool {
	if len(arg) == 0 {
		return false
	}

	// Make sure that arg won't be `rm -rf /`...
	if strings.ContainsAny(arg, "|&\"'` \t\r\n") {
		return false
	}

	return true
}

func ancestorServe(w http.ResponseWriter, req *http.Request) {
	if req.ParseForm() == nil {
		repo := req.Form.Get("repo")
		now := req.Form.Get("now")
		eld := req.Form.Get("eld")
		if !checkArg(repo) || !checkArg(now) || !checkArg(eld) {
			io.WriteString(w, "invalid input")
		} else {
			io.WriteString(w, checkAncestor(repo, now, eld))
		}
	} else {
		io.WriteString(w, "invalid input")
	}
}

func main() {
	g = readConfig()

	http.HandleFunc("/ancestor", ancestorServe)
	http.ListenAndServe(g.bind, nil)
}
