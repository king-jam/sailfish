package ec_manager

// this file should define the BMC Manager object golang data structures where
// we put all the data, plus the aggregate that pulls the data.  actual data
// population should happen in an impl class. ie. no dbus calls in this file

import (
	"context"

	"github.com/superchalupa/go-redfish/src/log"
	"github.com/superchalupa/go-redfish/src/ocp/model"
	domain "github.com/superchalupa/go-redfish/src/redfishresource"

	eh "github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/utils"
	ah "github.com/superchalupa/go-redfish/src/actionhandler"
)

func AddView(ctx context.Context, logger log.Logger, s *model.Model, ch eh.CommandHandler, eb eh.EventBus, ew *utils.EventWaiter) {
	ch.HandleCommand(
		ctx,
		&domain.CreateRedfishResource{
			ID:          model.GetUUID(s),
			Collection:  false,
			ResourceURI: model.GetOdataID(s),
			Type:        "#Manager.v1_0_2.Manager",
			Context:     "/redfish/v1/$metadata#Manager.Manager",
			Privileges: map[string]interface{}{
				"GET":    []string{"Login"},
				"POST":   []string{}, // cannot create sub objects
				"PUT":    []string{},
				"PATCH":  []string{"ConfigureManager"},
				"DELETE": []string{}, // can't be deleted
			},
			Properties: map[string]interface{}{
				"Id":                       s.GetProperty("unique_name").(string),
				"Name@meta":                s.Meta(model.PropGET("name")),
                // TODO: is this in AR somewhere?
				"ManagerType":              "BMC",
				"Description@meta":         s.Meta(model.PropGET("description")),
				"Model@meta":               s.Meta(model.PropGET("model")),
				"DateTime@meta":            map[string]interface{}{"GET": map[string]interface{}{"plugin": "datetime"}},
				"DateTimeLocalOffset@meta": s.Meta(model.PropGET("timezone")),
				"FirmwareVersion@meta":     s.Meta(model.PropGET("firmware_version")),
				"Links": map[string]interface{}{
					"ManagerForServers@meta": s.Meta(model.PropGET("bmc_manager_for_servers")),
					// TODO: Need standard method to count arrays
					// "ManagerForChassis@odata.count": 1,
					"ManagerForChassis@meta": s.Meta(model.PropGET("bmc_manager_for_chassis")),
					"ManagerInChassis@meta":  s.Meta(model.PropGET("in_chassis")),
				},

				"Status": map[string]interface{}{
					"HealthRollup": "OK",
					"State@meta":   s.Meta(model.PropGET("health_state")),
					"Health":       "OK",
				},

				"Redundancy@odata.count": 1,
				"Redundancy": []interface{}{
					map[string]interface{}{
						"@odata.type": "#Redundancy.v1_0_2.Redundancy",
						"Status": map[string]interface{}{
							"HealthRollup": "OK",
							"State@meta":   s.Meta(model.PropGET("redundancy_health_state")),
							"Health":       "OK",
						},
						"RedundancySet": []interface{}{
							map[string]interface{}{
								"@odata.id": "/redfish/v1/Managers/CMC.Integrated.1",
							},
							map[string]interface{}{
								"@odata.id": "/redfish/v1/Managers/CMC.Integrated.2",
							},
						},
						"Name": "ManagerRedundancy",
						"RedundancySet@odata.count": 2,
						"@odata.id":                 "/redfish/v1/Managers/CMC.Integrated.1#Redundancy",
						"@odata.context":            "/redfish/v1/$metadata#Redundancy.Redundancy",
						"Mode@meta":                 s.Meta(model.PropGET("redundancy_mode")),
						"MinNumNeeded@meta":         s.Meta(model.PropGET("redundancy_min")),
						"MaxNumSupported@meta":      s.Meta(model.PropGET("redundancy_max")),
					},
				},
				"SerialConsole": map[string]interface{}{
					"ConnectTypesSupported@odata.count": 0,
					"MaxConcurrentSessions":             0,
					"ConnectTypesSupported":             []interface{}{},
					"ServiceEnabled":                    false,
				},

				"CommandShell": map[string]interface{}{
					"ConnectTypesSupported@odata.count": 0,
					"MaxConcurrentSessions":             0,
					"ConnectTypesSupported":             []interface{}{},
					"ServiceEnabled":                    false,
				},

				"LogServices": map[string]interface{}{
					"@odata.id": model.GetOdataID(s) +  "/LogServices",
				},

				"GraphicalConsole": map[string]interface{}{
					"ConnectTypesSupported@odata.count": 0,
					"MaxConcurrentSessions":             0,
					"ConnectTypesSupported":             []interface{}{},
					"ServiceEnabled":                    false,
				},

				"Oem": map[string]interface{}{
					"@odata.type": "#DellManager.v1_0_0.DellManager",
					"OemAttributes": map[string]interface{}{
						"@odata.id": model.GetOdataID(s) +  "/Attributes",
					},
					"CertificateService": map[string]interface{}{
						"@odata.id": model.GetOdataID(s) +  "/CertificateService",
					},
				},

				"Actions": map[string]interface{}{
					"#Manager.Reset": map[string]interface{}{
						"target": model.GetOdataID(s) + "/Actions/Manager.Reset",
						"ResetType@Redfish.AllowableValues": []string{
							"GracefulRestart",
						},
					},
					"Oem": map[string]interface{}{
						"DellManager.v1_0_0#DellManager.ResetToDefaults": map[string]interface{}{
							"ResetType@Redfish.AllowableValues": []string{
								"ClearToShip",
								"Decommission",
								"ResetFactoryConfig",
								"ResetToEngineeringDefaults",
							},
							"target": model.GetOdataID(s) +  "/Actions/Oem/DellManager.ResetToDefaults",
						},
						"#Manager.ForceFailover": map[string]interface{}{
							"target": model.GetOdataID(s) +  "/Actions/Manager.ForceFailover",
						},
					},
				},

				/*
					******************************************************
					 DISABLED FOR NOW because the output is mangled by dumplogs
					******************************************************
									   "Actions": {
									        "Oem": {
									                       "OemManager.v1_0_0#OemManager.ImportSystemConfigurationPreview": {
									                  "ImportSystemConfigurationPreview@Redfish.AllowableValues": [
									                       "ImportBuffer"
									                  ],
									                  "target": model.GetOdataID(s) +  "/Actions/Oem/EID_674_Manager.ImportSystemConfigurationPreview",
									                  "ShareParameters": {
									                       "ProxySupport XXXXXX
									                            "Disabled",
									                            "EnabledProxyDefault XXXXXX
									                            "Enabled"
									                       ],
									                       "IgnoreCertificateWarning@Redfish.AllowableValues": [
									                            "Disabled",
									                            "Enabled"
									                       ],
									                       "ProxyType XXXXXX
									                            "HTTP",
									                            "SOCKS4"
									                       ],
									                       "ShareType@Redfish.AllowableValues": [
									                            "NFS",
									                            "CIFS",
									                            "HTTP",
									                            "HTTPS"
									                       ],
									                       "Target@Redfish.AllowableValues": [
									                            "ALL"
									                       ],
									                       "ShareParameters@Redfish.AllowableValues": [
									                            "IPAddress XXXXXX
									                            "ShareName",
									                            "FileName",
									                            "UserName",
									                            "Password XXXXXX
									                            "Workgroup XXXXXX
									                            "ProxyServer XXXXXX
									                            "ProxyUserName XXXXXX
									                            "ProxyPassword XXXXXX
									                            "ProxyPort XXXXXX
									                       ]
									                  }
									             },
									             "OemManager.v1_0_0#OemManager.ImportSystemConfiguration": {
									                  "HostPowerState@Redfish.AllowableValues": [
									                       "On",
									                       "Off"
									                  ],
									                  "target": model.GetOdataID(s) +  "/Actions/Oem/EID_674_Manager.ImportSystemConfiguration",
									                  "ShutdownType@Redfish.AllowableValues": [
									                       "Graceful",
									                       "Forced",
									                       "NoReboot"
									                  ],
									                  "ShareParameters": {
									                       "ProxySupport XXXXXX
									                            "Disabled",
									                            "EnabledProxyDefault XXXXXX
									                            "Enabled"
									                       ],
									                       "IgnoreCertificateWarning@Redfish.AllowableValues": [
									                            "Disabled",
									                            "Enabled"
									                       ],
									                       "ProxyType XXXXXX
									                            "HTTP",
									                            "SOCKS4"
									                       ],
									                       "ShareType@Redfish.AllowableValues": [
									                            "NFS",
									                            "CIFS",
									                            "HTTP",
									                            "HTTPS"
									                       ],
									                       "Target@Redfish.AllowableValues": [
									                            "ALL",
									                            "IDRAC",
									                            "BIOS",
									                            "NIC",
									                            "RAID"
									                       ],
									                       "ShareParameters@Redfish.AllowableValues": [
									                            "IPAddress XXXXXX
									                            "ShareName",
									                            "FileName",
									                            "UserName",
									                            "Password XXXXXX
									                            "Workgroup XXXXXX
									                            "ProxyServer XXXXXX
									                            "ProxyUserName XXXXXX
									                            "ProxyPassword XXXXXX
									                            "ProxyPort XXXXXX
									                       ]
									                  },
									                  "ImportSystemConfiguration@Redfish.AllowableValues": [
									                       "TimeToWait",
									                       "ImportBuffer"
									                  ]
									             },
									             "OemManager.v1_0_0#OemManager.ExportSystemConfiguration": {
									                  "IncludeInExport@Redfish.AllowableValues": [
									                       "Default",
									                       "IncludeReadOnly",
									                       "IncludePasswordHashValues XXXXXX
									                       "IncludeReadOnly,IncludePasswordHashValues XXXXXX
									                  ],
									                  "target": model.GetOdataID(s) +  "/Actions/Oem/EID_674_Manager.ExportSystemConfiguration",
									                  "ShareParameters": {
									                       "ProxySupport XXXXXX
									                            "Disabled",
									                            "EnabledProxyDefault XXXXXX
									                            "Enabled"
									                       ],
									                       "IgnoreCertificateWarning@Redfish.AllowableValues": [
									                            "Disabled",
									                            "Enabled"
									                       ],
									                       "ProxyType XXXXXX
									                            "HTTP",
									                            "SOCKS4"
									                       ],
									                       "ShareType@Redfish.AllowableValues": [
									                            "NFS",
									                            "CIFS",
									                            "HTTP",
									                            "HTTPS"
									                       ],
									                       "Target@Redfish.AllowableValues": [
									                            "ALL",
									                            "IDRAC",
									                            "BIOS",
									                            "NIC",
									                            "RAID"
									                       ],
									                       "ShareParameters@Redfish.AllowableValues": [
									                            "IPAddress XXXXXX
									                            "ShareName",
									                            "FileName",
									                            "UserName",
									                            "Password XXXXXX
									                            "Workgroup XXXXXX
									                            "ProxyServer XXXXXX
									                            "ProxyUserName XXXXXX
									                            "ProxyPassword XXXXXX
									                            "ProxyPort XXXXXX
									                       ]
									                  },
									                  "ExportUse@Redfish.AllowableValues": [
									                       "Default",
									                       "Clone",
									                       "Replace"
									                  ],
									                  "ExportFormat@Redfish.AllowableValues": [
									                       "XML",
									                       "JSON"
									                  ]
									             }
									        },
									   },
				*/

			}})

	ah.CreateAction(ctx, ch, eb, ew,
		logger,
		model.GetOdataID(s)+"/Actions/Manager.Reset",
		"manager.reset",
		s)

	ah.CreateAction(ctx, ch, eb, ew,
		logger,
		model.GetOdataID(s)+"/Actions/Oem/DellManager.ResetToDefaults",
		"manager.resettodefaults",
		s)

	ah.CreateAction(ctx, ch, eb, ew,
		logger,
		model.GetOdataID(s)+"/Actions/Manager.ForceFailover",
		"manager.forcefailover",
		s)
}