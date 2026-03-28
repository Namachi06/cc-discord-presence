package discord

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Opcodes for Discord IPC
const (
	opHandshake = 0
	opFrame     = 1
)

// Button represents a clickable button in Discord Rich Presence
type Button struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// Activity represents Discord Rich Presence activity
type Activity struct {
	Details    string     `json:"details,omitempty"`
	State      string     `json:"state,omitempty"`
	LargeImage string     `json:"large_image,omitempty"`
	LargeText  string     `json:"large_text,omitempty"`
	SmallImage string     `json:"small_image,omitempty"`
	SmallText  string     `json:"small_text,omitempty"`
	StartTime  *time.Time `json:"-"`
	Buttons    []Button   `json:"-"`
}

// Conn interface for both unix sockets and windows pipes
type Conn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
}

// Client handles Discord RPC connection
type Client struct {
	clientID    string
	conn        Conn
	connected   bool
	ConnectFunc func() (Conn, error) // nil = use default, for testing
}

// NewClient creates a new Discord RPC client
func NewClient(clientID string) *Client {
	return &Client{clientID: clientID}
}

// IsConnected returns whether the client has an active Discord connection.
func (c *Client) IsConnected() bool {
	return c.connected
}

// Connect establishes connection to Discord
func (c *Client) Connect() error {
	var conn Conn
	var err error
	if c.ConnectFunc != nil {
		conn, err = c.ConnectFunc()
	} else {
		conn, err = c.connectToDiscord()
	}
	if err != nil {
		c.connected = false
		return err
	}
	c.conn = conn

	// Send handshake
	handshake := map[string]interface{}{
		"v":         1,
		"client_id": c.clientID,
	}
	if err := c.send(opHandshake, handshake); err != nil {
		c.conn.Close()
		c.conn = nil
		c.connected = false
		return fmt.Errorf("handshake failed: %w", err)
	}

	// Read handshake response
	if _, err := c.receive(); err != nil {
		c.conn.Close()
		c.conn = nil
		c.connected = false
		return fmt.Errorf("handshake response failed: %w", err)
	}

	c.connected = true
	return nil
}

// Reconnect closes the old connection and attempts a fresh Connect.
func (c *Client) Reconnect() error {
	c.connected = false
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	return c.Connect()
}

// SetActivity updates the Discord Rich Presence
func (c *Client) SetActivity(activity Activity) error {
	if !c.connected {
		return fmt.Errorf("not connected")
	}

	// Build timestamps if StartTime is set
	var timestamps map[string]int64
	if activity.StartTime != nil {
		timestamps = map[string]int64{
			"start": activity.StartTime.Unix(),
		}
	}

	// Build assets
	assets := map[string]string{}
	if activity.LargeImage != "" {
		assets["large_image"] = activity.LargeImage
	}
	if activity.LargeText != "" {
		assets["large_text"] = activity.LargeText
	}
	if activity.SmallImage != "" {
		assets["small_image"] = activity.SmallImage
	}
	if activity.SmallText != "" {
		assets["small_text"] = activity.SmallText
	}

	activityData := map[string]interface{}{}
	if activity.Details != "" {
		activityData["details"] = activity.Details
	}
	if activity.State != "" {
		activityData["state"] = activity.State
	}
	if len(assets) > 0 {
		activityData["assets"] = assets
	}
	if timestamps != nil {
		activityData["timestamps"] = timestamps
	}
	if len(activity.Buttons) > 0 {
		var buttons []map[string]string
		for _, b := range activity.Buttons {
			buttons = append(buttons, map[string]string{
				"label": b.Label,
				"url":   b.URL,
			})
		}
		activityData["buttons"] = buttons
	}

	payload := map[string]interface{}{
		"cmd": "SET_ACTIVITY",
		"args": map[string]interface{}{
			"pid":      os.Getpid(),
			"activity": activityData,
		},
		"nonce": fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	return c.send(opFrame, payload)
}

// ClearActivity removes the Discord Rich Presence.
func (c *Client) ClearActivity() error {
	if !c.connected {
		return fmt.Errorf("not connected")
	}

	payload := map[string]interface{}{
		"cmd": "SET_ACTIVITY",
		"args": map[string]interface{}{
			"pid": os.Getpid(),
		},
		"nonce": fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	return c.send(opFrame, payload)
}

// Close disconnects from Discord
func (c *Client) Close() error {
	c.connected = false
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

func (c *Client) send(opcode int, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Discord IPC frame: [opcode:4bytes][length:4bytes][payload]
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, int32(opcode))
	binary.Write(buf, binary.LittleEndian, int32(len(payload)))
	buf.Write(payload)

	_, err = c.conn.Write(buf.Bytes())
	return err
}

func (c *Client) receive() ([]byte, error) {
	header := make([]byte, 8)
	if _, err := io.ReadFull(c.conn, header); err != nil {
		return nil, err
	}

	length := binary.LittleEndian.Uint32(header[4:8])
	const maxPayloadSize = 64 * 1024
	if length > maxPayloadSize {
		return nil, fmt.Errorf("payload too large: %d bytes", length)
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(c.conn, payload); err != nil {
		return nil, err
	}

	return payload, nil
}
