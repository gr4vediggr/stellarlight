package packets

type Msg = isPacket_Payload

func NewChat(msg string) Msg {
	return &Packet_ChatMessage{
		ChatMessage: &ChatMessage{
			Content: msg,
		},
	}
}

func NewId(id uint64) Msg {
	return &Packet_IdMessage{
		IdMessage: &IdMessage{
			Id: id,
		},
	}
}
