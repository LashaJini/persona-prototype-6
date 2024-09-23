package crawler

import (
	"context"

	"github.com/wholesome-ghoul/persona-prototype-6/god"
	"github.com/wholesome-ghoul/persona-prototype-6/reddit"
)

func fetchComments(
	ctx context.Context,
	g *god.God,
	id int,
	username string,
	receiver, sender chan bool,
) Content[reddit.Comment] {
	return fetchContent[reddit.Comment](ctx, g, id, username, receiver, sender)
}
