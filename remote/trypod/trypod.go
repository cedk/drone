package trypod

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/drone/drone/model"
	"github.com/drone/drone/remote"
)

type Opts struct {
	URL   string
	Token string
}

func New(opts Opts) (remote.Remote, error) {
	return &Trypod{
		URL:   opts.URL,
		Token: opts.Token,
	}, nil
}

type Trypod struct {
	URL   string
	Token string
}

func (t *Trypod) Login(res http.ResponseWriter, req *http.Request) (*model.User, error) {
	var (
		username = req.FormValue("username")
		password = req.FormValue("password")
	)

	// If the username of password is empty we re-direct to the login screen.
	if len(username) == 0 || len(password) == 0 {
		http.Redirect(res, req, "/login/form", http.StatusSeeOther)
		return nil, nil
	}

	resp, err := http.PostForm(t.URL+"/login",
		url.Values{"username": {username}, "password": {password}})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
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

	return &model.User{
		Login: reply.UserName,
		Token: reply.Token,
		Email: reply.Address,
	}, nil
}

func (t *Trypod) Auth(token, secret string) (string, error) {
	return "", fmt.Errorf("Not Implemented")
}

func (t *Trypod) Teams(u *model.User) ([]*model.Team, error) {
	return nil, nil
}

func (t *Trypod) TeamPerm(u *model.User, org string) (*model.Perm, error) {
	return nil, nil
}

func (t *Trypod) Repo(u *model.User, owner, name string) (*model.Repo, error) {
	resp, err := http.Get(t.URL + "/repo/" + name)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var repo = &model.Repo{}
	var reply Repo
	err = json.Unmarshal(body, &reply)
	if err != nil {
		return nil, err
	}
	repo.Owner = reply.Owner
	repo.Name = reply.Name
	repo.FullName = reply.Owner + "/" + reply.Name
	repo.Link = reply.URL
	repo.Clone = reply.URL
	repo.Branch = "default"
	return repo, nil
}

func (t *Trypod) Repos(u *model.User) ([]*model.Repo, error) {
	resp, err := http.Get(t.URL + "/repos")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var repos = []*model.Repo{}
	var reply []Repo
	err = json.Unmarshal(body, &reply)
	if err != nil {
		return nil, err
	}

	for _, repo := range reply {
		repos = append(repos, &model.Repo{
			Owner:    repo.Owner,
			Name:     repo.Name,
			FullName: repo.Owner + "/" + repo.Name,
			Link:     repo.URL,
			Clone:    repo.URL,
			Branch:   "default",
		})
	}
	return repos, nil
}

func (t *Trypod) Perm(u *model.User, owner, name string) (*model.Perm, error) {
	var p = &model.Perm{
		Admin: true,
		Pull:  true,
		Push:  true,
	}
	return p, nil
}

func (t *Trypod) File(u *model.User, repo *model.Repo, build *model.Build, f string) ([]byte, error) {
	url := fmt.Sprintf(t.URL+"/repo/%s/raw-file/%s/%s", repo.Name, build.Commit, f)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (t *Trypod) FileRef(u *model.User, repo *model.Repo, ref, f string) ([]byte, error) {
	url := fmt.Sprintf(t.URL+"/repo/%s/raw-file/%s/%s", repo.Name, ref, f)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (t *Trypod) Status(u *model.User, repo *model.Repo, b *model.Build, link string) error {
	url := fmt.Sprintf(t.URL+"/status/%s", t.Token)
	status := Status{
		URL:        link,
		Repository: repo.Name,
		Status:     b.Status,
		Branch:     b.Branch,
		Rev:        b.Commit,
		Author:     b.Author,
		Email:      b.Email,
		Message:    b.Message,
		Event:      b.Event,
	}
	payload, err := json.Marshal(status)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(payload)
	resp, err := http.Post(url, "application/json", buf)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}
	return nil
}

func (t *Trypod) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
	return &model.Netrc{}, nil
}

func (t *Trypod) Activate(u *model.User, r *model.Repo, link string) error {
	url := fmt.Sprintf(t.URL+"/repo/%s/activate/%s", r.Name, t.Token)
	buf := bytes.NewBufferString(link)
	resp, err := http.Post(url, "text/plain", buf)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}
	return nil
}

func (t *Trypod) Deactivate(u *model.User, r *model.Repo, link string) error {
	url := fmt.Sprintf(t.URL+"/repo/%s/deactivate/%s", r.Name, t.Token)
	resp, err := http.Post(url, "text/plain", nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}
	return nil
}

func (t *Trypod) Hook(r *http.Request) (*model.Repo, *model.Build, error) {
	defer r.Body.Close()
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}

	var commit Commit
	err = json.Unmarshal(payload, &commit)
	if err != nil {
		return nil, nil, err
	}

	var repo = &model.Repo{}
	repo.Owner = commit.Owner
	repo.Name = commit.Name
	repo.FullName = commit.Owner + "/" + commit.Name
	repo.Link = commit.Repository
	repo.Clone = commit.Repository
	repo.Branch = commit.Branch

	var build = &model.Build{}
	build.Event = model.EventPush
	build.Commit = commit.Rev
	build.Branch = commit.Branch
	build.Ref = commit.Rev
	build.Message = commit.Description
	build.Author = commit.Author
	build.Email = commit.Author
	return repo, build, nil
}
