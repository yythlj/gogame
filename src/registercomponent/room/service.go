package room

import (
	"context"
	"fmt"
	"game/src/common"
)

type (
	// UserMessage represents a message that user sent
	UserMessage struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	// NewUser message will be received when new user join room
	NewUser struct {
		Content string `json:"content"`
	}

	// AllMembers contains all members uid
	AllMembers struct {
		Members []string `json:"members"`
	}

	// JoinResponse represents the result of joining room
	JoinResponse struct {
		Code   int    `json:"code"`
		Result string `json:"result"`
	}
)

// Join room
func (r *Room) Join(ctx context.Context, msg []byte) (*JoinResponse, error) {
	s := common.GetDefaultApp().GetSessionFromCtx(ctx)
	if s.UID() == "" {
		return &JoinResponse{Result: "nologin"}, nil
	}
	if InGroup(ctx, "room") {
		return &JoinResponse{Result: "join fail: repeat join"}, nil
	}
	uids, _ := common.GetDefaultApp().GroupMembers(ctx, "room")
	s.Push("room.onMembers", &AllMembers{Members: uids})
	// notify others
	common.GetDefaultApp().GroupBroadcast(ctx, "game", "room", "room.onNewUser", &NewUser{Content: fmt.Sprintf("New user: %s", s.UID())})
	// new user join group
	common.GetDefaultApp().GroupAddMember(ctx, "room", s.UID()) // add session to group

	// on session close, remove it from group
	s.OnClose(func() {
		common.GetDefaultApp().GroupRemoveMember(ctx, "room", s.UID())
	})
	return &JoinResponse{Result: "success"}, nil
}

// Message sync last message to all members
func (r *Room) Message(ctx context.Context, msg *UserMessage) {
	if !InGroup(ctx, "room") {
		return
	}
	common.GetDefaultApp().GroupBroadcast(ctx, "game", "room", "onMessage", msg)
}
