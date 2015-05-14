package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type config struct {
	bind    string
	sh      string
	args    []string
	repodir string
	repos   map[string]string
	alias   map[string]string
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
	c.alias = make(map[string]string)

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
		case "repo":
			c.repodir = v
		case "repo.alias":
			for _, kvp := range strings.Split(v, ";") {
				kv := strings.Split(kvp, ":")
				if len(kv) == 2 && len(kv[0]) > 0 && len(kv[1]) > 0 {
					c.alias[kv[0]] = kv[1]
				}
			}
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

	defer conf.Close()
	return parseConfig(bufio.NewReader(conf))
}

func checkRepo(repo string) (string, error) {
	d, ok := g.repos[repo]
	if ok {
		return d, nil
	}

	if len(g.repodir) > 0 {
		p1 := filepath.Join(g.repodir, repo)
		fmt.Println(g.repodir, p1)
		fi, err := os.Stat(p1)
		if err == nil && fi.IsDir() {
			return p1, nil
		}

		p1 += ".git"
		fi, err = os.Stat(p1)
		if err == nil && fi.IsDir() {
			return p1, nil
		}
	}

	return "", fmt.Errorf("invalid repo: " + repo)
}

func checkAncestor(repo, now, eld string) string {
	d, err := checkRepo(repo)
	if err != nil {
		return fmt.Sprintf("\n%v", err)
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

type tRepo struct {
	Repo, Alias string
}

type tRepoSort []tRepo

type tAncestorIndex struct {
	Repos  []tRepo
	Result string
}

func dataAncestorIndex() tAncestorIndex {
	return dataAncestorResult("")
}

func dataAncestorResult(result string) tAncestorIndex {
	repos := make([]tRepo, 0, len(g.alias))
	for k, v := range g.alias {
		repos = append(repos, tRepo{k, v})
	}

	sort.Sort(tRepoSort(repos))

	info := ""
	r := strings.Split(result, "\n")
	if r[0] == "0" {
		info = "结果：不包含"
	} else if r[0] == "1" {
		info = "结果：包含"
	} else if len(r) > 1 {
		info = "错误：" + strings.Join(r[1:], "\n")
	}

	return tAncestorIndex{repos, info}
}

func (s tRepoSort) Len() int {
	return len(s)
}

func (s tRepoSort) Less(i, j int) bool {
	return s[i].Alias < s[j].Alias
}

func (s tRepoSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func ancestorServe(w http.ResponseWriter, req *http.Request) {
	if req.ParseForm() == nil {
		repo := req.Form.Get("repo")
		if len(repo) == 0 {
			alias := req.Form.Get("alias")
			if len(alias) > 0 {
				repo = alias
			}
		}

		if len(repo) == 0 {
			t, err := template.ParseFiles("ancestor.index.tmpl")
			if err == nil {
				err = t.Execute(w, dataAncestorIndex())
				if err == nil {
					return
				}
			}

			fmt.Fprintln(w, `<html><body>参数：?repo=...&amp;now=...&amp;eld=...</body></html>`, err)
			return
		}

		now := req.Form.Get("now")
		eld := req.Form.Get("eld")
		if !checkArg(repo) || !checkArg(now) || !checkArg(eld) {
			fmt.Fprintln(w, "invalid input")
		} else {
			result := checkAncestor(repo, now, eld)
			t, err := template.ParseFiles("ancestor.index.tmpl")
			if err == nil {
				err = t.Execute(w, dataAncestorResult(result))
				if err == nil {
					return
				}
			}

			fmt.Fprintln(w, `结果：`+result)
		}
	} else {
		fmt.Fprintln(w, "invalid input")
	}
}

func indexServe(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, `
<html><body><a href="./ancestor">查询提交包含关系</a></body></html>`)
}

func main() {
	g = readConfig()

	http.HandleFunc("/ancestor", ancestorServe)
	http.HandleFunc("/", indexServe)
	http.ListenAndServe(g.bind, nil)
}
