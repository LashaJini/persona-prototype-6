package crawler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/wholesome-ghoul/persona-prototype-6/god"
	"github.com/wholesome-ghoul/persona-prototype-6/reddit"
)

func fetchAbout(ctx context.Context, g *god.God, id int, username string) (reddit.UserAbout, error) {
	get := func(accessToken string) *http.Request {
		return buildGetRequest(ctx, accessToken, reddit.AboutURL(username))
	}

	<-g.CanMakeRequest
	g.Log.Info().Msgf("crawler #%d fetching about for username `%s`", id, username)
	res, err := g.Client.DoWithRefreshToken(get)
	if err != nil {
		return reddit.NewUser().About, err
	}

	var body reddit.UserAbout
	if err = json.NewDecoder(res.Body).Decode(&body); err != nil {
		return reddit.NewUser().About, err
	}
	res.Body.Close()

	return body, err
}
