package ar_mapper2

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/viper"

	eh "github.com/looplab/eventhorizon"
	eventpublisher "github.com/looplab/eventhorizon/publisher/local"

	"github.com/superchalupa/sailfish/src/eventwaiter"
	"github.com/superchalupa/sailfish/src/log"
	"github.com/superchalupa/sailfish/src/ocp/event"
	"github.com/superchalupa/sailfish/src/ocp/model"
	domain "github.com/superchalupa/sailfish/src/redfishresource"

	"github.com/superchalupa/sailfish/src/dell-resources/attributes"
)

// what we read from redfish.yaml
type mapping struct {
	Property string
	FQDD     string
	Group    string
	Index    string
	Name     string
}

// individual mapping: only for bookkeeping
type ModelMappings struct {
	logger         log.Logger
	mappings       []mapping
	cfgsect        string
	model          *model.Model
	requestedFQDD  string
	requestedGroup string
	requestedIndex string
}

type ARService struct {
	eb     eh.EventBus
	cfg    *viper.Viper
	cfgMu  *sync.RWMutex
	logger log.Logger

	mappingsMu    sync.RWMutex
	modelmappings map[string]ModelMappings

	hashMu sync.RWMutex
	hash   map[string][]update
}

// post-processed and optimized update
type update struct {
	model    *model.Model
	property string
}

func StartService(ctx context.Context, logger log.Logger, cfg *viper.Viper, cfgMu *sync.RWMutex, eb eh.EventBus) (*ARService, error) {
	logger = logger.New("module", "ar2")

	EventPublisher := eventpublisher.NewEventPublisher()
	eb.AddHandler(eh.MatchAnyEventOf(attributes.AttributeUpdated), EventPublisher)
	EventWaiter := eventwaiter.NewEventWaiter(eventwaiter.SetName("AR Mapper"), eventwaiter.NoAutoRun)
	EventPublisher.AddObserver(EventWaiter)
	go EventWaiter.Run()

	arservice := &ARService{
		eb:            eb,
		cfg:           cfg,
		cfgMu:         cfgMu,
		logger:        logger,
		modelmappings: make(map[string]ModelMappings),
		hash:          make(map[string][]update),
	}

	sp, err := event.NewEventStreamProcessor(ctx, EventWaiter, event.MatchAnyEvent(attributes.AttributeUpdated), event.SetListenerName("ar_service"))
	if err != nil {
		logger.Error("Failed to create event stream processor", "err", err)
		return nil, err
	}
	go sp.RunForever(func(event eh.Event) {
		data, ok := event.Data().(*attributes.AttributeUpdatedData)
		if !ok {
			return
		}

		logger.Debug("processing event", "event", event)
		key := fmt.Sprintf("%s:%s:%s:%s", data.FQDD, data.Group, data.Index, data.Name)
		arservice.hashMu.RLock()
		if arr, ok := arservice.hash[key]; ok {
			logger.Debug("matched quick hash", "key", key)
			for _, u := range arr {
				logger.Debug("updating property", "property", u.property, "value", data.Value)
				u.model.UpdateProperty(u.property, data.Value)
			}
		}
		arservice.hashMu.RUnlock()
	})

	return arservice, nil
}

// used by individual mappings to keep track of the mapping name for PUT/PATCH/POST, etc.
type breadcrumb struct {
	ars         *ARService
	mappingName string
	logger      log.Logger
}

func (ars *ARService) NewMapping(logger log.Logger, mappingName, cfgsection string, m *model.Model, fgin map[string]string) breadcrumb {
	mm := ModelMappings{
		logger:         logger.New("module", "ar2"),
		cfgsect:        cfgsection,
		model:          m,
		mappings:       []mapping{},
		requestedFQDD:  fgin["FQDD"],
		requestedGroup: fgin["Group"],
		requestedIndex: fgin["Index"],
	}

	ars.mappingsMu.Lock()
	ars.modelmappings[mappingName] = mm
	ars.mappingsMu.Unlock()

	ars.loadConfig(mappingName)

	// ars.logger.Info("updating mappings", "mappings", c.mappings)
	// c.createModelProperties(ctx)
	// go c.initialStartupBootstrap(ctx)

	return breadcrumb{ars: ars, mappingName: mappingName, logger: logger}
}

func (b breadcrumb) Close() {
	b.ars.mappingsMu.Lock()
	delete(b.ars.modelmappings, b.mappingName)
	b.ars.mappingsMu.Unlock()

	b.ars.resetConfig()
}

func (b breadcrumb) UpdateRequest(ctx context.Context, property string, value interface{}, auth *domain.RedfishAuthorizationProperty) (interface{}, error) {
	b.logger.Info("UpdateRequest", "property", property, "mappingName", b.mappingName)
	mappings, ok := b.ars.modelmappings[b.mappingName]
	if !ok {
		return nil, errors.New("Could not find mapping: " + b.mappingName)
	}

	for _, mapping := range mappings.mappings {
		if property != mapping.Property {
			continue
		}

		mappings.logger.Info("Sending Update Request", "mapping", mapping, "value", value)
		reqUUID := eh.NewUUID()

		data := &attributes.AttributeUpdateRequestData{
			ReqID:         reqUUID,
			FQDD:          mapping.FQDD,
			Group:         mapping.Group,
			Index:         mapping.Index,
			Name:          mapping.Name,
			Value:         value,
			Authorization: *auth,
		}
		b.ars.eb.PublishEvent(ctx, eh.NewEvent(attributes.AttributeUpdateRequest, data, time.Now()))

		break
	}
	// TODO: wait for event to come back matching request (maybe?)

	return value, nil
}

func (ars *ARService) resetConfig() {
	ars.hashMu.Lock()
	// clear out old mappings in preparation
	for k := range ars.hash {
		delete(ars.hash, k)
	}
	ars.hashMu.Unlock()

	for k, _ := range ars.modelmappings {
		ars.loadConfig(k)
	}
}

// this is the function that viper will call whenever the configuration changes at runtime
func (ars *ARService) loadConfig(mappingName string) {
	ars.mappingsMu.Lock()
	defer ars.mappingsMu.Unlock()
	ars.hashMu.Lock()
	defer ars.hashMu.Unlock()

	ars.logger.Info("Updating Config")

	ars.cfgMu.Lock()
	defer ars.cfgMu.Unlock()

	subCfg := ars.cfg.Sub("mappings")
	if subCfg == nil {
		ars.logger.Warn("missing config file section: 'mappings'")
		return
	}

	k := mappingName
	newmaps := []mapping{}
	err := subCfg.UnmarshalKey(ars.modelmappings[k].cfgsect, &newmaps)
	if err != nil {
		ars.logger.Warn("unamrshal failed", "err", err)
	}

	ars.logger.Info("Loading Config", "mappingName", k, "configsection", ars.modelmappings[k].cfgsect, "mappings", newmaps)

	for mappingIdx, mm := range newmaps {
		if mm.FQDD == "{FQDD}" {
			mm.FQDD = ars.modelmappings[k].requestedFQDD
			ars.modelmappings[k].logger.Debug("Replacing {FQDD} with real fqdd", "fqdd", mm.FQDD)
		}
		if mm.Group == "{GROUP}" {
			mm.Group = ars.modelmappings[k].requestedGroup
			ars.modelmappings[k].logger.Debug("Replacing {GROUP} with real group", "group", mm.Group)
		}
		if mm.Index == "{INDEX}" {
			mm.Index = ars.modelmappings[k].requestedIndex
			ars.modelmappings[k].logger.Debug("Replacing {INDEX} with real index", "index", mm.Index)
		}

		modelmapping := ars.modelmappings[k]
		modelmapping.mappings = append(ars.modelmappings[k].mappings, mm)
		ars.modelmappings[k] = modelmapping

		mapstring := fmt.Sprintf("%s:%s:%s:%s",
			ars.modelmappings[k].mappings[mappingIdx].FQDD,
			ars.modelmappings[k].mappings[mappingIdx].Group,
			ars.modelmappings[k].mappings[mappingIdx].Index,
			ars.modelmappings[k].mappings[mappingIdx].Name,
		)
		updArr, ok := ars.hash[mapstring]
		if !ok {
			updArr = []update{}
		}

		updArr = append(updArr, update{model: ars.modelmappings[k].model, property: mm.Property})
		ars.hash[mapstring] = updArr

		ars.logger.Info("Updated config array", "update_array", updArr)
		ars.logger.Info("finished optimizing hash", "hash", ars.hash)
	}
}
