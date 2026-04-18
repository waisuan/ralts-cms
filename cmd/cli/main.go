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
	appendMode := flag.Bool("append", false, "Append to existing data instead of wiping tables first (machine/user only)")
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
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		deps.Shutdown(ctx)
	}()

	switch *entityType {
	case "machine":
		if !*appendMode {
			deps.PostgresClient.Exec(context.Background(), "DELETE FROM machines")
			deps.PostgresClient.Exec(context.Background(), "DELETE FROM maintenance")
		}
		createMachines(deps, *entityCount)
	case "user":
		if !*appendMode {
			deps.PostgresClient.Exec(context.Background(), "DELETE FROM users")
		}
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

	// Realistic Malaysian production data for seed generation
	customers := []string{
		"", "National Kidney Foundation", "Calibration Equipment",
		"UITM Pulau Pinang", "Hospital Seberang Jaya", "Hospital Mersing",
		"Hospital Segamat", "Hospital Segamat ED", "Hospital Kemaman (ED)",
		"MTSB from KK Presint 9 Putrajaya", "INSTITUT PERUBATAN RESPIRATORI",
		"Klinik Kesihatan Anika, Klang", "Klinik Kesihatan Kuala Lipis",
		"Klinik Kesihatan Benta", "Klinik Kesihatan Padang Tengku",
		"Klinik Kesihatan Sg Koyan", "Klinik Kesihatan Seremban",
		"Klinik Kesihatan Cheneh", "Klinik Kesihatan Air Putih",
		"Klinik Kesihatan Batu 2 1/2", "Klinik Kesihatan Kerteh",
		"Klinik Kesihatan Penambang", "Klinik Kesihatan Buloh Kasap",
		"Klinik Kesihatan Lundang Paku", "Klinik Kesihatan Ajil",
		"KLINIK KESIHATAN GUNONG", "KLINIK KESIHATAN BACHOK",
		"KLINIK KESIHATAN GUA MUSANG", "KLINIK KESIHATAN LABOK",
		"KLINIK KESIHATAN BERIS KUBOR BESAR", "KLINIK KESIHATAN BANGGOL JUDAH",
		"Klinik Kesihatan Tengkawang", "Klinik Kesihatan Chukai",
		"Klinik Kesihatan Kuala Kemaman",
	}
	states := []string{
		"", "Selangor", "Kuala Lumpur", "Pulau Pinang", "Johor", "Perak",
		"Kedah", "Negeri Sembilan", "Melaka", "Pahang", "Terengganu",
		"Kelantan", "Wilayah Persekutuan", "Sabah", "Sarawak",
	}
	accountTypes := []string{"", "Premium", "Standard", "Basic", "Enterprise"}
	models := []string{"", "X100", "Y200", "Z300", "A400", "B500", "C600"}
	statuses := []string{
		"", "Placement", "Asset", "Asset ( In Use )",
		"Send For Calibration", "Done Calibration", "Reagent Rental",
	}
	brands := []string{"", "Sysmex", "Roche", "Abbott", "Beckman Coulter", "Siemens", "Bio-Rad"}
	districts := []string{
		"", "Klang", "Kuala Lipis", "Kemaman", "Kota Bharu", "Bachok",
		"Gua Musang", "Machang", "Segamat", "Mersing", "Seremban",
		"Wilayah Persekutuan Kuala Lumpur", "Hulu Terengganu",
		"Petaling", "Shah Alam", "Subang Jaya",
	}
	persons := []string{
		"", "Muhammad Syukri", "Norzilam Bt Othman",
		"Wan Nursyuhada Wan Hanafi", "Puan Nurhasyimah Bt Mohd Noor",
		"En. Ujang Bin Haimim", "Pn.Shakira", "Nurul Azwa Md Dali",
		"Pn Norpipah bt Abd Hamid", "Mohamad Hamdan Bin Mustafa",
		"Pn Norhayati Othman U32", "Pn Wan Robina bt Che Wan Abas",
		"En Nazri MA", "Suhailah Safar", "Nor Sufiati Awang U32(KUP)",
		"Pn Siti Nursalihah Mohd Anuar U29", "Pn Hamisah AB Hamid",
		"Jamei bin Hasan U29", "ZAWAWI BIN ABDUL RAZAK", "Hasniza Md Yusoff",
	}
	reporters := []string{
		"", "Muhammad Syukri", "Norzilam Bt Othman", "Pn.Shakira",
		"Mohamad Hamdan Bin Mustafa", "En Nazri MA", "Suhailah Safar",
	}
	notes := []string{
		"",
		"PPM service completed",
		"Reagent replacement done",
		"Calibration verified within range",
		"Pending part replacement from supplier",
		"Routine inspection completed - no issues found",
		"Escalated to regional service engineer",
		"Awaiting TNC renewal from HQ",
		"Machine relocated to new ward",
		"Software update applied v2.1.3",
		"Annual preventive maintenance completed",
	}
	attachments := []string{
		"", "service_report.pdf", "calibration_cert.pdf", "inspection_form.pdf",
		"maintenance_log.pdf", "tnc_certificate.pdf", "ppm_report.pdf",
		"work_order.pdf", "checklist.pdf",
	}

	maintenanceActions := []string{
		"PPM service completed - all checks passed",
		"Reagent and consumables replaced",
		"Calibration performed and verified",
		"Emergency repair - power supply unit replaced",
		"Software update and system configuration",
		"Corrective maintenance - sensor alignment",
		"Inspection completed - minor wear observed",
		"Full service overhaul",
		"Filter replacement and cleaning",
		"Board replacement and functional test",
	}
	maintenanceTypes := []string{"Preventive", "Corrective", "Emergency", "Inspection"}
	maintenanceTechs := []string{
		"Muhammad Syukri", "Ahmad Faizal", "Nur Aisyah", "Mohd Rizal",
		"Siti Aminah", "Hafiz Rahman", "Nurul Huda", "Azman Ismail",
	}

	serialPrefixes := []string{
		"UG-", "UD-", "UB-", "UA ", "UD ", "UB ", "P", "",
	}
	serialSuffixes := []string{"", "", "", "", "", "[R]", " [R]"}
	usedSerials := make(map[string]bool, count)

	for i := 0; i < count; i++ {
		var serialNumber string
		for {
			prefix := randomChoice(serialPrefixes)
			suffix := randomChoice(serialSuffixes)
			serialNumber = fmt.Sprintf("%s%08d%s", prefix, randomInt(10000, 99999999), suffix)
			if !usedSerials[serialNumber] {
				usedSerials[serialNumber] = true
				break
			}
		}

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

	usernames := []string{
		"syukri", "norzilam", "nursyuhada", "nurhasyimah", "ujang",
		"shakira", "azwa", "norpipah", "hamdan", "norhayati",
		"robina", "nazri", "suhailah", "sufiati", "nursalihah",
		"hamisah", "jamei", "zawawi", "hasniza", "faizal",
	}
	domains := []string{"clown-cms.com", "example.com", "test.org"}
	roles := []string{users.RoleNonAdmin, users.RoleAdmin}
	statuses := []string{users.StatusApproved, users.StatusPendingApproval}

	passwords := []string{"password123", "test123", "demo123", "user123"}

	for i := 0; i < count; i++ {
		username := usernames[i%len(usernames)]
		if i >= len(usernames) {
			username = fmt.Sprintf("%s%d", username, i/len(usernames))
		}
		domain := randomChoice(domains)
		email := fmt.Sprintf("%s@%s", username, domain)

		status := randomChoice(statuses)
		user := &users.User{
			Username: username,
			Email:    email,
			Password: randomChoice(passwords),
			Role:     randomChoice(roles),
			Status:   &status,
			Approved: status == users.StatusApproved,
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
