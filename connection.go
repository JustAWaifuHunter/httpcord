package httpcord

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type HttpConnection int

const (
	DefaultHttpConnection HttpConnection = iota + 1
	FastHttpConnection
)

type ConnectionContext struct {
	SendRes     func(res *InteractionResponse) bool
	Interaction Interaction
	clientToken string
}

type ConnectionOptions struct {
	// Use github.com/valyala/fasthttp instead net/http
	HttpConnection HttpConnection
	// Discord public key
	PublicKey string
	// Discord token (Necessary for external requests)
	Token string
}

type Connection struct {
	FastHandler    fasthttp.RequestHandler
	DefaultHandler http.HandlerFunc
}

var InteractionHandlers = make([]func(ctx ConnectionContext), 0, 10)

func parsePublicKey(key string) (ed25519.PublicKey, error) {
	return hex.DecodeString(key)
}

func verifyKey(body []byte, signature string, publicKey ed25519.PublicKey) bool {
	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	return ed25519.Verify(publicKey, body, sig)
}

func NewConnection(options ConnectionOptions) Connection {
	publicKey, err := parsePublicKey(options.PublicKey)

	if err != nil {
		panic(err)
	}

	handler := httpHandler(publicKey, options.Token)
	if options.HttpConnection == DefaultHttpConnection {
		return Connection{
			DefaultHandler: handler,
		}
	} else if options.HttpConnection == FastHttpConnection {
		return Connection{
			FastHandler: fasthttpadaptor.NewFastHTTPHandler(handler),
		}
	}

	return Connection{
		DefaultHandler: handler,
	}

}

func (c Connection) Connect(address string) error {
	if c.FastHandler != nil {
		return fasthttp.ListenAndServe(address, c.FastHandler)
	}

	return http.ListenAndServe(address, c.DefaultHandler)
}

func httpHandler(publicKey ed25519.PublicKey, token string) http.HandlerFunc {
	var res InteractionResponse

	return func(w http.ResponseWriter, r *http.Request) {
		je := json.NewEncoder(w)
		signature := r.Header.Get("X-Signature-Ed25519")
		timestamp := r.Header.Get("X-Signature-Timestamp")

		bodyBytes, err := ioutil.ReadAll(r.Body)

		if err != nil {
			panic(err)
		}

		body := append([]byte(timestamp), bodyBytes...)

		if !verifyKey(body, signature, publicKey) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var rawInteraction APIInteraction
		err = json.Unmarshal(bodyBytes, &rawInteraction)

		if err != nil {
			panic("Error on get interaction: " + err.Error())
		}

		interaction := ResolveInteraction(&rawInteraction)

		if interaction.Type == PingInteraction {
			w.Header().Set("Content-Type", "application/json")
			err := je.Encode(InteractionResponse{
				Type: PongResponse,
			})

			if err != nil {
				panic("Error writing response")
			}
			return
		}

		if (res.Type == ChannelMessageWithSourceResponse || res.Type == UpdateMessageResponse) && len(res.Data.Files) > 0 {
			m := multipart.NewWriter(w)
			w.Header().Set("Content-Type", m.FormDataContentType())

			for id, file := range res.Data.Files {
				attach, err := file.MakeAttach(Snowflake(rune(id+1)), m)

				if err != nil {
					panic("Error creating attachment: " + err.Error())
				}

				res.Data.Attachments = append(res.Data.Attachments, attach)
			}

			if field, err := m.CreateFormField("payload_json"); err != nil {
				panic("Error creating payload_json form field")
			} else if err := json.NewEncoder(field).Encode(res); err != nil {
				panic("Error encoding payload_json")
			}

			if err := m.Close(); err != nil {
				panic("Error on close multipart writer")
			}

			return
		}

		ctx := ConnectionContext{
			Interaction: interaction,
			SendRes: func(r *InteractionResponse) bool {
				w.Header().Set("Content-Type", "application/json")
				err = je.Encode(r)
				return err != nil
			},
			clientToken: token,
		}

		for _, h := range InteractionHandlers {
			h(ctx)
		}
	}
}

func (c Connection) AddInteractionHandler(handler func(ctx ConnectionContext)) {
	InteractionHandlers = append(InteractionHandlers, handler)
}

func (ctx *ConnectionContext) ReplyInteraction(data *InteractionCallbackData) {
	ctx.SendRes(&InteractionResponse{
		Type: ChannelMessageWithSourceResponse,
		Data: data,
	})
}

func (ctx *ConnectionContext) DeferReplyInteraction() {
	ctx.SendRes(&InteractionResponse{
		Type: DeferredChannelMessageWithSourceResponse,
	})
}

func (ctx *ConnectionContext) DeferUpdateInteraction() {
	ctx.SendRes(&InteractionResponse{
		Type: DeferredUpdateResponse,
	})
}

func (ctx *ConnectionContext) EditReply(data *WebhookEdit) {
	EditOriginalInteractionResponse(ctx.Interaction.ApplicationID.String(), ctx.Interaction.Token, data)
}

func (ctx *ConnectionContext) DeleteReply() {
	DeleteOriginalInteractionResponse(ctx.Interaction.ApplicationID.String(), ctx.Interaction.Token)
}

func (ctx *ConnectionContext) FollowUp(data *WebhookEdit) {
	FollowUpInteractionResponse(ctx.Interaction.ApplicationID.String(), ctx.Interaction.Token, data)
}
