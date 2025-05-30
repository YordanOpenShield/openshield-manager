package managergrpc

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"time"

	"openshield-manager/internal/config"
	"openshield-manager/proto"

	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Config files whitelist for syncing
var ConfigFilesWhitelist = []string{
	"tools.yml",
}

func SyncConfigs(agentAddress string) error {
	client, err := NewAgentClient(agentAddress)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // use background context
	defer cancel()

	log.Printf("[CONFIG SYNC] Syncing configs with agent at %s", agentAddress)

	// Step 1: Calculate local checksums
	localChecksums := make(map[string]string)
	err = filepath.Walk(config.ConfigPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		hash := sha256.Sum256(data)
		relativePath, _ := filepath.Rel(config.ConfigPath, path)
		localChecksums[relativePath] = hex.EncodeToString(hash[:])
		return nil
	})
	if err != nil {
		return err
	}
	// Step 2: Fetch agent checksums
	res, err := client.client.GetConfigChecksums(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}

	agentChecksums := make(map[string]string)
	for _, entry := range res.Files {
		agentChecksums[entry.Filename] = entry.Checksum
	}

	// Step 3: Compare and push updated/missing files
	for filename, localHash := range localChecksums {
		// Check against whitelist
		whitelisted := false
		for _, allowed := range ConfigFilesWhitelist {
			if allowed == filename {
				whitelisted = true
				break
			}
		}
		if !whitelisted {
			log.Printf("[CONFIG SYNC] Skipping file %s as it is not in the whitelist", filename)
			continue
		}

		// Check if the file exists in agent checksums and compare hashes
		agentHash, exists := agentChecksums[filename]
		if !exists || agentHash != localHash {
			// File is missing or outdated, send it
			fullPath := filepath.Join(config.ConfigPath, filename)
			data, err := os.ReadFile(fullPath)
			if err != nil {
				log.Printf("[CONFIG SYNC] Failed to read file %s: %v", fullPath, err)
				continue
			}

			_, err = client.client.SendConfigFile(ctx, &proto.FileContent{
				Filename: filename,
				Content:  data,
			})
			if err != nil {
				log.Printf("[CONFIG SYNC] Failed to send file %s: %v", filename, err)
			} else {
				log.Printf("[CONFIG SYNC] Synced file: %s", filename)
			}
		}
	}

	log.Println("[CONFIG SYNC] Config sync completed.")
	return nil
}
