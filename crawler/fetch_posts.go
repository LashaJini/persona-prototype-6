package crawler

import (
	"context"

	"github.com/wholesome-ghoul/persona-prototype-6/god"
	"github.com/wholesome-ghoul/persona-prototype-6/reddit"
)

func fetchPosts(
	ctx context.Context,
	g *god.God,
	id int,
	username string,
	receiver, sender chan bool,
) Content[reddit.Post] {
	return fetchContent[reddit.Post](ctx, g, id, username, receiver, sender)
}
