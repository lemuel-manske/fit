package protocol

import "fit/fit/internal"

type RepositoryDiscoverPayload struct {
	RepositoryID   internal.RepositoryID `json:"repositoryId,omitempty"`
	RepositoryName string                `json:"repositoryName,omitempty"`
}

type RepositoryOfferPayload struct {
	RepositoryID   internal.RepositoryID `json:"repositoryId"`
	RepositoryName string                `json:"repositoryName"`
	PeerID         internal.PeerID       `json:"peerId"`
	Head           internal.CommitID     `json:"head"`
}

type HeadRequestPayload struct{}

type HeadAnnouncePayload struct {
	Head internal.CommitID `json:"head"`
}

type CommitRequestPayload struct {
	CommitID internal.CommitID `json:"commitId"`
}

type CommitResponsePayload struct {
	Commit internal.Commit `json:"commit"`
}

type BlobRequestPayload struct {
	Hash internal.Hash `json:"hash"`
}

type BlobResponsePayload struct {
	Hash internal.Hash `json:"hash"`
	Data []byte        `json:"data"`
}
