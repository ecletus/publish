package publish

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ecletus/admin"
	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/ecletus/roles"
	"github.com/ecletus/worker"
)

const (
	// PublishPermission publish permission
	PublishPermission roles.PermissionMode = "publish"
)

type publishController struct {
	*Publish
}

type visiblePublishResourceInterface interface {
	VisiblePublishResource(*core.Context) bool
}

func (pc *publishController) Preview(context *admin.Context) {
	type resource struct {
		*admin.Resource
		Value interface{}
	}

	var drafts = []resource{}

	draftDB := context.DB().Set(publishDraftMode, true).Unscoped()

	context.Admin.WalkResources(func(res *admin.Resource) bool {
		if visibleInterface, ok := res.Value.(visiblePublishResourceInterface); ok {
			if !visibleInterface.VisiblePublishResource(context.Context) {
				return true
			}
		} else if res.Config.Invisible {
			return true
		}

		if context.HasPermission(res, PublishPermission) {
			results := res.NewSlice()
			if IsPublishableModel(res.Value) || IsPublishEvent(res.Value) {
				if pc.SearchHandler(draftDB.Where("publish_status = ?", DIRTY), context.Context).Find(results).RowsAffected > 0 {
					drafts = append(drafts, resource{
						Resource: res,
						Value:    results,
					})
				}
			}
		}
		return true
	})
	context.Execute("publish_drafts", drafts)
}

func (pc *publishController) Diff(context *admin.Context) {
	var (
		resourceID = context.Request.URL.Query().Get(":publish_unique_key")
		params     = strings.Split(resourceID, "__") // name__primary_keys
		res        = context.Admin.GetResourceByID(params[0])
	)

	draft := res.NewStruct(context.Site)
	pc.search(context.DB().Set(publishDraftMode, true), res, [][]string{params[1:]}).First(draft)

	production := res.NewStruct(context.Site)
	pc.search(context.DB().Set(publishDraftMode, false), res, [][]string{params[1:]}).First(production)

	context.Include(context.Writer, "publish_diff", map[string]interface{}{
		"Production": production,
		"Draft":      draft,
		"Resource":   res,
	})
}

func (pc *publishController) PublishOrDiscard(context *admin.Context) {
	var request = context.Request
	var ids = request.Form["checked_ids[]"]

	if scheduler := pc.Publish.WorkerScheduler; scheduler != nil {
		jobResource := scheduler.JobResource
		result := jobResource.NewStruct(context.Site).(worker.QorJobInterface)
		if request.Form.Get("publish_type") == "discard" {
			result.SetJob(scheduler.GetRegisteredJob(DISCARD_KEY))
		} else {
			result.SetJob(scheduler.GetRegisteredJob(PUBLISH_KEY))
		}

		workerArgument := &QorWorkerArgument{IDs: ids}
		if t, err := utils.ParseTime(request.Form.Get("scheduled_time"), context.Context); err == nil {
			workerArgument.ScheduleTime = &t
		}
		result.SetSerializableArgumentValue(workerArgument)

		jobResource.Crud(context.Context).Create(result)
		scheduler.AddJob(result)

		http.Redirect(context.Writer, context.Request, context.URLFor(jobResource), http.StatusFound)
	} else {
		records := pc.searchWithPublishIDs(context.DB().Set(publishDraftMode, true), context.Admin, ids)

		if request.Form.Get("publish_type") == "publish" {
			pc.Publish.Publish(records...)
		} else if request.Form.Get("publish_type") == "discard" {
			pc.Publish.Discard(records...)
		}

		http.Redirect(context.Writer, context.Request, context.Request.RequestURI, http.StatusFound)
	}
}

// ConfigureQorResourceBeforeInitialize configure qor resource when initialize qor admin
func (publish *Publish) ConfigureResourceBeforeInitialize(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		res.UseTheme("publish")

		if event := res.GetAdmin().GetResourceByID("PublishEvent"); event == nil {
			eventResource := res.GetAdmin().AddResource(&PublishEvent{}, &admin.Config{Invisible: true})
			eventResource.IndexAttrs("Name", "Description", "CreatedAt")
		}
	}
}

// ConfigureQorResource configure qor resource for qor admin
func (publish *Publish) ConfigureResource(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		controller := publishController{publish}
		res.Router.Get("/diff/{publish_unique_key}", controller.Diff)
		res.Router.Get("/", controller.Preview)
		res.Router.Post("/", controller.PublishOrDiscard)

		res.GetAdmin().RegisterFuncMap("publish_unique_key", func(res *admin.Resource, record interface{}, context *admin.Context) string {
			var publishKeys = []string{res.ToParam()}
			var scope = publish.DB.NewScope(record)
			for _, primaryField := range scope.PrimaryFields() {
				publishKeys = append(publishKeys, fmt.Sprint(primaryField.Field.Interface()))
			}
			return strings.Join(publishKeys, "__")
		})

		res.GetAdmin().RegisterFuncMap("is_publish_event_resource", func(res *admin.Resource) bool {
			return IsPublishEvent(res.Value)
		})
	}
}
