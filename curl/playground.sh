#!/bin/bash

curl -i -X POST \
	-H "Content-Type: application/x-www-form-urlencoded" \
	-d "text=person who loves cars" \
	localhost:8000/api/semantic-search
