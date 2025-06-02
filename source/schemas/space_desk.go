package schemas

import "go.mongodb.org/mongo-driver/v2/bson"

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
	MessagingProduct string                    `json:"messaging_product" bson:"messaging_product"`
	Metadata         SpaceDeskMessageMetadata  `json:"metadata" bson:"metadata"`
	Contacts         []SpaceDeskMessageContact `json:"contacts" bson:"contacts"`
	Messages         []SpaceDeskMessage        `json:"messages" bson:"messages"`
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

type SpaceDeskMessage struct {
	Type      string                `json:"type" bson:"type"`
	From      string                `json:"from" bson:"from"`
	ID        string                `json:"id" bson:"id"`
	Timestamp string                `json:"timestamp" bson:"timestamp"`
	Text      *SpaceDeskMessageText `json:"text,omitempty" bson:"text,omitempty"`
}

type SpaceDeskMessageText struct {
	Body string `json:"body" bson:"body"`
}
