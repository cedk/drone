package trypod

type User struct {
	ID       int64  `json:"id"`
	UserName string `json:"username"`
	RealName string `json:"realname"`
	Address  string `json:"address"`
	Token    string `json:"token"`
	Avatar   string `json:"avatar"`
}

type Repo struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Owner string `json:"owner"`
}

type Commit struct {
	Author      string `json:"author"`
	Avatar      string `json:"avatar"`
	Repository  string `json:"repository"`
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Rev         string `json:"rev"`
	Branch      string `json:"branch"`
	Description string `json:"description"`
}

type Status struct {
	URL        string `json:"url"`
	Repository string `json:"repository"`
	Status     string `json:"status"`
	Branch     string `json:"branch"`
	Rev        string `json:"rev"`
	Author     string `json:"author"`
	Email      string `json:"email"`
	Message    string `json:"message"`
	Event      string `json:"event"`
}
