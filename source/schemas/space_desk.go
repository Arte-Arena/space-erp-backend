package schemas

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type SpaceDesk struct {
}

type SpaceDeskMessageEvent struct {
	ID     bson.ObjectID           `json:"_id" bson:"_id,omitempty"`
	Object string                  `json:"object" bson:"object"`
	Entry  []SpaceDeskMessageEntry `json:"entry" bson:"entry"`
}

type SpaceDeskMessageEntry struct {
	ID      string                   `json:"id" bson:"id"`
	Changes []SpaceDeskMessageChange `json:"changes" bson:"changes"`
}

type SpaceDeskMessageChange struct {
	Value SpaceDeskMessageValue `json:"value" bson:"value"`
	Field string                `json:"field" bson:"field"`
}

type SpaceDeskMessageValue struct {
	MessagingProduct string                    `json:"messaging_product,omitempty" bson:"messaging_product,omitempty"`
	Metadata         *SpaceDeskMessageMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"`
	Contacts         []SpaceDeskMessageContact `json:"contacts,omitempty" bson:"contacts,omitempty"`
	Messages         []SpaceDeskMessage        `json:"messages,omitempty" bson:"messages,omitempty"`
}

type SpaceDeskMessageMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number" bson:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id" bson:"phone_number_id"`
}

type SpaceDeskMessageContact struct {
	Profile SpaceDeskMessageProfile `json:"profile" bson:"profile"`
	WAID    string                  `json:"wa_id" bson:"wa_id"`
}

type SpaceDeskMessageProfile struct {
	Name string `json:"name" bson:"name"`
}

type SpaceDeskMedia struct {
	ID       string `json:"id,omitempty" bson:"id,omitempty"`
	MimeType string `json:"mime_type,omitempty" bson:"mime_type,omitempty"`
	Sha256   string `json:"sha256,omitempty" bson:"sha256,omitempty"`
	File     string `json:"filename,omitempty" bson:"filename,omitempty"`
}

type ButtonInfo struct {
	Text    string `bson:"text,omitempty" json:"text,omitempty"`
	Payload string `bson:"payload,omitempty" json:"payload,omitempty"`
}

type SpaceDeskMessage struct {
	Type        string                   `json:"type" bson:"type"`
	From        string                   `json:"from,omitempty" bson:"from,omitempty"`
	To          string                   `json:"to,omitempty" bson:"to,omitempty"`
	ID          string                   `json:"id,omitempty" bson:"id,omitempty"`
	Timestamp   string                   `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	Text        *SpaceDeskMessageText    `json:"text,omitempty" bson:"text,omitempty"`
	Video       *SpaceDeskMedia          `json:"video,omitempty" bson:"video,omitempty"`
	Sticker     *SpaceDeskMedia          `json:"sticker,omitempty" bson:"sticker,omitempty"`
	Image       *SpaceDeskMedia          `json:"image,omitempty" bson:"image,omitempty"`
	Audio       *SpaceDeskMedia          `json:"audio,omitempty" bson:"audio,omitempty"`
	Document    *SpaceDeskMedia          `json:"document,omitempty" bson:"document,omitempty"`
	Button      *ButtonInfo              `bson:"button,omitempty" json:"button,omitempty"`
	Interactive *SpaceDeskInteractive    `json:"interactive,omitempty" bson:"interactive,omitempty"`
	Context     *SpaceDeskMessageContext `json:"context,omitempty" bson:"context,omitempty"`
	Poll        *SpaceDeskPoll           `json:"poll,omitempty" bson:"poll,omitempty"`
}

type SpaceDeskPoll struct {
	Name                   string   `json:"name" bson:"name"`
	Options                []string `json:"options" bson:"options"`
	SelectableOptionsCount int      `json:"selectable_options_count,omitempty" bson:"selectable_options_count,omitempty"`
}

type SpaceDeskMessageContext struct {
	From string `json:"from,omitempty" bson:"from,omitempty"`
	ID   string `json:"id,omitempty" bson:"id,omitempty"` // ID da mensagem original que foi respondida
}

type SpaceDeskInteractive struct {
	Type        string                           `json:"type" bson:"type"`
	ButtonReply *SpaceDeskInteractiveButtonReply `json:"button_reply,omitempty" bson:"button_reply,omitempty"`
	ListReply   *SpaceDeskInteractiveListReply   `json:"list_reply,omitempty" bson:"list_reply,omitempty"`
}

type SpaceDeskInteractiveButtonReply struct {
	ID    string `json:"id" bson:"id"`
	Title string `json:"title" bson:"title"`
}

type SpaceDeskInteractiveListReply struct {
	ID          string `json:"id" bson:"id"`
	Title       string `json:"title" bson:"title"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
}

type SpaceDeskMessageText struct {
	Body string `json:"body" bson:"body"`
}

type SpaceDeskChatMetadata struct {
	ID                bson.ObjectID   `json:"_id" bson:"_id,omitempty"`
	Name              string          `json:"name" bson:"name"`
	NickName          string          `json:"nick_name" bson:"nick_name"`
	ClientPhoneNumber string          `json:"cliente_phone_number" bson:"cliente_phone_number"`
	Description       string          `json:"description" bson:"description"`
	Status            string          `json:"status" bson:"status"`
	Type              string          `json:"type" bson:"type"`
	GroupIds          []bson.ObjectID `json:"group_ids" bson:"group_ids"`
	UserId            string          `json:"user_id" bson:"user_id"`
	CreatedAt         time.Time       `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" bson:"updated_at"`
	LastMessage       time.Time       `json:"last_message_timestamp" bson:"last_message_timestamp"`
}

type Group struct {
	ID     bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name   string        `bson:"name" json:"name"`
	Status string        `bson:"status" json:"status"`
	Type   string        `bson:"type" json:"type"`
	Chats  []string      `bson:"chats" json:"chats"`
}

type SpaceDeskStatus struct {
	ID           string              `json:"id" bson:"id"`
	Status       string              `json:"status" bson:"status"`
	Timestamp    string              `json:"timestamp" bson:"timestamp"`
	RecipientID  string              `json:"recipient_id" bson:"recipient_id"`
	Conversation *StatusConversation `json:"conversation,omitempty" bson:"conversation,omitempty"`
	Pricing      *StatusPricing      `json:"pricing,omitempty" bson:"pricing,omitempty"`
	Errors       []StatusError       `json:"errors,omitempty" bson:"errors,omitempty"`
}

// StatusConversation contém informações sobre a conversa associada ao status.
type StatusConversation struct {
	ID     string       `json:"id" bson:"id"`
	Origin StatusOrigin `json:"origin" bson:"origin"`
}

// StatusOrigin descreve a origem da conversa.
type StatusOrigin struct {
	Type string `json:"type" bson:"type"`
}

// StatusPricing detalha o custo da mensagem.
type StatusPricing struct {
	Billable     bool   `json:"billable" bson:"billable"`
	PricingModel string `json:"pricing_model" bson:"pricing_model"`
	Category     string `json:"category" bson:"category"`
}

// StatusError contém detalhes de um erro de entrega.
type StatusError struct {
	Code      int             `json:"code" bson:"code"`
	Title     string          `json:"title" bson:"title"`
	Message   string          `json:"message" bson:"message"`
	ErrorData StatusErrorData `json:"error_data" bson:"error_data"`
}

// StatusErrorData contém os detalhes específicos do erro.
type StatusErrorData struct {
	Details string `json:"details" bson:"details"`
}
