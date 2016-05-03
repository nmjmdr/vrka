package raft

type Transport interface {
	RequestForVote(voteRequest) (voteResponse,error)
}


type mockTransport struct {
	vres voteResponse
	err error
}


func NewMockTransport(vres voteResponse) *mockTransport {
	m := new(mockTransport)
	m.vres = vres

	return m
}


func NewMockTransportError(err error) *mockTransport {
	m := new(mockTransport)
	m.err = err
	return m
}

func (m *mockTransport) RequestForVote(vreq voteRequest) (voteResponse,error) {
	return m.vres,m.err
}
