package common

import (
	"context"
	"errors"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/server/common/logger"
	"github.com/streadway/amqp"
)

// AMQPPublisher represent the ability to push a message to a broker.
type AMQPPublisher interface {
	Publish(ctx context.Context, routingKey string, data []byte) error
}

// Session object wraps amqp library.
// It automatically reconnects when the connection fails,
// and blocks all pushes until the connection succeeds.
// It also confirms every outgoing message, so none are lost.
// Implementation is adapted from https://github.com/streadway/amqp/blob/master/example_client_test.go
type Session struct {
	name            string
	connection      *amqp.Connection
	channel         *amqp.Channel
	done            chan bool
	notifyConnClose chan *amqp.Error
	notifyChanClose chan *amqp.Error
	notifyConfirm   chan amqp.Confirmation
	isReady         bool
}

const (
	// When reconnecting to the server after connection failure
	reconnectDelay = 5 * time.Second

	// When setting up the channel after a channel exception
	reInitDelay = 2 * time.Second

	// When resending messages the server didn't confirm
	resendDelay = 5 * time.Second
)

var (
	errNotConnected  = errors.New("not connected to a server")
	errAlreadyClosed = errors.New("already closed: not connected to the server")
	errShutdown      = errors.New("session is shutting down")
)

// NewSession creates a new consumer state instance, and automatically
// attempts to connect to the server.
// Session's name will be used to define the exchange on which events are published.
func NewSession(name string, addr string) *Session {
	session := Session{
		name: name,
		done: make(chan bool),
	}
	go session.handleReconnect(addr)

	for !session.isReady {
		log.WithField("delay", reconnectDelay).Info("AMQP session not yet ready, waiting")
		<-time.After(reconnectDelay)
	}

	return &session
}

// handleReconnect will wait for a connection error on
// notifyConnClose, and then continuously attempt to reconnect.
func (session *Session) handleReconnect(addr string) {
	for {
		session.isReady = false
		log.Info("Attempting to connect to AMQP broker")

		conn, err := session.connect(addr)

		if err != nil {
			log.WithError(err).Warn("Failed to connect to broker. Retrying...")

			select {
			case <-session.done:
				return
			case <-time.After(reconnectDelay):
			}
			continue
		}

		if done := session.handleReInit(conn); done {
			break
		}
	}
}

// connect will create a new AMQP connection
func (session *Session) connect(addr string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(addr)

	if err != nil {
		return nil, err
	}

	session.changeConnection(conn)
	log.Info("Connected!")
	return conn, nil
}

// handleReconnect will wait for a channel error
// and then continuously attempt to re-initialize both channels
func (session *Session) handleReInit(conn *amqp.Connection) bool {
	for {
		session.isReady = false

		err := session.init(conn)

		if err != nil {
			log.Warn("Failed to initialize channel. Retrying...")

			select {
			case <-session.done:
				return true
			case <-time.After(reInitDelay):
			}
			continue
		}

		select {
		case <-session.done:
			return true
		case <-session.notifyConnClose:
			log.Info("AMQP connection closed. Reconnecting...")
			return false
		case <-session.notifyChanClose:
			log.Info("AMQP channel closed. Re-running init...")
		}
	}
}

// init will initialize channel & declare queue
func (session *Session) init(conn *amqp.Connection) error {
	ch, err := conn.Channel()

	if err != nil {
		return err
	}

	err = ch.Confirm(false)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(
		session.name, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return err
	}

	session.changeChannel(ch)
	session.isReady = true
	log.WithField("queue", session.name).Debug("AMQP session ready")

	return nil
}

// changeConnection takes a new connection to the queue,
// and updates the close listener to reflect this.
func (session *Session) changeConnection(connection *amqp.Connection) {
	session.connection = connection
	session.notifyConnClose = make(chan *amqp.Error)
	session.connection.NotifyClose(session.notifyConnClose)
}

// changeChannel takes a new channel to the queue,
// and updates the channel listeners to reflect this.
func (session *Session) changeChannel(channel *amqp.Channel) {
	session.channel = channel
	session.notifyChanClose = make(chan *amqp.Error)
	session.notifyConfirm = make(chan amqp.Confirmation, 1)
	session.channel.NotifyClose(session.notifyChanClose)
	session.channel.NotifyPublish(session.notifyConfirm)
}

// Publish will push data onto the queue, and wait for a confirm.
// If no confirms are received until within the resendTimeout,
// it continuously re-sends messages until a confirm is received.
// This will block until the server sends a confirm. Errors are
// only returned if the push action itself fails, see UnsafePush.
func (session *Session) Publish(ctx context.Context, routingKey string, data []byte) error {
	log := logger.Get(ctx).WithField("numBytes", len(data))

	if !session.isReady {
		return errors.New("failed to push message: not connected")
	}
	for {
		err := session.UnsafePush(ctx, (routingKey), data)
		if err != nil {
			log.WithError(err).Warn("Push failed. Retrying...")
			select {
			case <-session.done:
				return errShutdown
			case <-time.After(resendDelay):
			}
			continue
		}
		select {
		case confirm := <-session.notifyConfirm:
			if confirm.Ack {
				return nil
			}
		case <-time.After(resendDelay):
		}
		log.Warn("Push didn't confirm. Retrying...")
	}
}

// UnsafePush will push to the queue without checking for
// confirmation. It returns an error if it fails to connect.
// No guarantees are provided for whether the server will
// recieve the message.
func (session *Session) UnsafePush(ctx context.Context, routingKey string, data []byte) error {
	if !session.isReady {
		return errNotConnected
	}
	return session.channel.Publish(
		session.name, // Exchange
		routingKey,   // Routing key
		false,        // Mandatory
		false,        // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)
}

// Close will cleanly shutdown the channel and connection.
func (session *Session) Close() error {
	if !session.isReady {
		return errAlreadyClosed
	}
	err := session.channel.Close()
	if err != nil {
		return err
	}
	err = session.connection.Close()
	if err != nil {
		return err
	}
	close(session.done)
	session.isReady = false
	return nil
}
