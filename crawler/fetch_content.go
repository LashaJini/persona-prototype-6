package crawler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/wholesome-ghoul/persona-prototype-6/god"
	"github.com/wholesome-ghoul/persona-prototype-6/reddit"
)

func fetchContent[V reddit.PostOrComment](
	ctx context.Context,
	g *god.God,
	id int,
	username string,
	receiver, sender chan bool,
) Content[V] {
	g.Lock()
	latestScannedContentName := g.Get(username).Persona.Thirdparty.LatestScannedContentName
	latestActivityAt := g.Get(username).Persona.LatestActivityAt
	g.Unlock()

	after := ""
	content := Content[V]{}
	var dummy V

	var contentType string
	switch any(dummy).(type) {
	case reddit.Post:
		contentType = "posts"
	default:
		contentType = "comments"
	}

	for {
		get := func(accessToken string) *http.Request {
			return buildGetRequest(ctx, accessToken, dummy.APIUrl(username, after, limit))
		}

		<-g.CanMakeRequest
		g.Log.Info().Msgf("crawler #%d fetching %s for username `%s`", id, contentType, username)
		res, err := g.Client.DoWithRefreshToken(get)
		if err != nil {
			g.Log.Error().Stack().Err(err).Send()
			break
		}
		var body reddit.UserContentSchema[V]
		if err = json.NewDecoder(res.Body).Decode(&body); err != nil {
			g.Log.Error().Stack().Err(err).Send()
			break
		}
		res.Body.Close()

		fetchedData := body.Data.Children
		var newData []V
		filterNewContent(&fetchedData, &newData, latestScannedContentName, latestActivityAt)

		content.items = append(content.items, newData...)
		for range len(newData) {
			content.spamProbs = append(content.spamProbs, 0)
		}

		// first request
		if after == "" {
			var sentences []string
			for _, data := range newData {
				sentences = append(sentences, data.Text())
			}

			spam := freshContentIsSpam(ctx, g, sentences, receiver, sender)

			copy(content.spamProbs, spam.spamProbs)
			content.averageSpam = spam.averageSpam
			content.averageSpamCalculatedFor = len(spam.spamProbs)

			if spam.isSpam {
				break
			}
		}

		// INFO: reddit provides next/after content's name for us
		after = body.Data.After
		if after == "" {
			break
		}
	}

	return content
}
