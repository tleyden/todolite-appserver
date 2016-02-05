package libtodolite

import (
	"fmt"
)

type Attachments map[string]interface{}

const (
	Task    = "task"
	List    = "list"
	Profile = "profile"
)

type TodoItem struct {
	Revision    string      `json:"_rev"`
	Id          string      `json:"_id"`
	Checked     bool        `json:"checked"`
	CreatedAt   string      `json:"created_at"`
	ListId      string      `json:"list_id"`
	Title       string      `json:"title"`
	Type        string      `json:"type"`
	OcrDecoded  string      `json:"ocr_decoded"`
	Attachments Attachments `json:"_attachments"`
}

type TodoList struct {
	Revision  string   `json:"_rev"`
	Id        string   `json:"_id"`
	Checked   bool     `json:"checked"`
	CreatedAt string   `json:"created_at"`
	Title     string   `json:"title"`
	Type      string   `json:"type"`
	UpdatedAt string   `json:"updated_at"`
	Owner     string   `json:"owner"`
	Members   []string `json:"members"`
}

func (t TodoItem) AttachmentUrl(dbUrl string) string {

	attachmentUrl := ""
	for k, _ := range t.Attachments {
		attachmentUrl = fmt.Sprintf("%s/%s/%s", dbUrl, t.Id, k)
		return attachmentUrl
	}
	return attachmentUrl

}
