package publish

import (
	"errors"
	"fmt"

	"github.com/moisespsena-go/aorm"
	"github.com/ecletus/core"
)

// EventInterface defined methods needs for a publish event
type EventInterface interface {
	Publish(db *aorm.DB, event PublishEventInterface) error
	Discard(db *aorm.DB, event PublishEventInterface) error
}

var events = map[string]EventInterface{}

// RegisterEvent register publish event
func RegisterEvent(name string, event EventInterface) {
	events[name] = event
}

// PublishEvent default publish event model
type PublishEvent struct {
	aorm.Model
	Name          string
	Description   string
	Argument      string `sql:"size:65532"`
	PublishStatus bool
	PublishedBy   string
}

func getCurrentUser(db *aorm.DB) (string, bool) {
	if user, hasUser := db.Get("qor:current_user"); hasUser {
		var currentUser string
		if primaryField := db.NewScope(user).PrimaryField(); primaryField != nil {
			currentUser = fmt.Sprintf("%v", primaryField.Field.Interface())
		} else {
			currentUser = fmt.Sprintf("%v", user)
		}

		return currentUser, true
	}

	return "", false
}

// Publish publish data
func (publishEvent *PublishEvent) Publish(db *aorm.DB) error {
	if event, ok := events[publishEvent.Name]; ok {
		err := event.Publish(db, publishEvent)
		if err == nil {
			var updateAttrs = map[string]interface{}{"PublishStatus": PUBLISHED}
			if user, hasUser := getCurrentUser(db); hasUser {
				updateAttrs["PublishedBy"] = user
			}
			err = db.Model(publishEvent).Update(updateAttrs).Error
		}
		return err
	}
	return errors.New("event not found")
}

// Discard discard data
func (publishEvent *PublishEvent) Discard(db *aorm.DB) error {
	if event, ok := events[publishEvent.Name]; ok {
		err := event.Discard(db, publishEvent)
		if err == nil {
			var updateAttrs = map[string]interface{}{"PublishStatus": PUBLISHED}
			if user, hasUser := getCurrentUser(db); hasUser {
				updateAttrs["PublishedBy"] = user
			}
			err = db.Model(publishEvent).Update(updateAttrs).Error
		}
		return err
	}
	return errors.New("event not found")
}

// VisiblePublishResource force to display publish event in publish drafts even it is hidden in the menus
func (PublishEvent) VisiblePublishResource(*core.Context) bool {
	return true
}
