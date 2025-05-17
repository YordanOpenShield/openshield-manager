package managergrpc

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"time"

	"openshield-manager/proto"

	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/emptypb"
)

const scriptsDir = "scripts"

func SyncScripts(agentAddress string) error {
	client, err := NewAgentClient(agentAddress)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // use background context
	defer cancel()

	// Step 1: Calculate local checksums
	localChecksums := make(map[string]string)
	err = filepath.Walk(scriptsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		hash := sha256.Sum256(data)
		relativePath, _ := filepath.Rel(scriptsDir, path)
		localChecksums[relativePath] = hex.EncodeToString(hash[:])
		return nil
	})
	if err != nil {
		return err
	}
	// Step 2: Fetch agent checksums
	res, err := client.client.GetScriptChecksums(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}

	agentChecksums := make(map[string]string)
	for _, entry := range res.Scripts {
		agentChecksums[entry.Filename] = entry.Checksum
	}

	// Step 3: Compare and push updated/missing files
	for filename, localHash := range localChecksums {
		agentHash, exists := agentChecksums[filename]
		if !exists || agentHash != localHash {
			// File is missing or outdated, send it
			fullPath := filepath.Join(scriptsDir, filename)
			data, err := os.ReadFile(fullPath)
			if err != nil {
				log.Printf("[SCRIPT SYNC] Failed to read file %s: %v", fullPath, err)
				continue
			}

			_, err = client.client.SendScriptFile(ctx, &proto.FileContent{
				Filename: filename,
				Content:  data,
			})
			if err != nil {
				log.Printf("[SCRIPT SYNC] Failed to send file %s: %v", filename, err)
			} else {
				log.Printf("[SCRIPT SYNC] Synced file: %s", filename)
			}
		}
	}

	// Step 4: Delete files that exist on the agent but not on the manager
	for filename := range agentChecksums {
		if _, exists := localChecksums[filename]; !exists {
			// File exists on agent but not on manager, so delete it
			_, err := client.client.DeleteScriptFile(ctx, &proto.DeleteScriptRequest{
				Filename: filename,
			})
			if err != nil {
				log.Printf("Failed to delete file %s on agent: %v", filename, err)
			} else {
				log.Printf("Deleted outdated file on agent: %s", filename)
			}
		}
	}

	log.Println("[SCRIPT SYNC] Script sync completed.")
	return nil
}
