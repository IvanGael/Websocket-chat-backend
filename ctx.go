package main

import (
	"context"
	"time"
)

func manageRooms(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			roomsMutex.Lock()
			for id, room := range rooms {
				if time.Now().After(room.expiration) {
					room.cancel()
					delete(rooms, id)
				}
			}
			roomsMutex.Unlock()
		}
	}
}
