package packet

import (
	"github.com/anviod/ethernet-ip/bufferx"
	"github.com/anviod/ethernet-ip/types"
	"github.com/anviod/ethernet-ip/utils"
)

type MessageRouterRequest struct {
	Service         types.USInt
	RequestPathSize types.USInt
	RequestPath     []byte
	RequestData     []byte
}

func (m *MessageRouterRequest) Encode() []byte {
	if m.RequestPathSize == 0 {
		m.RequestPathSize = utils.Len(m.RequestPath)
	}

	io := bufferx.New(nil)
	io.WL(m.Service)
	io.WL(m.RequestPathSize)
	io.WL(m.RequestPath)
	io.WL(m.RequestData)

	return io.Bytes()
}

func NewMessageRouter(service types.USInt, path []byte, data []byte) *MessageRouterRequest {
	return &MessageRouterRequest{
		Service:         service,
		RequestPathSize: utils.Len(path),
		RequestPath:     path,
		RequestData:     data,
	}
}

type MessageRouterResponse struct {
	ReplyService           types.USInt
	Reserved               types.USInt
	GeneralStatus          types.USInt
	SizeOfAdditionalStatus types.USInt
	AdditionalStatus       []byte
	ResponseData           []byte
}

func (m *MessageRouterResponse) Decode(data []byte) {
	io := bufferx.NewReader(data)
	io.RL(&m.ReplyService)
	io.RL(&m.Reserved)
	io.RL(&m.GeneralStatus)
	io.RL(&m.SizeOfAdditionalStatus)
	m.AdditionalStatus = make([]byte, m.SizeOfAdditionalStatus*2)
	io.RL(&m.AdditionalStatus)
	m.ResponseData = io.ReadBytes(io.Len())
}
