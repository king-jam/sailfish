package event

import (
	"context"
	"fmt"
	"sync"

	"github.com/Knetic/govaluate"
	eh "github.com/looplab/eventhorizon"
	eventpublisher "github.com/looplab/eventhorizon/publisher/local"
	"github.com/superchalupa/sailfish/src/log"

	"github.com/superchalupa/sailfish/src/looplab/eventwaiter"
)

type Options func(*privateStateStructure) error

type listener interface {
	Wait(context.Context) (eh.Event, error)
	Close()
}

type privateStateStructure struct {
	ctx          context.Context
	filterFn     func(eh.Event) bool
	listener     *eventwaiter.EventListener
	listenerName string
	logger       log.Logger
	EW           eventwaiter.EventWaiter
}

var NewESP func(ctx context.Context, options ...Options) (d *privateStateStructure, err error)

func Setup(ch eh.CommandHandler, eb eh.EventBus) {
	EventPublisher := eventpublisher.NewEventPublisher()
	eb.AddHandler(eh.MatchAny(), EventPublisher)
	EventWaiter := eventwaiter.NewEventWaiter(eventwaiter.SetName("generic eventstream"), eventwaiter.NoAutoRun)
	EventPublisher.AddObserver(EventWaiter)
	go EventWaiter.Run()

	NewESP = func(ctx context.Context, options ...Options) (d *privateStateStructure, err error) {
		return NewEventStreamProcessor(ctx, EventWaiter, options...)
	}
}

func NewEventStreamProcessor(ctx context.Context, ew *eventwaiter.EventWaiter, options ...Options) (d *privateStateStructure, err error) {
	d = &privateStateStructure{
		ctx: ctx,
		// default filter is to process no events
		filterFn:     func(eh.Event) bool { return false },
		listenerName: "SET ME!",
		EW:           *ew,
	}
	err = nil

	for _, o := range options {
		err := o(d)
		if err != nil {
			return nil, err
		}
	}

	d.listener, err = ew.Listen(ctx, d.filterFn)
	if err != nil {
		return
	}

	d.listener.Name = d.listenerName

	return
}

func (d *privateStateStructure) Close() {
	if d.listener != nil {
		d.listener.Close()
		d.listener = nil
	}
}

type syncEvent interface {
	Done()
}

func (d *privateStateStructure) RunForever(fn func(eh.Event)) {
	defer d.Close()

	for {
		func() {
			event, err := d.listener.UnSyncWait(d.ctx)
			if err != nil {
				log.MustLogger("eventstream").Info("Shutting down listener", "err", err)
				return
			}

			// TODO: separation of concerns: this should be factored out into a middleware of some sort...
			// now that we are waiting on the listeners, we can .Done() the waitgroup for the eventwaiter itself
			if e, ok := event.(syncEvent); ok {
				defer e.Done()
			}

			fn(event)
		}()
	}
}

func SetListenerName(name string) func(p *privateStateStructure) error {
	return func(p *privateStateStructure) error {
		p.listenerName = name
		return nil
	}
}

func CustomFilter(fn func(eh.Event) bool) func(p *privateStateStructure) error {
	return func(p *privateStateStructure) error {
		p.filterFn = fn
		return nil
	}
}

func MatchAnyEvent(eventType ...eh.EventType) func(p *privateStateStructure) error {
	return func(p *privateStateStructure) error {
		p.filterFn = func(event eh.Event) bool {
			for _, evt := range eventType {
				if event.EventType() == evt {
					return true
				}
			}
			return false
		}
		return nil
	}
}

func ExpressionFilter(logger log.Logger, expr string, parameters map[string]interface{}, functions map[string]govaluate.ExpressionFunction) func(p *privateStateStructure) error {
	return func(p *privateStateStructure) error {
		functions["string"] = func(args ...interface{}) (interface{}, error) {
			return fmt.Sprint(args[0]), nil
		}

		expression, err := govaluate.NewEvaluableExpressionWithFunctions(expr, functions)
		if err != nil {
			logger.Crit("Expression construction (lexing) failed.", "expression", expr)
			return err
		}

		// make a copy of the incoming parameters because maps are pass-by-reference and we'll race with caller
		expressionMu := sync.Mutex{}
		expressionParameters := map[string]interface{}{}
		for k, v := range parameters {
			expressionParameters[k] = v
		}
		fn := func(ev eh.Event) bool {
			expressionMu.Lock()
			defer expressionMu.Unlock()
			expressionParameters["type"] = string(ev.EventType())
			expressionParameters["data"] = ev.Data()
			expressionParameters["event"] = ev
			result, err := expression.Evaluate(expressionParameters)
			if err == nil {
				if ret, ok := result.(bool); ok {
					return ret
				}
				// LOG ERRROR: expression didn't return BOOL
				logger.Error("Expression did not return a bool.", "expression", expr, "parsed", expression.String())
			}
			// LOG ERRROR: expression evaluation failed
			logger.Crit("Expression evaluation failed.", "expression", expr, "parsed", expression.String(), "err", err, "data", ev.Data())
			return false
		}

		p.filterFn = fn
		return nil
	}
}
