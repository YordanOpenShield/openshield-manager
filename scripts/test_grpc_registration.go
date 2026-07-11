//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"openshield-manager/internal/db"
	"openshield-manager/internal/models"
	"openshield-manager/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Initialize DB connection (reuses existing config)
	db.ConnectDatabase()

	// 1. Get the default organization
	var org models.Organization
	if err := db.DB.Where("slug = ?", "default").First(&org).Error; err != nil {
		log.Fatalf("Failed to find default org: %v", err)
	}
	fmt.Printf("Found organization: %s (%s)\n", org.Name, org.ID)

	// 2. Get the super admin user to use as creator
	var admin models.User
	if err := db.DB.Where("email = ?", "admin@openshield.local").First(&admin).Error; err != nil {
		log.Fatalf("Failed to find admin user: %v", err)
	}
	fmt.Printf("Found admin user: %s\n", admin.Email)

	// 3. Create a registration token for the default org
	tokenStr := "osh_reg_" + uuid.New().String()
	regToken := models.RegistrationToken{
		Token:          tokenStr,
		OrganizationID: org.ID,
		Status:         models.RegTokenActive,
		MaxUses:        5,
		UseCount:       0,
		CreatedBy:      admin.ID,
		ExpiresAt:      nil, // never expires
	}
	if err := db.DB.Create(&regToken).Error; err != nil {
		log.Fatalf("Failed to create registration token: %v", err)
	}
	fmt.Printf("\n✅ Created registration token: %s\n", tokenStr)
	fmt.Printf("   Organization: %s\n", org.Name)
	fmt.Printf("   Max uses: 5\n")

	// 4. Connect to the registration gRPC server (port 50053, no TLS)
	conn, err := grpc.NewClient(
		"127.0.0.1:50053",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to registration server: %v", err)
	}
	defer conn.Close()

	client := proto.NewManagerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 5. Register the agent WITH the registration token
	deviceID := "test-device-" + uuid.New().String()[:8]
	regResp, err := client.RegisterAgent(ctx, &proto.RegisterAgentRequest{
		DeviceId:          deviceID,
		RegistrationToken: tokenStr,
	})
	if err != nil {
		log.Fatalf("❌ Registration failed: %v", err)
	}

	fmt.Printf("\n✅ Agent registered successfully!\n")
	fmt.Printf("   Agent ID: %s\n", regResp.Id)
	fmt.Printf("   Agent Token: %s\n", regResp.Token)

	// 6. Verify the agent has the OrganizationID set
	var agent models.Agent
	if err := db.DB.Where("id = ?", regResp.Id).First(&agent).Error; err != nil {
		log.Fatalf("Failed to find agent: %v", err)
	}

	if agent.OrganizationID != nil && *agent.OrganizationID == org.ID {
		fmt.Printf("✅ OrganizationID correctly assigned: %s\n", agent.OrganizationID)
	} else {
		fmt.Printf("❌ OrganizationID NOT assigned! Got: %v, Expected: %s\n", agent.OrganizationID, org.ID)
	}

	// 7. Verify the token use count was incremented
	var updatedToken models.RegistrationToken
	if err := db.DB.Where("token = ?", tokenStr).First(&updatedToken).Error; err != nil {
		log.Fatalf("Failed to find token: %v", err)
	}
	fmt.Printf("✅ Token use count: %d/%d\n", updatedToken.UseCount, updatedToken.MaxUses)

	// 8. Register another agent with the SAME token to test multi-use
	deviceID2 := "test-device-" + uuid.New().String()[:8]
	regResp2, err := client.RegisterAgent(ctx, &proto.RegisterAgentRequest{
		DeviceId:          deviceID2,
		RegistrationToken: tokenStr,
	})
	if err != nil {
		log.Fatalf("❌ Second registration failed: %v", err)
	}

	var agent2 models.Agent
	if err := db.DB.Where("id = ?", regResp2.Id).First(&agent2).Error; err != nil {
		log.Fatalf("Failed to find agent2: %v", err)
	}

	if agent2.OrganizationID != nil && *agent2.OrganizationID == org.ID {
		fmt.Printf("\n✅ Second agent also correctly assigned to org!\n")
	} else {
		fmt.Printf("\n❌ Second agent org NOT assigned!\n")
	}

	// Check token use count
	db.DB.Where("token = ?", tokenStr).First(&updatedToken)
	fmt.Printf("✅ Token use count after 2 registrations: %d/%d\n", updatedToken.UseCount, updatedToken.MaxUses)

	fmt.Println("\n🎉 Registration flow test complete!")
}
