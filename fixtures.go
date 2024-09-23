package main

import (
	"github.com/wholesome-ghoul/persona-prototype-6/reddit"
)

func SuspendedRedditUserAbout() reddit.UserAbout {
	var about reddit.UserAbout
	about.Data.IsSuspended = true

	return about
}
