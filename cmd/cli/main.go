// Package main provides a CLI tool for generating test data
// including machines, users, and maintenance records for the Ralts-CMS application,
// as well as creating admin accounts and querying audit events.
package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"ralts-cms/internal/audit"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/maintenance"
	"ralts-cms/internal/users"
	"time"
)

// This is a CLI tool that accepts the following arguments:
// 1. The type of entity: machine, user, admin, events
// 2. The number of entities to create (required for machine/user)
// 3. For admin creation: --username and --password flags
// 4. For events: --minutes flag to filter by time
//
// Each entity should be unique and have a random value for the fields.
// Each entity should be saved to the database.
func main() {
	entityType := flag.String("type", "", "The type of entity to create/query (machine, user, admin, events)")
	entityCount := flag.Int("count", 0, "The number of entities to create (not required for admin/events)")
	adminUsername := flag.String("username", "", "Username for admin account (required when type=admin)")
	adminPassword := flag.String("password", "", "Password for admin account (required when type=admin)")
	// Event query flags
	eventMinutes := flag.Int("minutes", 30, "Show events from the last N minutes (default: 30)")
	eventAction := flag.String("action", "", "Filter by action type (created, updated, deleted, viewed, listed, login, logout, password_changed)")
	eventResource := flag.String("resource", "", "Filter by resource type (machine, maintenance, user, session, attachment)")
	eventLimit := flag.Int("limit", 50, "Maximum number of events to return (default: 50)")
	flag.Parse()

	if *entityType == "" {
		log.Fatal("Please provide entity type (machine, user, admin, events)")
	}

	// Validate admin-specific flags
	if *entityType == "admin" {
		if *adminUsername == "" || *adminPassword == "" {
			log.Fatal("For admin creation, both --username and --password are required")
		}
	} else if *entityType != "events" && *entityCount == 0 {
		log.Fatal("Please provide both entity type and count")
	}

	deps := deps.Initialise()

	switch *entityType {
	case "machine":
		deps.PostgresClient.Exec(context.Background(), "DELETE FROM machines")
		deps.PostgresClient.Exec(context.Background(), "DELETE FROM maintenance")
		createMachines(deps, *entityCount)
	case "user":
		deps.PostgresClient.Exec(context.Background(), "DELETE FROM users")
		createUsers(deps, *entityCount)
	case "admin":
		createAdminAccount(deps, *adminUsername, *adminPassword)
	case "events":
		queryEvents(deps, *eventMinutes, *eventAction, *eventResource, *eventLimit)
	default:
		log.Fatalf("Invalid entity type: %s. Valid options: machine, user, admin, events", *entityType)
	}
}

func createMachines(deps *deps.Dependencies, count int) {
	ctx := context.Background()

	// Sample data for random generation
	customers := []string{"", "Acme Corp", "Beta Industries", "Gamma Solutions", "Delta Systems", "Epsilon Tech", "Zeta Manufacturing", "Eta Services", "Theta Logistics", "Iota Consulting", "Kappa Solutions"}
	states := []string{"", "Selangor", "Kuala Lumpur", "Penang", "Johor", "Perak", "Kedah", "Negeri Sembilan", "Melaka", "Pahang", "Terengganu"}
	accountTypes := []string{"", "Premium", "Standard", "Basic", "Enterprise", "Professional"}
	models := []string{"", "X100", "Y200", "Z300", "A400", "B500", "C600", "D700", "E800", "F900", "G1000"}
	statuses := []string{"", "Active", "Inactive", "Maintenance", "Repair", "Operational", "Standby", "Offline", "Testing"}
	brands := []string{"", "BrandA", "BrandB", "BrandC", "BrandD", "BrandE", "BrandF", "BrandG", "BrandH", "BrandI", "BrandJ"}
	districts := []string{"", "Petaling", "Klang", "Shah Alam", "Subang Jaya", "Puchong", "Damansara", "Bangsar", "Ampang", "Cheras", "Kepong"}
	persons := []string{"", "John Doe", "Jane Smith", "Bob Johnson", "Alice Brown", "Charlie Wilson", "Diana Davis", "Edward Miller", "Fiona Garcia", "George Martinez", "Helen Rodriguez"}
	reporters := []string{"", "Tech Team A", "Tech Team B", "Maintenance Crew", "Service Department", "Engineering Team", "Operations Staff", "Support Team", "Field Engineers"}
	notes := []string{
		"",
		"Regular maintenance scheduled",
		"Equipment running smoothly",
		"Minor adjustments needed",
		"Performance optimization required",
		"Routine inspection completed",
		"Upgrade recommended",
		"Preventive maintenance due",
		"System integration pending",
		"Quality control check needed",
		"Safety inspection required",
	}
	attachments := []string{"", "report.pdf", "manual.pdf", "specs.pdf", "maintenance.pdf", "inspection.pdf", "certificate.pdf", "warranty.pdf", "guide.pdf", "checklist.pdf", "protocol.pdf"}

	maintenanceActions := []string{
		"Routine maintenance performed",
		"Filter replacement completed",
		"Oil change and lubrication",
		"Calibration and testing",
		"Component inspection",
		"Software update installed",
		"Hardware upgrade completed",
		"Safety check performed",
		"Performance optimization",
		"Preventive maintenance",
	}
	maintenanceTypes := []string{"Preventive", "Corrective", "Emergency", "Scheduled", "Breakdown", "Inspection", "Calibration", "Upgrade"}
	maintenanceTechs := []string{"Mike Johnson", "Sarah Williams", "David Brown", "Lisa Davis", "Tom Wilson", "Amy Garcia", "Chris Martinez", "Rachel Rodriguez"}

	for i := 0; i < count; i++ {
		// Generate unique serial number
		serialNumber := fmt.Sprintf("SN-%06d", i+1)

		// Create machine with random data
		machine := &machines.Machine{
			SerialNumber:    serialNumber,
			Customer:        randomChoice(customers),
			State:           randomChoice(states),
			AccountType:     randomChoice(accountTypes),
			Model:           randomChoice(models),
			Status:          randomChoice(statuses),
			Brand:           randomChoice(brands),
			District:        randomChoice(districts),
			PersonInCharge:  randomChoice(persons),
			ReportedBy:      randomChoice(reporters),
			AdditionalNotes: randomChoice(notes),
			Attachment:      randomChoice(attachments),
			TncDate:         randomDate(),
			PpmDate:         randomDate(),
		}

		// Create the machine
		err := deps.MachinesRepository.Create(ctx, machine)
		if err != nil {
			log.Printf("Failed to create machine %s: %v", serialNumber, err)
			continue
		}

		log.Printf("Created machine: %s", serialNumber)

		// Randomly create 0-5 maintenance records for this machine
		maintenanceCount := randomInt(0, count)
		for j := 0; j < maintenanceCount; j++ {
			workOrderNumber := fmt.Sprintf("WO-%s-%03d", serialNumber, j+1)

			maintenance := &maintenance.Maintenance{
				MachineSerialNumber: serialNumber,
				WorkOrderNumber:     workOrderNumber,
				WorkOrderDate:       randomDate(),
				ActionTaken:         randomChoice(maintenanceActions),
				ReportedBy:          randomChoice(maintenanceTechs),
				WorkOrderType:       randomChoice(maintenanceTypes),
				Attachment:          stringPtr(randomChoice(attachments)),
			}

			err := deps.MaintenanceRepository.Create(ctx, maintenance)
			if err != nil {
				log.Printf("Failed to create maintenance %s for machine %s: %v", workOrderNumber, serialNumber, err)
				continue
			}

			log.Printf("Created maintenance: %s for machine: %s", workOrderNumber, serialNumber)
		}
	}

	log.Printf("Successfully created %d machines with maintenance records", count)
}

func createUsers(deps *deps.Dependencies, count int) {
	ctx := context.Background()

	// Sample data for random user generation
	firstNames := []string{"John", "Jane", "Bob", "Alice", "Charlie", "Diana", "Edward", "Fiona", "George", "Helen", "Ian", "Julia", "Kevin", "Lisa", "Mike", "Nancy", "Oscar", "Patricia", "Quinn", "Rachel"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin"}
	domains := []string{"example.com", "test.org", "demo.net", "sample.co", "mock.io", "fake.com", "dummy.org", "placeholder.net"}
	roles := []string{"user", "admin", "manager", "technician", "supervisor"}
	statuses := []string{"active", "inactive", "suspended", "pending"}

	// Common passwords for testing
	passwords := []string{"password123", "test123", "demo123", "user123", "admin123", "secure123", "temp123", "default123"}

	for i := 0; i < count; i++ {
		// Generate unique email
		firstName := randomChoice(firstNames)
		lastName := randomChoice(lastNames)
		domain := randomChoice(domains)
		email := fmt.Sprintf("%s.%s.%d@%s", firstName, lastName, i+1, domain)

		// Create user with random data
		status := randomChoice(statuses)
		user := &users.User{
			Username: fmt.Sprintf("%s %s", firstName, lastName),
			Email:    email,
			Password: randomChoice(passwords),
			Role:     randomChoice(roles),
			Status:   &status,
			// Avatar is optional, so we'll leave it as nil for most users
			// Occasionally add an avatar
			Avatar: func() *string {
				if randomInt(1, 101) <= 20 { // 20% chance of having avatar
					avatar := fmt.Sprintf("avatar-%d.jpg", randomInt(1, 11))
					return &avatar
				}
				return nil
			}(),
		}

		// Create the user
		err := deps.UsersRepository.Create(ctx, user)
		if err != nil {
			log.Printf("Failed to create user %s: %v", email, err)
			continue
		}

		log.Printf("Created user: %s (%s)", email, user.Username)
	}

	log.Printf("Successfully created %d users", count)
}

// Helper functions for random data generation

func randomChoice(choices []string) string {
	if len(choices) == 0 {
		return ""
	}
	index := randomInt(0, len(choices))
	return choices[index]
}

func randomInt(minVal, maxVal int) int {
	if minVal >= maxVal {
		return minVal
	}

	// Generate random number between minVal and maxVal-1
	delta := maxVal - minVal
	randomNum, err := rand.Int(rand.Reader, big.NewInt(int64(delta)))
	if err != nil {
		// Fallback to time-based random
		return minVal + int(time.Now().UnixNano()%int64(delta))
	}

	return minVal + int(randomNum.Int64())
}

func randomDate() time.Time {
	// Generate random dates: 30% past, 40% present (within 30 days), 30% future
	now := time.Now()

	// Random number to determine date type
	randType := randomInt(1, 101)

	var targetDate time.Time

	switch {
	case randType <= 30: // Past dates (up to 2 years ago)
		daysAgo := randomInt(1, 730) // 1-730 days ago
		targetDate = now.AddDate(0, 0, -daysAgo)

	case randType <= 70: // Present dates (within 30 days)
		daysOffset := randomInt(-30, 31) // -30 to +30 days
		targetDate = now.AddDate(0, 0, daysOffset)

	default: // Future dates (up to 1 year ahead)
		daysAhead := randomInt(1, 365) // 1-365 days ahead
		targetDate = now.AddDate(0, 0, daysAhead)
	}

	// Add random time within the day
	hours := randomInt(0, 24)
	minutes := randomInt(0, 60)
	seconds := randomInt(0, 60)

	return time.Date(
		targetDate.Year(), targetDate.Month(), targetDate.Day(),
		hours, minutes, seconds, 0, time.UTC,
	)
}

// stringPtr returns a pointer to a string value, returning nil for empty strings
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// createAdminAccount creates a new admin user account with provided credentials
func createAdminAccount(deps *deps.Dependencies, username, password string) {
	ctx := context.Background()

	// Validate username
	if len(username) < 3 {
		log.Fatalf("Username must be at least 3 characters long")
	}

	// Validate password
	if len(password) < 8 {
		log.Fatalf("Password must be at least 8 characters long")
	}

	// Generate email from username (simple approach)
	email := username + "@admin.local"

	// Check if user already exists (by username)
	existingUser, err := deps.UsersRepository.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		log.Fatalf("User with username %s already exists", username)
	}

	// Create admin user
	status := users.StatusApproved // Admin accounts are automatically approved
	adminUser := &users.User{
		Username: username,
		Email:    email,
		Password: password, // Will be hashed by SetPassword method
		Role:     users.RoleAdmin,
		Approved: true, // Admin accounts are automatically approved
		Status:   &status,
		Avatar:   nil, // No avatar for CLI-created admin
	}

	// Create the admin user
	err = deps.UsersRepository.Create(ctx, adminUser)
	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("✅ Admin account created successfully!\n")
	fmt.Printf("   Username: %s\n", username)
	fmt.Printf("   Email: %s\n", email)
	fmt.Printf("   Role: ADMIN\n")
	fmt.Printf("   Status: APPROVED\n")
	fmt.Println("You can now log in to the admin panel with these credentials.")
}

// queryEvents retrieves and displays audit events from the database
func queryEvents(deps *deps.Dependencies, minutes int, action, resource string, limit int) {
	ctx := context.Background()

	// Build query options
	options := &audit.ListOptions{
		Limit:  int32(limit),
		Offset: 0,
	}

	if action != "" {
		options.Action = &action
	}
	if resource != "" {
		options.ResourceType = &resource
	}

	// Fetch events from repository
	events, err := deps.AuditRepository.List(ctx, options)
	if err != nil {
		log.Fatalf("Failed to fetch events: %v", err)
	}

	// Filter by time if minutes is specified
	cutoffTime := time.Now().Add(-time.Duration(minutes) * time.Minute)
	var filteredEvents []*audit.Event
	for _, event := range events {
		if event.CreatedAt.After(cutoffTime) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	// Display results
	fmt.Printf("\n📋 Audit Events (last %d minutes)\n", minutes)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	if len(filteredEvents) == 0 {
		fmt.Println("No events found matching the criteria.")
		return
	}

	fmt.Printf("Found %d events:\n\n", len(filteredEvents))

	for i, event := range filteredEvents {
		userID := "anonymous"
		if event.UserID != nil {
			userID = *event.UserID
		}

		fmt.Printf("[%d] %s\n", i+1, event.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("    Action:   %s\n", event.Action)
		fmt.Printf("    Resource: %s (%s)\n", event.ResourceType, event.ResourceID)
		fmt.Printf("    User ID:  %s\n", userID)

		if event.Details != nil && len(event.Details) > 0 {
			detailsJSON, _ := json.MarshalIndent(event.Details, "              ", "  ")
			fmt.Printf("    Details:  %s\n", string(detailsJSON))
		}
		fmt.Println()
	}

	// Summary
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Total: %d events\n", len(filteredEvents))

	// Count by action
	actionCounts := make(map[string]int)
	for _, event := range filteredEvents {
		actionCounts[event.Action]++
	}

	if len(actionCounts) > 1 {
		fmt.Printf("By action: ")
		first := true
		for action, count := range actionCounts {
			if !first {
				fmt.Printf(", ")
			}
			fmt.Printf("%s=%d", action, count)
			first = false
		}
		fmt.Println()
	}

	// Count by resource
	resourceCounts := make(map[string]int)
	for _, event := range filteredEvents {
		resourceCounts[event.ResourceType]++
	}

	if len(resourceCounts) > 1 {
		fmt.Printf("By resource: ")
		first := true
		for resource, count := range resourceCounts {
			if !first {
				fmt.Printf(", ")
			}
			fmt.Printf("%s=%d", resource, count)
			first = false
		}
		fmt.Println()
	}
}

// exportEventsJSON exports events as JSON to stdout (for piping to other tools)
func exportEventsJSON(events []*audit.Event) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(events); err != nil {
		log.Fatalf("Failed to encode events as JSON: %v", err)
	}
}
