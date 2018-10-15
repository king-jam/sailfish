package attributes

import (
	"context"
	"errors"

	eh "github.com/looplab/eventhorizon"
	"github.com/spf13/viper"

	"github.com/superchalupa/sailfish/src/log"
	"github.com/superchalupa/sailfish/src/ocp/testaggregate"
	"github.com/superchalupa/sailfish/src/ocp/view"
)

func RunRegistryFunctions(ch eh.CommandHandler, eb eh.EventBus) {
	// controller
	RegisterARMapper(ch, eb)
}

func RegisterARMapper(ch eh.CommandHandler, eb eh.EventBus) {
	testaggregate.RegisterControllerFunction("ARDumper", func(ctx context.Context, logger log.Logger, cfgMgr *viper.Viper, vw *view.View, cfg interface{}, parameters map[string]interface{}) error {
		cfgParams, ok := cfg.(map[interface{}]interface{})
		if !ok {
			logger.Crit("Failed to type assert cfg to string", "cfg", cfg)
			return errors.New("Failed to type assert expression to string")
		}

		// ctx, logger, viper, *model, cfg_name, params
		modelName, ok := cfgParams["modelname"]
		if !ok {
			modelName = "default"
		}
		modelNameStr, ok := modelName.(string)
		if !ok {
			modelNameStr = "default"
		}

		addToView, ok := cfgParams["AddToView"]
		if !ok {
			addToView = true
		}
		addToViewBool, ok := addToView.(bool)
		if !ok {
			addToViewBool = true
		}

		fqdd, ok := parameters["fqddlist"]
		if !ok {
			logger.Crit("Required parameter 'fqddlist' missing from parameters to InstantiateFromCfg()")
			return nil
		}

		fqddlist, ok := fqdd.([]string)
		if !ok {
			logger.Crit("Required parameter 'fqddlist' should be an array")
			return nil
		}

		logger.Info("Creating ar_dumper controller", "modelName", modelNameStr, "fqddList", fqddlist)
		dumper, _ := NewController(ctx, vw.GetModel(modelNameStr), fqddlist, ch, eb)

		if addToViewBool {
			// there can be only one
			vw.ApplyOption(view.WithController("ar_dump", dumper))
		}

		return nil
	})
}