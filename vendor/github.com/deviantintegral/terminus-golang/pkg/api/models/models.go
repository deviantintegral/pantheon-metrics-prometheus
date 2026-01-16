// Package models defines data structures for Pantheon API resources.
package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/deviantintegral/terminus-golang/pkg/output"
)

// Site represents a Pantheon site
type Site struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Label              string                 `json:"label"`
	Created            int64                  `json:"created"`
	Framework          string                 `json:"framework"`
	Organization       string                 `json:"organization"`
	Service            string                 `json:"service_level"`
	PlanName           string                 `json:"plan_name"`
	Upstream           interface{}            `json:"upstream"` // Can be string or object
	UpstreamLabel      string                 `json:"upstream_label,omitempty"`
	PHP                string                 `json:"php_version"`
	Holder             string                 `json:"holder_type"`
	HolderID           string                 `json:"holder_id"`
	Owner              string                 `json:"owner"`
	Frozen             bool                   `json:"frozen"`
	IsFrozen           bool                   `json:"is_frozen"`
	PreferredZone      string                 `json:"preferred_zone"`
	PreferredZoneLabel string                 `json:"preferred_zone_label"`
	Info               map[string]interface{} `json:"info,omitempty"`
	// Membership information (not from API, populated during listing)
	MembershipUserID string `json:"-"`
	MembershipRole   string `json:"-"`
	MembershipIsTeam bool   `json:"-"` // True if from direct user membership, false if from org
}

// SiteListItem represents a site in list output (excludes upstream field)
type SiteListItem struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Label              string                 `json:"label"`
	Created            int64                  `json:"created"`
	Framework          string                 `json:"framework"`
	Organization       string                 `json:"organization"`
	Service            string                 `json:"service_level"`
	PlanName           string                 `json:"plan_name"`
	PHP                string                 `json:"php_version"`
	Holder             string                 `json:"holder_type"`
	HolderID           string                 `json:"holder_id"`
	Owner              string                 `json:"owner"`
	Frozen             bool                   `json:"frozen"`
	IsFrozen           bool                   `json:"is_frozen"`
	PreferredZone      string                 `json:"preferred_zone"`
	PreferredZoneLabel string                 `json:"preferred_zone_label"`
	Info               map[string]interface{} `json:"info,omitempty"`
	// Membership information (not from API, populated during listing)
	MembershipUserID string `json:"-"`
	MembershipRole   string `json:"-"`
	MembershipIsTeam bool   `json:"-"` // True if from direct user membership, false if from org
}

// ToListItem converts a Site to a SiteListItem (excludes upstream)
func (s *Site) ToListItem() *SiteListItem {
	return &SiteListItem{
		ID:                 s.ID,
		Name:               s.Name,
		Label:              s.Label,
		Created:            s.Created,
		Framework:          s.Framework,
		Organization:       s.Organization,
		Service:            s.Service,
		PlanName:           s.PlanName,
		PHP:                s.PHP,
		Holder:             s.Holder,
		HolderID:           s.HolderID,
		Owner:              s.Owner,
		Frozen:             s.Frozen,
		IsFrozen:           s.IsFrozen,
		PreferredZone:      s.PreferredZone,
		PreferredZoneLabel: s.PreferredZoneLabel,
		Info:               s.Info,
		MembershipUserID:   s.MembershipUserID,
		MembershipRole:     s.MembershipRole,
		MembershipIsTeam:   s.MembershipIsTeam,
	}
}

// Serialize implements the Serializer interface for SiteListItem.
// This method returns fields in the same order as PHP Terminus for CSV compatibility.
// Note: SiteListItem excludes upstream fields compared to Site.
func (s *SiteListItem) Serialize() []output.SerializedField {
	// Format created timestamp
	createdStr := ""
	if s.Created > 0 {
		createdStr = time.Unix(s.Created, 0).Format("2006-01-02 15:04:05")
	}

	// Determine frozen status - use Frozen field primarily, fallback to IsFrozen
	frozenStr := "false"
	if s.Frozen || s.IsFrozen {
		frozenStr = "true"
	}

	// Use PreferredZoneLabel for friendly region name (e.g., "United States" instead of "us-central1")
	region := s.PreferredZoneLabel

	// Use PlanName for friendly plan name (e.g., "Sandbox" instead of "free")
	plan := s.PlanName

	// Memberships field - format as "user_id: role" to match PHP Terminus
	// Example: "9cbc8751-968b-4d4f-9d23-909aea390817: Team"
	memberships := ""
	if s.MembershipUserID != "" && s.MembershipRole != "" {
		memberships = fmt.Sprintf("%s: %s", s.MembershipUserID, formatRole(s.MembershipRole))
	}

	return []output.SerializedField{
		{Name: "Name", Value: s.Name},
		{Name: "ID", Value: s.ID},
		{Name: "Plan", Value: plan},
		{Name: "Framework", Value: s.Framework},
		{Name: "Region", Value: region},
		{Name: "Owner", Value: s.Owner},
		{Name: "Created", Value: createdStr},
		{Name: "Memberships", Value: memberships},
		{Name: "Is Frozen?", Value: frozenStr},
	}
}

// DefaultFields implements the DefaultFielder interface for SiteListItem.
// These are the fields that should be displayed by default, matching PHP Terminus.
func (s *SiteListItem) DefaultFields() []string {
	return []string{"Name", "ID", "Plan", "Framework", "Region", "Owner", "Created", "Memberships", "Is Frozen?"}
}

// UnmarshalJSON implements custom unmarshaling to extract upstream label
func (s *Site) UnmarshalJSON(data []byte) error {
	// Define a temporary struct to unmarshal the raw data
	type SiteAlias Site
	aux := &struct {
		*SiteAlias
	}{
		SiteAlias: (*SiteAlias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Extract upstream information if upstream is an object
	if s.Upstream != nil {
		if upstreamMap, ok := s.Upstream.(map[string]interface{}); ok {
			// Extract label
			if label, ok := upstreamMap["label"].(string); ok {
				s.UpstreamLabel = label
			}

			// Format upstream as "id: url" to match PHP terminus
			var upstreamID, upstreamURL string
			// The API returns product_id, not id
			if id, ok := upstreamMap["product_id"].(string); ok {
				upstreamID = id
			}
			if url, ok := upstreamMap["url"].(string); ok {
				upstreamURL = url
			}

			if upstreamID != "" && upstreamURL != "" {
				s.Upstream = fmt.Sprintf("%s: %s", upstreamID, upstreamURL)
			}
		}
	}

	return nil
}

// Serialize implements the Serializer interface for Site.
// This method returns fields in the same order as PHP Terminus for CSV compatibility.
func (s *Site) Serialize() []output.SerializedField {
	// Format created timestamp
	createdStr := ""
	if s.Created > 0 {
		createdStr = time.Unix(s.Created, 0).Format("2006-01-02 15:04:05")
	}

	// Format upstream as string
	upstreamStr := ""
	if s.Upstream != nil {
		upstreamStr = fmt.Sprintf("%v", s.Upstream)
	}

	// Determine frozen status - use Frozen field primarily, fallback to IsFrozen
	frozenStr := "false"
	if s.Frozen || s.IsFrozen {
		frozenStr = "true"
	}

	// Use PreferredZoneLabel for friendly region name (e.g., "United States" instead of "us-central1")
	region := s.PreferredZoneLabel

	// Use PlanName for friendly plan name (e.g., "Sandbox" instead of "free")
	plan := s.PlanName

	return []output.SerializedField{
		{Name: "ID", Value: s.ID},
		{Name: "Name", Value: s.Name},
		{Name: "Label", Value: s.Label},
		{Name: "Created", Value: createdStr},
		{Name: "Framework", Value: s.Framework},
		{Name: "Organization", Value: s.Organization},
		{Name: "Plan", Value: plan},
		{Name: "Upstream", Value: upstreamStr},
		{Name: "Upstream Label", Value: s.UpstreamLabel},
		{Name: "Holder Type", Value: s.Holder},
		{Name: "Holder ID", Value: s.HolderID},
		{Name: "Owner", Value: s.Owner},
		{Name: "Region", Value: region},
		{Name: "Is Frozen?", Value: frozenStr},
	}
}

// DefaultFields implements the DefaultFielder interface for Site.
// These are the fields that should be displayed by default, matching PHP Terminus.
func (s *Site) DefaultFields() []string {
	return []string{"Name", "ID", "Plan", "Framework", "Region", "Owner", "Created", "Is Frozen?"}
}

// Environment represents a site environment
type Environment struct {
	ID                  string                 `json:"id"`
	SiteID              string                 `json:"site_id"`
	Domain              string                 `json:"domain"`
	OnServerDevelopment bool                   `json:"on_server_development"`
	Locked              bool                   `json:"locked"`
	Initialized         bool                   `json:"initialized"`
	ConnectionMode      string                 `json:"connection_mode"`
	PHP                 string                 `json:"php_version"`
	Drush               int                    `json:"drush_version"`
	TargetRef           string                 `json:"target_ref"`
	TargetCommit        string                 `json:"target_commit"`
	DiffstatCodeCommits int                    `json:"diffstat_code_commits"`
	RancherID           string                 `json:"rancher_id"`
	Info                map[string]interface{} `json:"info,omitempty"`
}

// Workflow represents an asynchronous operation
type Workflow struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"`
	Description      string                 `json:"description"`
	SiteID           string                 `json:"site_id"`
	EnvironmentID    string                 `json:"environment"`
	UserID           string                 `json:"user_id"`
	FinishedAt       float64                `json:"finished_at"`
	StartedAt        float64                `json:"started_at"`
	CreatedAt        float64                `json:"created_at"`
	Result           string                 `json:"result"`
	TotalTime        float64                `json:"total_time"`
	CurrentOperation string                 `json:"current_operation"`
	Step             int                    `json:"step"`
	FinalTask        *Task                  `json:"final_task,omitempty"`
	WaitingForTask   *Task                  `json:"waiting_for_task,omitempty"`
	Operations       []Operation            `json:"operations,omitempty"`
	Params           map[string]interface{} `json:"params,omitempty"`
	Active           bool                   `json:"active"`
	HasActiveOps     bool                   `json:"has_active_ops"`
}

// Task represents a workflow task
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Result      string                 `json:"result"`
	Messages    interface{}            `json:"messages,omitempty"`
	StartTime   float64                `json:"start_time"`
	EndTime     float64                `json:"end_time"`
	Params      map[string]interface{} `json:"params,omitempty"`
	SiteID      string                 `json:"site_id,omitempty"`
}

// Operation represents a workflow operation
type Operation struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Result      string  `json:"result"`
	Duration    float64 `json:"duration"`
}

// Message represents a workflow message
type Message struct {
	Level   string  `json:"level"`
	Message string  `json:"message"`
	Time    float64 `json:"time"`
}

// IsFinished returns true if the workflow has finished
func (w *Workflow) IsFinished() bool {
	return w.FinishedAt > 0 || w.Result != ""
}

// IsSuccessful returns true if the workflow completed successfully
func (w *Workflow) IsSuccessful() bool {
	return w.Result == "succeeded"
}

// IsFailed returns true if the workflow failed
func (w *Workflow) IsFailed() bool {
	return w.Result == "failed" || w.Result == "aborted"
}

// GetMessage returns the workflow message
func (w *Workflow) GetMessage() string {
	if w.FinalTask != nil && w.FinalTask.Messages != nil {
		// Messages can be either an array or an object, try to extract a message
		if msgs, ok := w.FinalTask.Messages.([]interface{}); ok && len(msgs) > 0 {
			if msg, ok := msgs[0].(map[string]interface{}); ok {
				if message, ok := msg["message"].(string); ok {
					return message
				}
			}
		}
	}
	return w.Description
}

// Backup represents a site backup
type Backup struct {
	ID             string `json:"id"`
	SiteID         string `json:"site_id"`
	EnvironmentID  string `json:"env_id"`
	ArchiveType    string `json:"type"`
	Timestamp      int64  `json:"timestamp"`
	FinishTime     int64  `json:"finish_time"`
	Size           int64  `json:"size"`
	Folder         string `json:"folder"`
	TTL            int64  `json:"ttl"`
	ExpiryTime     int64  `json:"expiry_time"`
	InitiatorEmail string `json:"initiator_email"`
	InitiatorName  string `json:"initiator_name"`
}

// GetDate returns the backup date as a time.Time
func (b *Backup) GetDate() time.Time {
	return time.Unix(b.Timestamp, 0)
}

// Organization represents an organization
type Organization struct {
	ID    string `json:"id"`
	Name  string `json:"name"`  // Corresponds to machine_name in API
	Label string `json:"label"` // Corresponds to name in API
}

// UnmarshalJSON implements custom unmarshaling to extract name and label from profile
func (o *Organization) UnmarshalJSON(data []byte) error {
	// Define a temporary struct to unmarshal the raw data
	type OrganizationAlias Organization
	aux := &struct {
		Profile *orgProfile `json:"profile,omitempty"`
		*OrganizationAlias
	}{
		OrganizationAlias: (*OrganizationAlias)(o),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Extract name and label from profile if present
	if aux.Profile != nil {
		o.Name = aux.Profile.MachineName
		o.Label = aux.Profile.Name
	}

	return nil
}

// Serialize implements the Serializer interface for Organization.
func (o *Organization) Serialize() []output.SerializedField {
	return []output.SerializedField{
		{Name: "ID", Value: o.ID},
		{Name: "Name", Value: o.Name},
		{Name: "Label", Value: o.Label},
	}
}

// DefaultFields implements the DefaultFielder interface for Organization.
func (o *Organization) DefaultFields() []string {
	return []string{"ID", "Name", "Label"}
}

// orgProfile represents an organization's profile (internal use only for unmarshaling)
type orgProfile struct {
	MachineName      string  `json:"machine_name"`
	ChangeServiceURL string  `json:"change_service_url"`
	Name             string  `json:"name"`
	EmailDomain      *string `json:"email_domain"`
	OrgLogoWidth     int     `json:"org_logo_width"`
	OrgLogoHeight    int     `json:"org_logo_height"`
	BaseDomain       *string `json:"base_domain"`
	BillingURL       string  `json:"billing_url"`
	TermsOfService   string  `json:"terms_of_service"`
	OrgLogo          string  `json:"org_logo"`
}

// UnmarshalJSON implements custom unmarshaling for orgProfile to handle string or int values for OrgLogoWidth and Height
func (p *orgProfile) UnmarshalJSON(data []byte) error {
	type Alias orgProfile
	aux := &struct {
		OrgLogoWidth  interface{} `json:"org_logo_width"`
		OrgLogoHeight interface{} `json:"org_logo_height"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.OrgLogoWidth.(type) {
	case float64:
		p.OrgLogoWidth = int(v)
	case string:
		if v != "" {
			val, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			p.OrgLogoWidth = val
		}
	}

	switch v := aux.OrgLogoHeight.(type) {
	case float64:
		p.OrgLogoHeight = int(v)
	case string:
		if v != "" {
			val, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			p.OrgLogoHeight = val
		}
	}

	return nil
}

// User represents a user
type User struct {
	ID        string       `json:"id"`
	Email     string       `json:"email"`
	Profile   *UserProfile `json:"profile,omitempty"`
	FirstName string       `json:"firstname"`
	LastName  string       `json:"lastname"`
}

// UnmarshalJSON implements custom unmarshaling to flatten profile data
func (u *User) UnmarshalJSON(data []byte) error {
	// Define a temporary struct to unmarshal the raw data
	type UserAlias User
	aux := &struct {
		*UserAlias
	}{
		UserAlias: (*UserAlias)(u),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Flatten profile data into top-level fields
	if u.Profile != nil {
		u.FirstName = u.Profile.FirstName
		u.LastName = u.Profile.LastName
		// Clear the profile so it doesn't appear in output
		u.Profile = nil
	}

	return nil
}

// UserProfile represents a user's profile
type UserProfile struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
}

// Domain represents a domain attached to an environment
type Domain struct {
	ID            string `json:"id"`
	Domain        string `json:"domain"`
	SiteID        string `json:"site_id"`
	EnvironmentID string `json:"environment"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	Deletable     bool   `json:"deletable"`
}

// DNSRecord represents a DNS record
type DNSRecord struct {
	Type   string `json:"type"`
	Target string `json:"target"`
	Status string `json:"status"`
}

// Upstream represents an upstream framework
type Upstream struct {
	ID           string                 `json:"id"`
	Label        string                 `json:"label"`
	MachineName  string                 `json:"machine_name"`
	Type         string                 `json:"type"`
	Framework    string                 `json:"framework"`
	Organization string                 `json:"organization_id"`
	URL          string                 `json:"url"`
	Branch       string                 `json:"branch"`
	Product      map[string]interface{} `json:"product,omitempty"`
}

// TeamMember represents a team member
type TeamMember struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Role      string `json:"role"`
}

// SSHKey represents an SSH key
type SSHKey struct {
	ID  string `json:"id"`
	Key string `json:"key"`
	Hex string `json:"hex"`
}

// Tag represents a site tag
type Tag struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	SiteID string `json:"site_id"`
	OrgID  string `json:"org_id"`
}

// PaymentMethod represents a payment method
type PaymentMethod struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
	Last4 string `json:"last4"`
}

// Plan represents a service plan
type Plan struct {
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	Label                string  `json:"label"`
	SKU                  string  `json:"sku"`
	BillingCycle         string  `json:"billing_cycle"`
	Price                float64 `json:"price"`
	MonthlyPrice         float64 `json:"monthly_price"`
	AutomatedBackups     bool    `json:"automated_backups"`
	CacheServer          bool    `json:"cache_server"`
	CustomUpstreams      bool    `json:"custom_upstreams"`
	MultidevEnvironments int     `json:"multidev_environments"`
	NewRelic             bool    `json:"new_relic"`
	SecureRuntimeAccess  bool    `json:"secure_runtime_access"`
	StorageGB            int     `json:"storage_gb"`
	SupportPlan          string  `json:"support_plan"`
}

// MachineToken represents a machine token
type MachineToken struct {
	ID         string `json:"id"`
	DeviceName string `json:"device_name"`
	Email      string `json:"email"`
	TokenName  string `json:"token_name"`
}

// Lock represents HTTP basic auth lock
type Lock struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Locked   bool   `json:"locked"`
}

// SolrConfig represents Solr configuration
type SolrConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Path    string `json:"path"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
}

// NewRelicConfig represents New Relic configuration
type NewRelicConfig struct {
	Enabled   bool   `json:"enabled"`
	AccountID string `json:"account_id"`
	APIKey    string `json:"api_key"`
}

// UpstreamUpdate represents upstream update information
type UpstreamUpdate struct {
	UpdatesAvailable bool   `json:"updates_available"`
	BehindBy         int    `json:"behind_by"`
	RemoteHead       string `json:"remote_head"`
	LocalHead        string `json:"local_head"`
}

// ConnectionInfo represents connection information for an environment
type ConnectionInfo struct {
	SFTPHost      string `json:"sftp_host"`
	SFTPPort      int    `json:"sftp_port"`
	SFTPUsername  string `json:"sftp_username"`
	SFTPCommand   string `json:"sftp_command"`
	GitHost       string `json:"git_host"`
	GitPort       int    `json:"git_port"`
	GitUsername   string `json:"git_username"`
	GitCommand    string `json:"git_command"`
	MySQLHost     string `json:"mysql_host"`
	MySQLPort     int    `json:"mysql_port"`
	MySQLUsername string `json:"mysql_username"`
	MySQLDatabase string `json:"mysql_database"`
	MySQLCommand  string `json:"mysql_command"`
	RedisHost     string `json:"redis_host"`
	RedisPort     int    `json:"redis_port"`
	RedisPassword string `json:"redis_password"`
	RedisCommand  string `json:"redis_command"`
}

// Branch represents a git branch for a site
type Branch struct {
	ID  string `json:"id"`
	SHA string `json:"sha"`
}

// UpstreamUpdateCommit represents a commit in upstream updates
type UpstreamUpdateCommit struct {
	Hash     string `json:"hash"`
	Datetime string `json:"datetime"`
	Message  string `json:"message"`
	Author   string `json:"author"`
}

// SiteOrganizationMembership represents a site's membership in an organization
type SiteOrganizationMembership struct {
	OrgID   string `json:"org_id"`
	OrgName string `json:"org_name"`
}

// Metrics represents traffic metrics for an environment
type Metrics struct {
	Timestamp     int64  `json:"-"` // Unix timestamp, used as key in JSON output
	Datetime      string `json:"datetime"`
	Visits        int64  `json:"visits"`
	PagesServed   int64  `json:"pages_served"`
	CacheHits     int64  `json:"cache_hits"`
	CacheMisses   int64  `json:"cache_misses"`
	CacheHitRatio string `json:"cache_hit_ratio"`
}

// MetricsTimeseries wraps a slice of Metrics for JSON output matching PHP terminus format
type MetricsTimeseries struct {
	Metrics []*Metrics
}

// MarshalJSON implements json.Marshaler for MetricsTimeseries
// Output format: {"timeseries": {"timestamp": {...}, ...}}
func (mt *MetricsTimeseries) MarshalJSON() ([]byte, error) {
	timeseries := make(map[string]*Metrics)
	for _, m := range mt.Metrics {
		key := strconv.FormatInt(m.Timestamp, 10)
		timeseries[key] = m
	}

	wrapper := struct {
		Timeseries map[string]*Metrics `json:"timeseries"`
	}{
		Timeseries: timeseries,
	}

	return json.Marshal(wrapper)
}

// GetItems returns the metrics slice for table output
func (mt *MetricsTimeseries) GetItems() []*Metrics {
	return mt.Metrics
}

// Serialize implements the Serializer interface for Metrics.
func (m *Metrics) Serialize() []output.SerializedField {
	// For table output, show date-only format (matching PHP terminus)
	period := m.Datetime
	if len(period) > 10 {
		period = period[:10] // Extract just the date part (2025-12-18)
	}

	return []output.SerializedField{
		{Name: "Period", Value: period},
		{Name: "Visits", Value: formatNumberWithCommas(m.Visits)},
		{Name: "Pages Served", Value: formatNumberWithCommas(m.PagesServed)},
		{Name: "Cache Hits", Value: formatNumberWithCommas(m.CacheHits)},
		{Name: "Cache Misses", Value: formatNumberWithCommas(m.CacheMisses)},
		{Name: "Cache Hit Ratio", Value: m.CacheHitRatio},
	}
}

// formatNumberWithCommas formats an int64 with comma separators (e.g., 16489 -> "16,489")
func formatNumberWithCommas(n int64) string {
	str := strconv.FormatInt(n, 10)
	if len(str) <= 3 {
		return str
	}

	// Handle negative numbers
	negative := false
	if str[0] == '-' {
		negative = true
		str = str[1:]
	}

	// Insert commas from right to left
	// Pre-allocate capacity: original length + number of commas ((len-1)/3)
	numCommas := (len(str) - 1) / 3
	result := make([]byte, 0, len(str)+numCommas)
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}

	if negative {
		return "-" + string(result)
	}
	return string(result)
}

// DefaultFields implements the DefaultFielder interface for Metrics.
func (m *Metrics) DefaultFields() []string {
	return []string{"Period", "Visits", "Pages Served", "Cache Hits", "Cache Misses", "Cache Hit Ratio"}
}

// formatRole converts API role names to friendly display names to match PHP Terminus
// Examples: "team_member" -> "Team", "organization_admin" -> "Organization Admin"
func formatRole(role string) string {
	// Handle special cases
	switch role {
	case "team_member":
		return "Team"
	case "organization_admin":
		return "Organization Admin"
	case "org_admin":
		return "Organization Admin"
	case "developer":
		return "Developer"
	case "admin":
		return "Admin"
	default:
		// For unknown roles, capitalize each word and replace underscores with spaces
		words := strings.Split(role, "_")
		for i, word := range words {
			if word != "" {
				words[i] = strings.ToUpper(word[:1]) + word[1:]
			}
		}
		return strings.Join(words, " ")
	}
}
