package reddit

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/wholesome-ghoul/persona-prototype-6/constants"
)

type User struct {
	About                    UserAbout
	LatestScannedContentName string
}

func NewUser() *User {
	return &User{
		About: UserAbout{},
	}
}

func (u User) Value() (driver.Value, error) {
	return json.Marshal(u)
}

func (u *User) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &u)
}

func (u *User) IsGold() bool      { return u.About.IsGold() }
func (u *User) IsEmployee() bool  { return u.About.IsEmployee() }
func (u *User) IsSuspended() bool { return u.About.IsSuspended() }
func (u *User) IsVerified() bool  { return u.About.IsVerified() }
func (u *User) TotalKarma() int   { return u.About.TotalKarma() }

type UserAbout struct {
	Data struct {
		IsEmployee       bool    `json:"is_employee"`
		IsSuspended      bool    `json:"is_suspended"`
		IsVerified       bool    `json:"verified"`
		IsGold           bool    `json:"is_gold"`
		HasVerifiedEmail bool    `json:"has_verified_email"`
		LinkKarma        int     `json:"link_karma"`
		TotalKarma       int     `json:"total_karma"`
		CommentKarma     int     `json:"comment_karma"`
		CreatedUTC       float64 `json:"created_utc"`
		Subreddit        struct {
			NSFW bool `json:"over_18"`
		} `json:"subreddit"`
	} `json:"data"`
}

func (ua *UserAbout) IsGold() bool      { return ua.Data.IsGold }
func (ua *UserAbout) IsEmployee() bool  { return ua.Data.IsEmployee }
func (ua *UserAbout) IsSuspended() bool { return ua.Data.IsSuspended }
func (ua *UserAbout) IsVerified() bool  { return ua.Data.IsVerified }
func (ua *UserAbout) TotalKarma() int   { return ua.Data.TotalKarma }

type UserContentSchema[K PostOrComment] struct {
	Data struct {
		After    string `json:"after"`
		Children []K    `json:"children"`
	} `json:"data"`
}

type PostOrComment interface {
	Post | Comment
	CreatedAtEpoch() int64
	GetName() string
	Value() (driver.Value, error)
	ThirdpartyURL() string
	ContentTypeName() string
	Text() string

	APIUrl(username, after string, limit int) string
}

type PostData struct {
	Subreddit  string  `json:"subreddit"`
	Selftext   string  `json:"selftext"`
	Title      string  `json:"title"`
	Name       string  `json:"name"`
	Downs      int     `json:"down"`
	Ups        int     `json:"ups"`
	Score      int     `json:"score"`
	CreatedUTC float64 `json:"created_utc"`
	ID         string  `json:"id"`
	Permalink  string  `json:"permalink"`
}

func (p PostData) MarshalJSON() ([]byte, error) {
	type Alias PostData

	tmp := struct {
		*Alias
		Selftext string `json:"selftext,omitempty"`
		Title    string `json:"title,omitempty"`
	}{
		Alias:    (*Alias)(&p),
		Selftext: "",
		Title:    "",
	}

	return json.Marshal(&tmp)
}

type Post struct {
	Data PostData `json:"data"`
}

func (p Post) CreatedAtEpoch() int64        { return int64(p.Data.CreatedUTC) }
func (p Post) GetName() string              { return p.Data.Name }
func (p Post) ThirdpartyURL() string        { return fmt.Sprintf("https://reddit.com%s", p.Data.Permalink) }
func (p Post) ContentTypeName() string      { return constants.CONTENT_TYPE_POST }
func (p Post) Value() (driver.Value, error) { return json.Marshal(p) }
func (p Post) Text() string {
	if len(p.Data.Selftext) == 0 {
		return p.Data.Title
	}

	return p.Data.Selftext
}

func (p Post) APIUrl(username, after string, limit int) string {
	return postsURL(username, after, limit)
}

type ContentSlice[V PostOrComment] []V

func (p *ContentSlice[V]) Texts() []string {
	var texts []string
	for _, content := range *p {
		texts = append(texts, content.Text())
	}

	return texts
}

type CommentData struct {
	SubredditID string  `json:"subreddit_id"`
	LinkTitle   string  `json:"link_title"`
	Ups         int     `json:"ups"`
	Subreddit   string  `json:"subreddit"`
	LinkAuthor  string  `json:"link_author"`
	ID          string  `json:"id"`
	Score       int     `json:"score"`
	Body        string  `json:"body"`
	Name        string  `json:"name"`
	Downs       int     `json:"downs"`
	Permalink   string  `json:"permalink"`
	CreatedUTC  float64 `json:"created_utc"`
}

func (p CommentData) MarshalJSON() ([]byte, error) {
	type Alias CommentData

	tmp := struct {
		*Alias
		Body string `json:"body,omitempty"`
	}{
		Alias: (*Alias)(&p),
		Body:  "",
	}

	return json.Marshal(&tmp)
}

type Comment struct {
	Data CommentData `json:"data"`
}

func (c Comment) CreatedAtEpoch() int64        { return int64(c.Data.CreatedUTC) }
func (c Comment) GetName() string              { return c.Data.Name }
func (c Comment) Value() (driver.Value, error) { return json.Marshal(c) }
func (c Comment) ThirdpartyURL() string        { return fmt.Sprintf("https://reddit.com%s", c.Data.Permalink) }
func (c Comment) ContentTypeName() string      { return constants.CONTENT_TYPE_COMMENT }
func (c Comment) Text() string                 { return c.Data.Body }
func (c Comment) APIUrl(username, after string, limit int) string {
	return commentsURL(username, after, limit)
}

func GetLatestActivityEpoch[V PostOrComment](contents *ContentSlice[V]) int64 {
	var latestActivityEpoch int64
	for _, content := range *contents {
		createdAtEpoch := content.CreatedAtEpoch()
		if latestActivityEpoch < createdAtEpoch {
			latestActivityEpoch = createdAtEpoch
		}
	}

	return latestActivityEpoch
}

func GetOldestActivityEpoch[V PostOrComment](contents *ContentSlice[V]) int64 {
	oldestActivityEpoch := time.Now().UTC().Unix()
	for _, content := range *contents {
		createdAtEpoch := content.CreatedAtEpoch()
		if oldestActivityEpoch > createdAtEpoch {
			oldestActivityEpoch = createdAtEpoch
		}
	}

	return oldestActivityEpoch
}

func AboutURL(username string) string {
	return fmt.Sprintf("https://oauth.reddit.com/user/%s/about", username)
}

func commentsURL(username, after string, limit int) string {
	return fmt.Sprintf("https://oauth.reddit.com/user/%s/comments?limit=%d&sort=new&after=%s", username, limit, after)
}

func postsURL(username, after string, limit int) string {
	return fmt.Sprintf("https://oauth.reddit.com/user/%s/submitted?limit=%d&sort=new&after=%s", username, limit, after)
}
