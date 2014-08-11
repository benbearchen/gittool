package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
    "strings"
)

var (
	bind  string            = ":8001"
	sh    string            = "sh"
	args  []string          = []string{"ancestor.sh"}
	repos map[string]string = make(map[string]string)
)

func readConfig() {
	sh = "d:\\software\\cygwin\\bin\\bash"
	args = []string{"--login", "/cygdrive/d/go/src/github.com/benbearchen/gittool/ancestor/ancestor.sh"}
	repos["intro"] = "/cygdrive/d/work/git/intro"
}

func checkAncestor(repo, now, eld string) string {
	d, ok := repos[repo]
	if !ok {
		return "\ninvalid repo: " + repo
	}

	arg := args[:]
	arg = append(arg, d, now, eld)
	b, err := exec.Command(sh, arg...).Output()
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
	readConfig()

	http.HandleFunc("/ancestor", ancestorServe)
	http.ListenAndServe(bind, nil)
}
