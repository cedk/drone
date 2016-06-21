package trypod

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"gopkg.in/yaml.v1"

	"github.com/drone/drone/plugin/notify"
	"github.com/drone/drone/plugin/notify/webhook"
	"github.com/drone/drone/shared/build/script"
	"github.com/drone/drone/shared/model"
)

type Trypod struct {
	url   string
	owner string
	Open  bool
	token string
}

type User struct {
	Id       int64  `json:"id"`
	UserName string `json:"username"`
	RealName string `json:"realname"`
	Address  string `json:"address"`
}

type Repo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Commit struct {
	Author      string `json:"author"`
	Name        string `json:"name"`
	Repository  string `json:"repository"`
	Rev         string `json:"rev"`
	Branch      string `json:"branch"`
	Description string `json:"description"`
}

func New(url string, owner string, open bool, token string) *Trypod {
	return &Trypod{
		url:   url,
		owner: owner,
		Open:  open,
		token: token,
	}
}

func (r *Trypod) Authorize(res http.ResponseWriter, req *http.Request) (*model.Login, error) {
	username := req.FormValue("username")
	password := req.FormValue("password")

	resp, err := http.PostForm(r.url+"/login",
		url.Values{"username": {username}, "password": {password}})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var reply User
	err = json.Unmarshal(body, &reply)
	if err != nil {
		return nil, err
	}

	var login = new(model.Login)
	login.Login = reply.UserName
	login.Name = reply.RealName
	login.Email = reply.Address
	return login, nil
}

func (r *Trypod) GetKind() string {
	return model.RemoteTrypod
}

func (r *Trypod) GetHost() string {
	uri, _ := url.Parse(r.url)
	return uri.Host
}

func (r *Trypod) GetRepos(user *model.User) ([]*model.Repo, error) {
	var repos []*model.Repo

	resp, err := http.Get(r.url + "/repos")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var list []Repo
	err = json.Unmarshal(body, &list)
	if err != nil {
		return nil, err
	}

	var remote = r.GetKind()
	var hostname = r.GetHost()

	for _, item := range list {
		var repo = model.Repo{
			UserID:   user.ID,
			Remote:   remote,
			Host:     hostname,
			Owner:    r.owner,
			Name:     item.Name,
			Scm:      model.Mercurial,
			Private:  false,
			CloneURL: item.URL,
			GitURL:   item.URL,
			SSHURL:   item.URL,
			URL:      item.URL,
			Role:     &model.Perm{},
		}
		// Everybody has full access
		repo.Role.Admin = true
		repo.Role.Write = true
		repo.Role.Read = true

		repos = append(repos, &repo)
	}
	return repos, nil
}

func (r *Trypod) GetScript(user *model.User, repo *model.Repo, hook *model.Hook) ([]byte, error) {
	url := fmt.Sprintf("%s/raw-file/%s/.drone.yml", repo.URL, hook.Sha)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("http status not OK")
	}
	defer resp.Body.Close()
	var yml []byte
	yml, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	script, err := script.ParseBuild(string(yml))
	if err != nil {
		return nil, err
	}
	if script.Notifications == nil {
		script.Notifications = &notify.Notification{}
	}
	if script.Notifications.Webhook == nil {
		script.Notifications.Webhook = &webhook.Webhook{}
	}
	webhook := script.Notifications.Webhook
	webhook.URL = []string{r.url + "/notify/" + r.token}
	webhook.Success = new(bool)
	*webhook.Success = true
	webhook.Failure = new(bool)
	*webhook.Failure = true
	yml, err = yaml.Marshal(script)
	if err != nil {
		return nil, err
	}
	return yml, nil
}

func (r *Trypod) Activate(user *model.User, repo *model.Repo, link string) error {
	// TODO
	return nil
}

func (r *Trypod) ParseHook(req *http.Request) (*model.Hook, error) {
	defer req.Body.Close()
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var commit Commit
	err = json.Unmarshal(payload, &commit)
	if err != nil {
		return nil, err
	}

	return &model.Hook{
		Owner:     r.owner,
		Repo:      commit.Name,
		Sha:       commit.Rev,
		Branch:    commit.Branch,
		Author:    commit.Author,
		Timestamp: time.Now().UTC().String(),
		Message:   commit.Description,
	}, nil
}

func (r *Trypod) OpenRegistration() bool {
	return r.Open
}

func (r *Trypod) GetToken(user *model.User) (*model.Token, error) {
	return nil, nil
}
