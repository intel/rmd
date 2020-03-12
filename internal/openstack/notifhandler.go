// +build openstack

package openstack

import (
	"encoding/json"
	"errors"
	"strings"

	wres "github.com/intel/rmd/modules/workload"
	log "github.com/sirupsen/logrus"
)

// NovaNotification struct containing event type and payload
type NovaNotification struct {
	Payload   NovaNotificationPayload `json:"payload"`
	EventType string                  `json:"event_type"`
}

// NovaNotificationPayload event payload
type NovaNotificationPayload struct {
	Data struct {
		DisplayName string `json:"display_name"`
		Host        string `json:"host"`
		Node        string `json:"node"`
		InstanceID  string `json:"uuid,omitempty"`
		State       string `json:"state,omitempty"`
		Message     string `json:"message,omitempty"`
		Flavor      struct {
			Data struct {
				ID    string `json:"flavorid,omitempty"`
				Extra struct {
					SwiftURL string `json:"rmd:swift_workload_policy_url,omitempty"`
					GlanceID string `json:"rmd:glance_workload_policy_uuid,omitempty"`
				} `json:"extra_specs,omitempty"`
			} `json:"nova_object.data"`
		} `json:"flavor"`
	} `json:"nova_object.data"`
}

// handleNovaNotification handles notification from Openstack Nova
func handleNovaNotification(msg osloMsg) {
	notif := NovaNotification{}
	err := json.Unmarshal([]byte(msg.Payload), &notif)
	if err != nil {
		log.Errorf("Nova unmarshal error: %s", err)
		return
	}

	//notif could come from different node, we are only interested in our hostname
	if hostname != notif.Payload.Data.Host && hostname != notif.Payload.Data.Node {
		log.Infof("Skipped notification payload: %+v", notif)
		return
	}

	if msg.Type == "versioned" {
		handleVersionedNotification(notif)
	} else {
		handleUnversionedNotification(notif)
	}
}

func handleVersionedNotification(notif NovaNotification) {
	log.Debugf("Versioned notification handling")

	switch notif.EventType {
	case notificationCreateType:
		log.Debug("Notification about new instance creation (launching)")
		processCreateNotification(&notif)
	case notificationDeleteType:
		log.Debug("Notification about working instance removal (stoping)")
		processDeleteNotification(&notif)
	default:
		log.Error("Invalid notification type: ", notif.EventType)
		return
	}
}

func handleUnversionedNotification(notif NovaNotification) {
	log.Errorf("Unversioned notification handling not supported")
}

// versioned notification handling functions
func processCreateNotification(notif *NovaNotification) {
	err := validateCreateNoti(notif)
	if err != nil {
		log.Error(err)
		return
	}

	// get PID ...
	pid, err := getPIDByName(notif.Payload.Data.InstanceID)
	if err != nil {
		// failed to get PID
		log.Error("Failed to get PID for Instance (ID: ", notif.Payload.Data.InstanceID, " ): ", err)
		return
	}
	// ... and taskset for given instance
	tset, err := getPIDTaskSet(pid)
	if err != nil {
		// failed to get PIDs
		log.Error("Failed to get Taskset for PID ", pid, " : ", err)
		return
	}

	if len(tset) == 0 {
		// no CPU affinity set: we're not supporting this situation
		log.Error("Instance no pinned to any CPU core - skipping instance")
		return
	}

	// Current implementation is always trying to get policy from Swift but
	// data structures and methods prepared are ready to add also Glance support
	wrkld, err := getSwiftWorkloadPolicyByURL(notif.Payload.Data.Flavor.Data.Extra.SwiftURL)
	if err != nil {
		log.Error("Failed to get policy (", notif.Payload.Data.Flavor.Data.Extra.SwiftURL, ") for instance: ", notif.Payload.Data.InstanceID)
		return
	}

	log.Info("Instance UUID: ", notif.Payload.Data.InstanceID, " Swift policy URL: ", notif.Payload.Data.Flavor.Data.Extra.SwiftURL)

	wrkld.UUID = notif.Payload.Data.InstanceID
	wrkld.CoreIDs = tset
	wrkld.Origin = "OPENSTACK"

	log.Debug("Workload created based on policy and notification: ", wrkld)

	err = wres.Validate(wrkld)
	if err != nil {
		log.Error("Failed to validate new workload for OpenStack instance: ", err.Error())
		return
	}

	err = wres.Enforce(wrkld)
	if err != nil {
		log.Error("Failed to enforce new workload for OpenStack instance: ", err.Error())
		return
	}

	err = wres.Create(wrkld)
	if err != nil {
		log.Error("Failed to add new workload for OpenStack instance: ", err.Error())
	}

	return
}

func processDeleteNotification(notif *NovaNotification) {
	if len(notif.Payload.Data.InstanceID) == 0 {
		log.Error("Missing data in notification: InstanceID")
		return
	}

	wrkld, err := wres.GetByUUID(notif.Payload.Data.InstanceID)
	if err != nil {
		log.Error("Failed to fetch workload for given UUID (", notif.Payload.Data.InstanceID, "): ", err.Error())
		return
	}

	// workloads created by OPENSTACK should be handled only by OPENSTACK
	if wrkld.Origin == "OPENSTACK" {
		log.Debug("Origin set as OPENSTACK - Trying to release and delete workload...")
		log.Debug("Releasing...")
		err = wres.Release(&wrkld)
		if err != nil {
			log.Error("Failed to release workload for OpenStack instance: ", err.Error())
			return
		}
		log.Debug("Deleting...")
		err = wres.Delete(&wrkld)
		if err != nil {
			log.Error("Failed to delete workload for deleted OpenStack instance: ", err.Error())
			return
		}
	} else {
		log.Debug("OPENSTACK origin cannot delete non-OPENSTACK workload")
	}
	log.Debug("Deletion done")
	return
}

func validateCreateNoti(notif *NovaNotification) error {
	missing := make([]string, 0)
	if len(notif.Payload.Data.InstanceID) == 0 {
		missing = append(missing, "InstanceID")
	}
	if len(notif.Payload.Data.Flavor.Data.Extra.SwiftURL) == 0 {
		missing = append(missing, "SwiftURL")
	}

	if len(missing) > 0 {
		return errors.New("Missing data in notification: " + strings.Join(missing, ","))
	}

	return nil
}
