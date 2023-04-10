package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nuid"
	"github.com/opentracing/opentracing-go"
	"github.com/topfreegames/pitaya/v2/acceptor"
	"github.com/topfreegames/pitaya/v2/agent"
	"github.com/topfreegames/pitaya/v2/conn/message"
	"github.com/topfreegames/pitaya/v2/conn/packet"
	"github.com/topfreegames/pitaya/v2/constants"
	pcontext "github.com/topfreegames/pitaya/v2/context"
	e "github.com/topfreegames/pitaya/v2/errors"
	"github.com/topfreegames/pitaya/v2/logger"
	"github.com/topfreegames/pitaya/v2/metrics"
	"github.com/topfreegames/pitaya/v2/route"
	"github.com/topfreegames/pitaya/v2/session"
	"github.com/topfreegames/pitaya/v2/tracing"
	"strings"
	"time"
)

type CustomMsgCtrl interface {
	Handle(msg *UnHandleMsg)
	Dispatch(concurrency int)
}

var MsgCtrlInstance CustomMsgCtrl = nil

type UnHandleMsg struct {
	ctx   context.Context
	agent agent.Agent
	route *route.Route
	msg   *message.Message
	typ   string
}

func (h *HandlerService) DispatchCustom(concurrency int) {
	close(h.chLocalProcess)
	close(h.chRemoteProcess)
	MsgCtrlInstance.Dispatch(concurrency)
}

func (h *HandlerService) ProcessLocalMsg(lmp *UnHandleMsg) {
	metrics.ReportMessageProcessDelayFromCtx(lmp.Ctx(), h.metricsReporters, "local")
	h.localProcess(lmp.Ctx(), lmp.Agent(), lmp.Router(), lmp.Msg())
}

func (h *HandlerService) ProcessRemoteMsg(rmp *UnHandleMsg) {
	metrics.ReportMessageProcessDelayFromCtx(rmp.Ctx(), h.metricsReporters, "remote")
	h.remoteService.remoteProcess(rmp.Ctx(), nil, rmp.Agent(), rmp.Router(), rmp.Msg())
}

func (u *UnHandleMsg) Agent() agent.Agent {
	return u.agent
}

func (u *UnHandleMsg) Msg() *message.Message {
	return u.msg
}

func (u *UnHandleMsg) Ctx() context.Context {
	return u.ctx
}

func (u *UnHandleMsg) Router() *route.Route {
	return u.route
}

func (u *UnHandleMsg) Typ() string {
	return u.typ
}

func (h *HandlerService) HandleCustom(conn acceptor.PlayerConn) {
	// create a client agent and startup write goroutine
	a := h.agentFactory.CreateAgent(conn)

	// startup agent goroutine
	go a.Handle()

	logger.Log.Debugf("New session established: %s", a.String())

	// guarantee agent related resource is destroyed
	defer func() {
		a.GetSession().Close()
		logger.Log.Debugf("Session read goroutine exit, SessionID=%d, UID=%s", a.GetSession().ID(), a.GetSession().UID())
	}()

	for {
		msg, err := conn.GetNextMessage()

		if err != nil {
			if err != constants.ErrConnectionClosed {
				logger.Log.Errorf("Error reading next available message: %s", err.Error())
			}

			return
		}

		packets, err := h.decoder.Decode(msg)
		if err != nil {
			logger.Log.Errorf("Failed to decode message: %s", err.Error())
			return
		}

		if len(packets) < 1 {
			logger.Log.Warnf("Read no packets, data: %v", msg)
			continue
		}

		// process all packet
		for i := range packets {
			if err := h.processPacketCustom(a, packets[i]); err != nil {
				logger.Log.Errorf("Failed to process packet: %s", err.Error())
				return
			}
		}
	}
}

func (h *HandlerService) processPacketCustom(a agent.Agent, p *packet.Packet) error {
	switch p.Type {
	case packet.Handshake:
		logger.Log.Debug("Received handshake packet")
		if err := a.SendHandshakeResponse(); err != nil {
			logger.Log.Errorf("Error sending handshake response: %s", err.Error())
			return err
		}
		logger.Log.Debugf("Session handshake Id=%d, Remote=%s", a.GetSession().ID(), a.RemoteAddr())

		// Parse the json sent with the handshake by the client
		handshakeData := &session.HandshakeData{}
		err := json.Unmarshal(p.Data, handshakeData)
		if err != nil {
			a.SetStatus(constants.StatusClosed)
			return fmt.Errorf("Invalid handshake data. Id=%d", a.GetSession().ID())
		}

		a.GetSession().SetHandshakeData(handshakeData)
		a.SetStatus(constants.StatusHandshake)
		err = a.GetSession().Set(constants.IPVersionKey, a.IPVersion())
		if err != nil {
			logger.Log.Warnf("failed to save ip version on session: %q\n", err)
		}

		logger.Log.Debug("Successfully saved handshake data")

	case packet.HandshakeAck:
		a.SetStatus(constants.StatusWorking)
		logger.Log.Debugf("Receive handshake ACK Id=%d, Remote=%s", a.GetSession().ID(), a.RemoteAddr())

	case packet.Data:
		if a.GetStatus() < constants.StatusWorking {
			return fmt.Errorf("receive data on socket which is not yet ACK, session will be closed immediately, remote=%s",
				a.RemoteAddr().String())
		}

		msg, err := message.Decode(p.Data)
		if err != nil {
			return err
		}
		h.processMessageCustom(a, msg)

	case packet.Heartbeat:
		// expected
	}

	a.SetLastAt()
	return nil
}

func (h *HandlerService) processMessageCustom(a agent.Agent, msg *message.Message) {
	requestID := nuid.New()
	ctx := pcontext.AddToPropagateCtx(context.Background(), constants.StartTimeKey, time.Now().UnixNano())
	ctx = pcontext.AddToPropagateCtx(ctx, constants.RouteKey, msg.Route)
	ctx = pcontext.AddToPropagateCtx(ctx, constants.RequestIDKey, requestID)
	tags := opentracing.Tags{
		"local.id":   h.server.ID,
		"span.kind":  "server",
		"msg.type":   strings.ToLower(msg.Type.String()),
		"user.id":    a.GetSession().UID(),
		"request.id": requestID,
	}
	ctx = tracing.StartSpan(ctx, msg.Route, tags)
	ctx = context.WithValue(ctx, constants.SessionCtxKey, a.GetSession())

	r, err := route.Decode(msg.Route)
	if err != nil {
		logger.Log.Errorf("Failed to decode route: %s", err.Error())
		a.AnswerWithError(ctx, msg.ID, e.NewError(err, e.ErrBadRequestCode))
		return
	}

	if r.SvType == "" {
		r.SvType = h.server.Type
	}
	typ := ""
	if r.SvType == h.server.Type {
		typ = "local"
	} else {
		if h.remoteService != nil {
			typ = "remote"
		} else {
			logger.Log.Warnf("request made to another server type but no remoteService running")
			return
		}
	}
	decorateMsg := &UnHandleMsg{
		ctx:   ctx,
		agent: a,
		route: r,
		msg:   msg,
		typ:   typ,
	}
	MsgCtrlInstance.Handle(decorateMsg)

}
