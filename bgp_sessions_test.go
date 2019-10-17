package packngo

import (
	"testing"
)

func TestAccBGPSession(t *testing.T) {
	skipUnlessAcceptanceTestsAllowed(t)

	c, projectID, teardown := setupWithProject(t)
	defer teardown()

	configRequest := CreateBGPConfigRequest{
		DeploymentType: "local",
		Asn:            65000,
		Md5:            "c3RhY2twb2ludDIwMTgK",
	}

	_, err := c.BGPConfig.Create(projectID, configRequest)
	if err != nil {
		t.Fatal(err)
	}

	hn := randString8()
	cr := DeviceCreateRequest{
		Hostname:     hn,
		Facility:     []string{testFacility()},
		Plan:         "baremetal_0",
		ProjectID:    projectID,
		BillingCycle: "hourly",
		OS:           "ubuntu_18_04",
	}

	d, _, err := c.Devices.Create(&cr)
	if err != nil {
		t.Fatal(err)
	}

	d, err = waitDeviceActive(d.ID, c)
	if err != nil {
		t.Fatal(err)
	}

	defer deleteDevice(t, c, d.ID)

	aTrue := true

	bgpSession, _, err := c.BGPSessions.Create(d.ID,
		CreateBGPSessionRequest{
			AddressFamily: "ipv4",
			DefaultRoute:  &aTrue})
	if err != nil {
		t.Fatal(err)
	}

	sessionID := bgpSession.ID

	sessions, _, err := c.Devices.ListBGPSessions(d.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	var check *BGPSession
	for _, s := range sessions {
		if s.ID == sessionID {
			check = &s
			break
		}
	}
	if check == nil {
		t.Fatal("BGP Session not returned.")
	}

	if !*(check.DefaultRoute) {
		t.Fatal("BGP Session Default Route should have been set.")
	}

	cs, _, err := c.BGPConfig.Get(projectID,
		&GetOptions{Includes: []string{"sessions"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(cs.Sessions) != 1 {
		t.Fatal("only one Session should be listed in project BGP conf")
	}
	if cs.Sessions[0].ID != sessionID {
		t.Fatal("BGP Session ID mismatch")
	}
	sessions, _, err = c.Projects.ListBGPSessions(projectID, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range sessions {
		if s.ID == sessionID {
			check = &s
			break
		}
	}

	if check == nil {
		t.Fatal("BGP Session not returned.")
	}

	bgpSession, _, err = c.BGPSessions.DisableDefaultRoute(sessionID)
	if bgpSession == nil {
		t.Fatal("Disabling default route failed")
	}

	if *bgpSession.DefaultRoute {
		t.Fatal("Default route feature should have been disabled")
	}

	bgpSession, _, err = c.BGPSessions.EnableDefaultRoute(sessionID)
	if bgpSession == nil {
		t.Fatal("Enabling default route failed")
	}

	if !*bgpSession.DefaultRoute {
		t.Fatal("Default route feature should have been re-enabled")
	}

	_, err = c.BGPSessions.Delete(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	session, _, err := c.BGPSessions.Get(sessionID, nil)
	if session != nil {
		t.Fatal("Session not deleted")
	}
	if err == nil {
		t.Fatal("Session not deleted")
	}

}
