package petstore

import "net/http"

type Petstore struct{}

type CreateRequest struct{}
type CreateResponse struct{}

func (p *Petstore) Create(w http.ResponseWriter, r *http.Request) {
	_ = &CreateRequest{}
	_ = &CreateResponse{}
}

type UpdateRequest struct{}
type UpdateResponse struct{}

func (p *Petstore) Update(w http.ResponseWriter, r *http.Request) {
	_ = &UpdateRequest{}
	_ = &UpdateResponse{}
}

type DeleteRequest struct{}
type DeleteResponse struct{}

func (p *Petstore) Delete(w http.ResponseWriter, r *http.Request) {
	_ = &DeleteRequest{}
	_ = &DeleteResponse{}
}

type ListRequest struct{}
type ListResponse struct{}

func (p *Petstore) List(w http.ResponseWriter, r *http.Request) {
	_ = &ListRequest{}
	_ = &ListResponse{}
}
