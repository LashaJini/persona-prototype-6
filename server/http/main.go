package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	rpcclient "github.com/wholesome-ghoul/persona-prototype-6/client/rpc"
	"github.com/wholesome-ghoul/persona-prototype-6/constants"
	pb "github.com/wholesome-ghoul/persona-prototype-6/protos"
)

type apiHandler struct {
	rpc *rpcclient.Client
}

func (apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "API handler")
}

func (api *apiHandler) SemanticSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	text := r.FormValue("text")
	if text == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "text parameter is empty"}`)
		return
	}

	ctx := context.Background()
	search := &pb.Search{
		Text: text,
	}
	contentIDs, err := api.rpc.SemanticSearch(ctx, search)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, item := range contentIDs.Items {
		// TODO:
		fmt.Println(item)
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	api := &apiHandler{
		rpc: rpcclient.NewClient(constants.PYTHON_GRPC_SERVER_PORT),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/", api.ServeHTTP)
	mux.HandleFunc("POST /api/semantic-search", api.SemanticSearch)

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// The "/" pattern matches everything, so we need to check
		// that we're at the root here.
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		fmt.Fprintf(w, "Welcome to the home page!")
	})

	log.Fatal(http.ListenAndServe(":8000", mux))
}
