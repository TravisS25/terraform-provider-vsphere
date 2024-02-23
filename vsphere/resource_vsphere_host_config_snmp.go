package vsphere

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostservicestate"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	esxissh "github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/ssh"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/crypto/ssh"
)

func resourceVSphereHostConfigSNMP() *schema.Resource {
	return &schema.Resource{
		Create:        resourceVSphereHostConfigSNMPCreate,
		Read:          resourceVSphereHostConfigSNMPRead,
		Delete:        resourceVSphereHostConfigSNMPDelete,
		Update:        resourceVSphereHostConfigSNMPUpdate,
		CustomizeDiff: resourceVSphereHostConfigSNMPCustomDiff,
		Importer: &schema.ResourceImporter{
			State: resourceVSphereHostConfigSNMPImport,
		},

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "ID of the host system to set up snmp",
				ExactlyOneOf: []string{"hostname"},
			},
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Hostname of the host system to set up snmp",
			},
			// The only reason to have to set a user and password per host is as of this writing,
			// it does not appear that the "methods.ReconfigureSnmpAgent" function in the govmomi
			// library allows us to set a remoteuser with auth and privacy.  It appears that the "Option"
			// slice in the "types.HostSnmpConfigSpec" struct that allows arbitrary key/value pairs
			// will allow us to set things like the auth/privacy, loglevel etc. but does not have
			// a "remoteusers" option
			//
			// Right now we are using the user and password as creds to literally ssh into each host
			// and set the users through the cli command "esxcli".  If this changes or there is a better
			// way of doing this, this api should be updated to not have to use these creds
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User of host.  Only required if using snmp v3",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Password of host.  Only required if using snmp v3",
			},
			"known_hosts_path": {
				Type:     schema.TypeString,
				Optional: true,
				Description: `File path to 'known_hosts' file that must contain the hostname of esxi host.
				This is used to verify a host against their current public ssh key.  Must be full path
				`,
			},
			"ssh_port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     22,
				Description: "Port to connect to esxi host for ssh",
			},
			"ssh_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     8,
				Description: "Number in seconds it should take to establish connection before timing out",
			},
			"engine_id": {
				Type:        schema.TypeString,
				Description: "Sets SNMPv3 engine id",
				Optional:    true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(10, 32),
					validation.StringMatch(
						regexp.MustCompile("^[0-9a-fA-F]+$"),
						"Must be hexadecimal characters",
					),
				),
			},
			"authentication_protocol": {
				Type:         schema.TypeString,
				Description:  "Protocol used ensure the identity of users of SNMP v3",
				Optional:     true,
				Default:      "none",
				ValidateFunc: validation.StringInSlice([]string{"none", "SHA1"}, false),
			},
			"privacy_protocol": {
				Type:         schema.TypeString,
				Description:  "Protocol used to allow encryption of SNMP v3 messages",
				Optional:     true,
				Default:      "none",
				ValidateFunc: validation.StringInSlice([]string{"none", "AES128"}, false),
			},
			"log_level": {
				Type:         schema.TypeString,
				Description:  "Log level the host snmp agent will output",
				Optional:     true,
				Default:      "warning",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "warning", "error"}, false),
			},
			"remote_user": {
				Type:        schema.TypeSet,
				Description: "Set of users to use for auth against snmp agent",
				Optional:    true,
				MaxItems:    5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of user",
						},
						"authentication_password": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							Description:  "Password to use to auth user",
							ValidateFunc: validation.StringMatch(regexp.MustCompile(".{8,}"), "Must be at least 8 characters"),
						},
						"privacy_secret": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							Description:  "Secret to use for encryption of messages",
							ValidateFunc: validation.StringMatch(regexp.MustCompile(".{16}"), "Must be exactly 16 characters"),
						},
					},
				},
			},
			"snmp_port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     161,
				Description: "Port for the agent listen on",
			},
			"read_only_communities": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Communities that are read only.  Only valid for version 1 and 2",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"trap_target": {
				Type:        schema.TypeSet,
				Description: "Targets to send snmp message",
				Optional:    true,
				MaxItems:    3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hostname": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Hostname of receiver for notifications from host",
						},
						"port": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Port of receiver for notifications from host",
						},
						"community": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Community of receiver for notifications from host",
						},
					},
				},
			},
		},
	}
}

func resourceVSphereHostConfigSNMPCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	host, hr, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return fmt.Errorf("error retrieving host on snmp create: %s", err)
	}

	if err = hostConfigSNMPUpdate(client, d, host, true); err != nil {
		return fmt.Errorf("error updating snmp on host '%s': %s", host.Name(), err)
	}

	d.SetId(hr.Value)
	return nil
}

func resourceVSphereHostConfigSNMPRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	host, _, err := hostsystem.CheckIfHostnameOrID(client, d.Id())
	if err != nil {
		return fmt.Errorf("error retrieving host on snmp read: %s", err)
	}

	return hostConfigSNMPRead(client, d, host)
}

func resourceVSphereHostConfigSNMPUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	host, _, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return fmt.Errorf("error retrieving host on snmp update: %s", err)
	}

	return hostConfigSNMPUpdate(client, d, host, true)
}

func resourceVSphereHostConfigSNMPDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	host, _, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return fmt.Errorf("error retrieving host on snmp delete: %s", err)
	}

	return hostConfigSNMPUpdate(client, d, host, false)
}

func resourceVSphereHostConfigSNMPImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Client).vimClient
	_, hr, err := hostsystem.CheckIfHostnameOrID(client, d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving host on snmp import: %s", err)
	}

	user := os.Getenv("TF_VAR_VSPHERE_ESXI_USER")
	password := os.Getenv("TF_VAR_VSPHERE_ESXI_PASSWORD")
	sshPort := os.Getenv("TF_VAR_VSPHERE_ESXI_SSH_PORT")
	sshTimeout := os.Getenv("TF_VAR_VSPHERE_ESXI_SSH_TIMEOUT")

	if user == "" {
		return nil, fmt.Errorf("must set 'TF_VAR_VSPHERE_ESXI_USER' env variable")
	}
	if password == "" {
		return nil, fmt.Errorf("must set 'TF_VAR_VSPHERE_ESXI_PASSWORD' env variable")
	}
	if sshPort == "" {
		sshPort = "22"
	}
	if sshTimeout == "" {
		sshTimeout = "8"
	}

	port, err := strconv.Atoi(sshPort)
	if err != nil {
		return nil, fmt.Errorf("'ssh_port' must be integer")
	}

	timeout, err := strconv.Atoi(sshTimeout)
	if err != nil {
		return nil, fmt.Errorf("'ssh_timeout' must be integer")
	}

	d.SetId(hr.Value)
	d.Set(hr.IDName, hr.Value)
	d.Set("user", user)
	d.Set("password", password)
	d.Set("ssh_port", port)
	d.Set("ssh_timeout", timeout)

	return []*schema.ResourceData{d}, nil
}

func resourceVSphereHostConfigSNMPCustomDiff(ctx context.Context, rd *schema.ResourceDiff, meta interface{}) error {
	users := rd.Get("remote_user").(*schema.Set).List()
	ap := rd.Get("authentication_protocol").(string)
	pp := rd.Get("privacy_protocol").(string)
	engineID := rd.Get("engine_id").(string)
	knownHostsPath := rd.Get("known_hosts_path").(string)

	if knownHostsPath != "" {
		_, err := os.Stat(knownHostsPath)
		if err != nil {
			return fmt.Errorf("error with 'known_hosts_path' attribute: %s", err)
		}

		var tfID string
		client := meta.(*Client).vimClient

		if rd.Get("hostname").(string) != "" {
			tfID = rd.Get("hostname").(string)
		} else {
			tfID = rd.Get("host_system_id").(string)
		}

		host, _, err := hostsystem.CheckIfHostnameOrID(client, tfID)
		if err != nil {
			return fmt.Errorf("error retrieving host during custom diff: %s", err)
		}

		if _, err = esxissh.GetKnownHostsOutput(knownHostsPath, host.Name()); err != nil {
			return fmt.Errorf("error retrieving output to verify host '%s': %s", host.Name(), err)
		}
	}

	if len(users) > 0 {
		if engineID == "" {
			return fmt.Errorf("'engine_id' required if setting any 'remote_user' resource")
		}
	}

	for _, u := range users {
		user := u.(map[string]interface{})

		if user["authentication_password"].(string) != "" && ap == "none" {
			return fmt.Errorf("'authentication_protocol' must be set if any 'remote_user' resource has 'authentication_password' set")
		}
		if user["privacy_secret"].(string) != "" && pp == "none" {
			return fmt.Errorf("'privacy_protocol' must be set if any 'remote_user' resource has 'privacy_secret' set")
		}
	}

	return nil
}

func hostConfigSNMPRead(client *govmomi.Client, d *schema.ResourceData, host *object.HostSystem) error {
	err := startSSHServiceForSNMP(client, host)
	if err != nil {
		return fmt.Errorf("error starting ssh service on host '%s': %s", host.Name(), err)
	}

	cb := ssh.InsecureIgnoreHostKey()

	if d.Get("known_hosts_path").(string) != "" {
		cb = esxissh.GetDefaultHostKeyCallback(d.Get("known_hosts_path").(string))
	}

	result, err := esxissh.RunCommand(
		"/bin/esxcli system snmp get",
		host.Name(),
		d.Get("ssh_port").(int),
		esxissh.GetDefaultClientConfig(
			d.Get("user").(string),
			d.Get("password").(string),
			d.Get("ssh_timeout").(int),
			cb,
		),
	)
	if err != nil {
		return fmt.Errorf("error executing command to gather snmp settings host '%s': %s", host.Name(), err)
	}

	outBuf := bytes.Buffer{}
	if _, err = outBuf.ReadFrom(result); err != nil {
		return fmt.Errorf("error reading output result on host '%s': %s", host.Name(), err)
	}

	for {
		line, err := outBuf.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("error reading configuration output on host '%s': %s", host.Name(), err)
			}

			break
		}

		lineArr := strings.Split(line, ":")
		key := strings.TrimSpace(strings.ToLower(lineArr[0]))
		value := strings.TrimSpace(lineArr[1])

		switch key {
		case "authentication":
			d.Set("authentication_protocol", value)
		case "communities":
			if value != "" {
				communities := strings.Split(value, ",")
				d.Set("read_only_communities", communities)
			} else {
				d.Set("read_only_communities", []interface{}{})
			}
		case "engineid":
			d.Set("engine_id", value)
		case "loglevel":
			d.Set("log_level", value)
		case "port":
			port, _ := strconv.Atoi(value)
			d.Set("snmp_port", port)
		case "privacy":
			d.Set("privacy_protocol", value)
		case "targets":
			if value != "" {
				targets := strings.Split(value, ",")
				trapTargets := make([]map[string]interface{}, 0, len(targets))

				for _, target := range targets {
					hp := strings.Split(target, "@")
					hostname := strings.TrimSpace(hp[0])
					pc := strings.Split(hp[1], " ")

					port, _ := strconv.Atoi(strings.TrimSpace(pc[0]))
					community := strings.TrimSpace(pc[1])

					trapTargets = append(trapTargets, map[string]interface{}{
						"hostname":  hostname,
						"port":      int32(port),
						"community": community,
					})
				}

				d.Set("trap_target", trapTargets)
			} else {
				d.Set("trap_target", []interface{}{})
			}
		}
	}

	return nil
}

func hostConfigSNMPUpdate(client *govmomi.Client, d *schema.ResourceData, host *object.HostSystem, isUpdate bool) error {
	moHost, err := hostsystem.Properties(host)
	if err != nil {
		return fmt.Errorf(
			"error retrieving host properties on host '%s': %s",
			host.Name(),
			err,
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	users := d.Get("remote_user").(*schema.Set).List()

	// This flag is used to indicate if we should shut down the ssh service after completing our snmp commands
	// This will only be set to true if the ssh service is currently NOT running
	// shutdownSSH := false
	// sshPolicy := ""
	enabled := true

	if len(users) > 0 {
		if err = startSSHServiceForSNMP(client, host); err != nil {
			return fmt.Errorf("error starting ssh service on host '%s': %s", host.Name(), err)
		}
	}

	cb := ssh.InsecureIgnoreHostKey()

	if d.Get("known_hosts_path").(string) != "" {
		cb = esxissh.GetDefaultHostKeyCallback(d.Get("known_hosts_path").(string))
	}

	if isUpdate {
		var roc []string
		communityList := d.Get("read_only_communities").(*schema.Set).List()

		if len(communityList) == 0 {
			roc = []string{"reset"}
		} else {
			roc = make([]string, 0, len(communityList))

			for _, item := range communityList {
				roc = append(roc, item.(string))
			}
		}

		var tts []types.HostSnmpDestination

		ttList := d.Get("trap_target").(*schema.Set).List()

		if len(ttList) == 0 {
			tts = []types.HostSnmpDestination{}
		} else {
			tts = make([]types.HostSnmpDestination, 0, len(ttList))
			for _, item := range ttList {
				tt := item.(map[string]interface{})
				tts = append(tts, types.HostSnmpDestination{
					HostName:  tt["hostname"].(string),
					Port:      int32(tt["port"].(int)),
					Community: tt["community"].(string),
				})
			}
		}

		options := []types.KeyValue{
			{
				Key:   "loglevel",
				Value: d.Get("log_level").(string),
			},
			{
				Key:   "authentication",
				Value: d.Get("authentication_protocol").(string),
			},
			{
				Key:   "privacy",
				Value: d.Get("privacy_protocol").(string),
			},
		}

		if len(users) > 0 {
			options = append(options, types.KeyValue{
				Key:   "engineid",
				Value: d.Get("engine_id").(string),
			})
		}

		if _, err = methods.ReconfigureSnmpAgent(
			ctx,
			client,
			&types.ReconfigureSnmpAgent{
				This: *moHost.ConfigManager.SnmpSystem,
				Spec: types.HostSnmpConfigSpec{
					Enabled:             &enabled,
					Port:                int32(d.Get("snmp_port").(int)),
					ReadOnlyCommunities: roc,
					TrapTargets:         tts,
					Option:              options,
				},
			},
		); err != nil {
			return fmt.Errorf("error reconfiguring snmp agent on host '%s': %s", host.Name(), err)
		}

		if len(users) > 0 {
			baseHashCmd := "/bin/esxcli system snmp hash"
			setUserStr := ""

			ap := d.Get("authentication_protocol").(string)
			pp := d.Get("privacy_protocol").(string)
			eID := d.Get("engine_id").(string)
			outBuf := bytes.Buffer{}

			for idx, u := range users {
				esxHashCmd := baseHashCmd

				user := u.(map[string]interface{})
				setUserStr += user["name"].(string) + "/"

				if user["authentication_password"] != "" {
					esxHashCmd += " --auth-hash " + user["authentication_password"].(string)
				}
				if user["privacy_secret"] != "" {
					esxHashCmd += " --priv-hash " + user["privacy_secret"].(string)
				}

				ah := "-"
				ph := "-"

				if esxHashCmd != baseHashCmd {
					esxHashCmd += " --raw-secret"

					result, err := esxissh.RunCommand(
						esxHashCmd,
						host.Name(),
						d.Get("ssh_port").(int),
						esxissh.GetDefaultClientConfig(
							d.Get("user").(string),
							d.Get("password").(string),
							d.Get("ssh_timeout").(int),
							cb,
						),
					)
					if err != nil {
						return fmt.Errorf("error executing hash command on host '%s': %s", host.Name(), err)
					}

					if _, err = outBuf.ReadFrom(result); err != nil {
						return fmt.Errorf("error reading stdout of esxcli hash command on host '%s': %s", host.Name(), err)
					}

					authLine, err := outBuf.ReadString('\n')
					if err != nil {
						return fmt.Errorf("error reading stdout of esxcli hash command on host '%s': %s", host.Name(), err)
					}

					authHash := strings.TrimSpace(strings.Split(authLine, ":")[1])

					privLine, err := outBuf.ReadString('\n')
					if err != nil {
						return fmt.Errorf("error reading stdout of esxcli hash command on host '%s': %s", host.Name(), err)
					}

					privHash := strings.TrimSpace(strings.Split(privLine, ":")[1])

					outBuf.Reset()

					if authHash != "" {
						ah = authHash
					}
					if privHash != "" {
						ph = privHash
					}
				}

				setUserStr += ap + "/" + ah + "/" + pp + "/" + ph + "/" + eID

				if idx != len(users)-1 {
					setUserStr += ","
				}
			}

			if _, err = esxissh.RunCommand(
				fmt.Sprintf("/bin/esxcli system snmp set --remote-users %s", setUserStr),
				host.Name(),
				d.Get("ssh_port").(int),
				esxissh.GetDefaultClientConfig(
					d.Get("user").(string),
					d.Get("password").(string),
					d.Get("ssh_timeout").(int),
					cb,
				),
			); err != nil {
				return fmt.Errorf("error executing esxcli user command on host '%s': %s", host.Name(), err)
			}
		}
	} else {
		enabled = false

		if _, err = methods.ReconfigureSnmpAgent(
			ctx,
			client,
			&types.ReconfigureSnmpAgent{
				This: *moHost.ConfigManager.SnmpSystem,
				Spec: types.HostSnmpConfigSpec{
					Enabled:             &enabled,
					ReadOnlyCommunities: []string{"reset"},
					Port:                161,
					Option: []types.KeyValue{
						{
							Key:   "loglevel",
							Value: "warning",
						},
						{
							Key:   "privacy",
							Value: "none",
						},
						{
							Key:   "authentication",
							Value: "none",
						},
					},
				},
			},
		); err != nil {
			return fmt.Errorf("error deleting snmp settings on host '%s': %s", host.Name(), err)
		}

		if _, err = esxissh.RunCommand(
			"/bin/esxcli system snmp set --targets 'reset'",
			host.Name(),
			d.Get("ssh_port").(int),
			esxissh.GetDefaultClientConfig(
				d.Get("user").(string),
				d.Get("password").(string),
				d.Get("ssh_timeout").(int),
				cb,
			),
		); err != nil {
			return fmt.Errorf("error executing esxcli targets reset command on host '%s': %s", host.Name(), err)
		}

		if _, err = esxissh.RunCommand(
			"/bin/esxcli system snmp set --remote-users 'reset'",
			host.Name(),
			d.Get("ssh_port").(int),
			esxissh.GetDefaultClientConfig(
				d.Get("user").(string),
				d.Get("password").(string),
				d.Get("ssh_timeout").(int),
				cb,
			),
		); err != nil {
			return fmt.Errorf("error executing esxcli remote user reset command on host '%s': %s", host.Name(), err)
		}
	}

	return nil
}

func startSSHServiceForSNMP(client *govmomi.Client, host *object.HostSystem) error {
	hostServices, err := hostservicestate.GetHostServies(client, host, defaultAPITimeout)
	if err != nil {
		return fmt.Errorf("error retrieving host services on snmp update on host '%s': %s", host.Name(), err)
	}

	for _, srv := range hostServices {
		if srv.Key == string(hostservicestate.HostServiceKeySSH) && !srv.Running {
			if err = hostservicestate.SetServiceState(
				client,
				host,
				map[string]interface{}{
					"key":    srv.Key,
					"policy": srv.Policy,
				},
				defaultAPITimeout,
				true,
			); err != nil {
				return fmt.Errorf("error starting ssh service while updating snmp on host '%s': %s", host.Name(), err)
			}
		}
	}

	return nil
}
