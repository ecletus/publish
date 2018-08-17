package publish

import "github.com/moisespsena-go/aorm"

type publishJoinTableHandler struct {
	aorm.JoinTableHandler
}

func (handler publishJoinTableHandler) Table(db *aorm.DB) string {
	if IsDraftMode(db) {
		return handler.TableName + "_draft"
	}
	return handler.TableName
}

func (handler publishJoinTableHandler) Add(h aorm.JoinTableHandlerInterface, db *aorm.DB, source1 interface{}, source2 interface{}) error {
	// production mode
	if !IsDraftMode(db) {
		if err := handler.JoinTableHandler.Add(h, db.Set(publishDraftMode, true), source1, source2); err != nil {
			return err
		}
	}
	return handler.JoinTableHandler.Add(h, db, source1, source2)
}

func (handler publishJoinTableHandler) Delete(h aorm.JoinTableHandlerInterface, db *aorm.DB, sources ...interface{}) error {
	// production mode
	if !IsDraftMode(db) {
		if err := handler.JoinTableHandler.Delete(h, db.Set(publishDraftMode, true), sources...); err != nil {
			return err
		}
	}
	return handler.JoinTableHandler.Delete(h, db, sources...)
}
