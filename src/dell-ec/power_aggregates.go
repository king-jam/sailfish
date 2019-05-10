package dell_ec

import (
	"context"
	"errors"
	"fmt"
	"sync"

	eh "github.com/looplab/eventhorizon"
	"github.com/spf13/viper"

	"github.com/superchalupa/sailfish/src/log"
	"github.com/superchalupa/sailfish/src/ocp/awesome_mapper2"
	"github.com/superchalupa/sailfish/src/ocp/testaggregate"
	"github.com/superchalupa/sailfish/src/ocp/view"
	domain "github.com/superchalupa/sailfish/src/redfishresource"
)

func initpowercontrol(logger log.Logger) {
	powercap_enabled := false
	awesome_mapper2.AddFunction("set_power_cap_enable", func(args ...interface{}) (interface{}, error) {
		powercap_setting, ok := args[0].(string)
		if !ok {
			logger.Crit("Mapper configuration error: Need power cap setting as a string", "args[0]", args[0], "TYPE", fmt.Sprintf("%T", args[0]))
			return nil, errors.New("Mapper configuration error: power cap setting not a string")
		}
		if powercap_setting == "Enabled" {
			powercap_enabled = true
		} else {
			powercap_enabled = false
		}
		return nil, nil
	})

	awesome_mapper2.AddFunction("check_power_cap", func(args ...interface{}) (interface{}, error) {
		powercap_value, ok := args[0].(float64)
		if !ok {
			logger.Crit("Mapper configuration error: Need power cap value as a number", "args[0]", args[0], "TYPE", fmt.Sprintf("%T", args[0]))
			return nil, errors.New("Mapper configuration error: power cap value is not a number")
		}
		if powercap_enabled == true {
			return int(powercap_value), nil
		} else {
			return 0, nil
		}
	})

}

func RegisterAggregate(s *testaggregate.Service) {
	s.RegisterAggregateFunction("power",
		func(ctx context.Context, subLogger log.Logger, cfgMgr *viper.Viper, cfgMgrMu *sync.RWMutex, vw *view.View, extra interface{}, params map[string]interface{}) ([]eh.Command, error) {
			return []eh.Command{
				&domain.CreateRedfishResource{
					ResourceURI: vw.GetURI(),
					Type:        "#Power.v1_0_2.Power",
					Context:     params["rooturi"].(string) + "/$metadata#Power.Power",
					Privileges: map[string]interface{}{
						"GET": []string{"Login"},
					},
					Properties: map[string]interface{}{
						"Id":          "Power",
						"Description": "Power",
						"Name":        "Power",
						"@odata.etag": `W/"abc123"`,

						"PowerSupplies@meta":             vw.Meta(view.GETProperty("power_supply_uris"), view.GETFormatter("expand"), view.GETModel("default")),
						"PowerSupplies@odata.count@meta": vw.Meta(view.GETProperty("power_supply_uris"), view.GETFormatter("count"), view.GETModel("default")),
						"PowerControl@meta":              vw.Meta(view.GETProperty("power_control_uris"), view.GETFormatter("expand"), view.GETModel("default")),
						"PowerControl@odata.count@meta":  vw.Meta(view.GETProperty("power_control_uris"), view.GETFormatter("count"), view.GETModel("default")),
						"Oem": map[string]interface{}{
							"Dell": map[string]interface{}{
								"PowerSuppliesSummary": map[string]interface{}{
									"Status": map[string]interface{}{
										"HealthRollup@meta": vw.Meta(view.GETProperty("psu_rollup"), view.GETModel("global_health")),
									},
								},
								"PowerTrends@meta":             vw.Meta(view.GETProperty("power_trends_uri"), view.GETFormatter("expand"), view.GETModel("default")),
								"PowerTrends@odata.count@meta": vw.Meta(view.GETProperty("power_trends_uri"), view.GETFormatter("count"), view.GETModel("default")),
							},
						}}},
			}, nil
		})

	s.RegisterAggregateFunction("power_trends",
		func(ctx context.Context, subLogger log.Logger, cfgMgr *viper.Viper, cfgMgrMu *sync.RWMutex, vw *view.View, extra interface{}, params map[string]interface{}) ([]eh.Command, error) {
			return []eh.Command{
				&domain.CreateRedfishResource{

					ResourceURI: vw.GetURI(),
					Type:        "#DellPower.v1_0_0.DellPowerTrends",
					Context:     "/redfish/v1/$metadata#Power.Power",
					Privileges: map[string]interface{}{
						"GET": []string{"Login"},
					},
					Properties: map[string]interface{}{
						"Name":                        "System Power",
						"MemberId":                    "PowerHistogram",
						"histograms@meta":             vw.Meta(view.GETProperty("trend_histogram_uris"), view.GETFormatter("expand"), view.GETModel("default")),
						"histograms@odata.count@meta": vw.Meta(view.GETProperty("trend_histogram_uris"), view.GETFormatter("count"), view.GETModel("default")),
					}},
			}, nil
		})

	s.RegisterAggregateFunction("trend_histogram",
		func(ctx context.Context, subLogger log.Logger, cfgMgr *viper.Viper, cfgMgrMu *sync.RWMutex, vw *view.View, extra interface{}, params map[string]interface{}) ([]eh.Command, error) {
			return []eh.Command{
				&domain.CreateRedfishResource{
					ResourceURI: vw.GetURI(),
					Type:        "#DellPower.v1_0_0.DellPowerTrend",
					Context:     "/redfish/v1/$metadata#Power.Power",
					Privileges: map[string]interface{}{
						"GET": []string{"Login"},
					},
					Properties: map[string]interface{}{
						"Name":                     "System Power History",
						"MemberId":                 "PowerHistogram",
						"HistoryMaxWattsTime@meta": vw.Meta(view.GETProperty("max_watts_time"), view.GETModel("default")),
						"HistoryMaxWatts@meta":     vw.Meta(view.GETProperty("max_watts"), view.GETModel("default")),
						"HistoryMinWattsTime@meta": vw.Meta(view.GETProperty("min_watts_time"), view.GETModel("default")),
						"HistoryMinWatts@meta":     vw.Meta(view.GETProperty("min_watts"), view.GETModel("default")),
						"HistoryAverageWatts@meta": vw.Meta(view.GETProperty("average_watts"), view.GETModel("default")),
					}},
			}, nil
		})

	s.RegisterAggregateFunction("power_control",
		func(ctx context.Context, subLogger log.Logger, cfgMgr *viper.Viper, cfgMgrMu *sync.RWMutex, vw *view.View, extra interface{}, params map[string]interface{}) ([]eh.Command, error) {
			return []eh.Command{
				&domain.CreateRedfishResource{
					ResourceURI: vw.GetURI(),
					Type:        "#Power.v1_0_0.PowerControl",
					Context:     "/redfish/v1/$metadata#Power.v1_0_0.PowerControl",
					Privileges: map[string]interface{}{
						"GET":   []string{"Login"},
						"PATCH": []string{"ConfigureManager"},
					},
					Properties: map[string]interface{}{
						"Name":                     "System Power Control",
						"MemberId":                 "PowerControl",
						"PowerAvailableWatts@meta": vw.Meta(view.PropGET("headroom_watts")),
						"PowerCapacityWatts@meta":  vw.Meta(view.PropGET("capacity_watts")), //System.Chassis.1#ChassisPower.1#SystemInputMaxPowerCapacity
						"PowerConsumedWatts@meta":  vw.Meta(view.PropGET("consumed_watts")),

						"Oem": map[string]interface{}{
							"EnergyConsumptionStartTime@meta": vw.Meta(view.PropGET("energy_consumption_start_time")),
							"EnergyConsumptionkWh@meta":       vw.Meta(view.PropGET("energy_consumption_kwh")),
							"HeadroomWatts@meta":              vw.Meta(view.PropGET("headroom_watts")),
							"MaxPeakWatts@meta":               vw.Meta(view.PropGET("max_peak_watts")),
							"MaxPeakWattsTime@meta":           vw.Meta(view.PropGET("max_peak_watts_time")),
							"MinPeakWatts@meta":               vw.Meta(view.PropGET("min_peak_watts")),
							"MinPeakWattsTime@meta":           vw.Meta(view.PropGET("min_peak_watts_time")),
							"PeakHeadroomWatts@meta":          vw.Meta(view.PropGET("peak_headroom_watts")),
						},
						"PowerLimit": map[string]interface{}{
							"LimitInWatts@meta": vw.Meta(view.PropGET("limit_in_watts")),
						},
						"PowerMetrics": map[string]interface{}{
							"AverageConsumedWatts": 0,
							"IntervalInMin":        0,
							"MaxConsumedWatts":     0,
							"MinConsumedWatts":     0,
						},
						"RelatedItem@meta":             vw.Meta(view.GETProperty("power_related_items"), view.GETFormatter("formatOdataList"), view.GETModel("default")),
						"RelatedItem@odata.count@meta": vw.Meta(view.GETProperty("power_related_items"), view.GETFormatter("count"), view.GETModel("default")),
					},
				}}, nil
		})

	s.RegisterAggregateFunction("psu_slot",
		func(ctx context.Context, subLogger log.Logger, cfgMgr *viper.Viper, cfgMgrMu *sync.RWMutex, vw *view.View, extra interface{}, params map[string]interface{}) ([]eh.Command, error) {
			return []eh.Command{
				&domain.CreateRedfishResource{
					ResourceURI: vw.GetURI(),
					Type:        "#Power.v1_0_0.PowerSupply",
					Context:     "/redfish/v1/$metadata#Power.v1_0_0.PowerSupply",
					Privileges: map[string]interface{}{
						"GET":   []string{"Login"},
						"PATCH": []string{"ConfigureManager"},
					},
					Properties: map[string]interface{}{
						"Name@meta":               vw.Meta(view.PropGET("name")),
						"MemberId@meta":           vw.Meta(view.PropGET("unique_name")),
						"PowerCapacityWatts@meta": vw.Meta(view.PropGET("capacity_watts")),
						"LineInputVoltage@meta":   vw.Meta(view.GETProperty("line_input_voltage"), view.GETModel("default")),
						"FirmwareVersion@meta":    vw.Meta(view.PropGET("firmware_version")),

						"Status": map[string]interface{}{
							"HealthRollup@meta": vw.Meta(view.PropGET("obj_status")),
							"State@meta":        vw.Meta(view.PropGET("state")),
							"Health@meta":       vw.Meta(view.PropGET("obj_status")),
						},

						"Oem": map[string]interface{}{
							"Dell": map[string]interface{}{
								"@odata.type":       "#DellPower.v1_0_0.DellPowerSupply",
								"ComponentID@meta":  vw.Meta(view.PropGET("component_id")),
								"InputCurrent@meta": vw.Meta(view.GETProperty("input_current"), view.GETModel("default")),
								"Attributes@meta":   vw.Meta(view.GETProperty("attributes"), view.GETFormatter("attributeFormatter"), view.GETModel("default"), view.PropPATCH("attributes", "ar_dump")),
							},
						},
						// this should be a link using getformatter
						"Redundancy":             []interface{}{},
						"Redundancy@odata.count": 0,
					},
				}}, nil
		})

}
